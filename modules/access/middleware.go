package access

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iancoleman/strcase"
	"github.com/pro-assistance-dev/sprob/helper"
)

// Middleware - enforcement матрицы доступа (этап 2) и журнал корректуры (этап 3).
type Middleware struct {
	helper  *helper.Helper
	matrix  *MatrixStore
	enforce string // "true" | "monitor" | "" (пусто — выключен)
}

func NewMiddleware(h *helper.Helper) *Middleware {
	return &Middleware{
		helper:  h,
		matrix:  NewMatrixStore(h),
		enforce: os.Getenv("ACCESS_ENFORCE"),
	}
}

// Matrix возвращает store матрицы (для endpoint'ов).
func (m *Middleware) Matrix() *MatrixStore { return m.matrix }

// Enforce включён ли enforcement (ACCESS_ENFORCE=true).
func (m *Middleware) Enforce() bool { return m.enforce == "true" }

// Monitor включён ли режим мониторинга (ACCESS_ENFORCE=monitor): проверки
// выполняются, но вместо 403 пишется лог WOULD-403 (Б3, TECH_DEBT.md).
func (m *Middleware) Monitor() bool { return m.enforce == "monitor" }

// resolveEntity определяет сущность по пути запроса: /api/{key}...
func resolveEntity(path string) *Entity {
	trimmed := strings.TrimPrefix(path, "/api/")
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" {
		return nil
	}
	seg := strings.SplitN(trimmed, "/", 2)[0]
	// POST /ftsp идёт по тому же ключу
	return GetEntity(seg)
}

// bodyCaptureWriter буферизует ответ, чтобы можно было замаскировать поля до отправки.
type bodyCaptureWriter struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (w *bodyCaptureWriter) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return len(b), nil
}

// AccessControl - middleware разграничения доступа.
//
// Невалидный токен (подпись не прошла проверку, см. jwks.go) → 401 (Б1).
// GET/FTSP: из ответа вырезаются поля без права R (маскирование).
// POST/PUT: поля без права W/M -> 403 (для ярусов 2/3 - проверка на уровне сущности).
// DELETE: без права W на сущность -> 403.
//
// Режимы (переменная окружения ACCESS_ENFORCE):
//   - (пусто, по умолчанию) — выключен: только маскирование чтения при наличии ролей;
//   - monitor — проверки выполняются, вместо 403 пишется лог WOULD-403 (Б3);
//   - true — жёсткий enforcement.
func (m *Middleware) AccessControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Б1 (TECH_DEBT.md): токен присутствует, но его подпись не прошла проверку
		// (JWKS keycloak / TOKEN_SECRET) — отвечаем 401 ДО любых проверок сущности.
		// Запросы без токена остаются анонимными (datasync, uploads и т.п.).
		uc := RolesFromRequest(c.Request.Context(), m.helper, c.Request)
		if uc.TokenInvalid {
			log.Printf("[access][jwt] 401: invalid token for %s %s from %s",
				c.Request.Method, c.Request.URL.Path, c.ClientIP())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		entity := resolveEntity(c.Request.URL.Path)
		if entity == nil {
			c.Next()
			return
		}
		ctx := c.Request.Context()

		isRead := c.Request.Method == http.MethodGet ||
			(c.Request.Method == http.MethodPost && strings.HasSuffix(c.Request.URL.Path, "/ftsp"))

		enforced := m.Enforce()
		monitor := m.Monitor()

		if !enforced && !monitor {
			// Без enforcement маскируем ответы только если у пользователя есть роли.
			if isRead && len(uc.Roles) > 0 && !hasRole(uc.Roles, AdminRole) {
				w := &bodyCaptureWriter{ResponseWriter: c.Writer, buf: &bytes.Buffer{}}
				c.Writer = w
				c.Next()
				m.maskResponse(c, w, entity, uc)
				return
			}
			c.Next()
			return
		}

		// deny: в режиме enforcement прерывает запрос 403; в monitor — только логирует
		// WOULD-403 и возвращает false (запрос продолжается как раньше).
		deny := func(code int, msg string) bool {
			if monitor {
				log.Printf("[access][enforce] WOULD-%d: %s %s roles=%v — %s",
					code, c.Request.Method, c.Request.URL.Path, uc.Roles, msg)
				return false
			}
			c.AbortWithStatusJSON(code, gin.H{"error": msg})
			return true
		}

		// ===== Enforcement / Monitor =====
		switch {
		case isRead:
			if len(uc.Roles) == 0 {
				if !deny(http.StatusForbidden, "нет ролей FM") {
					c.Next()
				}
				return
			}
			w := &bodyCaptureWriter{ResponseWriter: c.Writer, buf: &bytes.Buffer{}}
			c.Writer = w
			c.Next()
			m.maskResponse(c, w, entity, uc)
		case c.Request.Method == http.MethodPost, c.Request.Method == http.MethodPut:
			payload := readFormJSON(c)
			if entity.EntityLevel {
				if !m.matrix.CanWriteEntity(ctx, entity.Key, uc.Roles) {
					if !deny(http.StatusForbidden, "нет права записи для сущности "+entity.Key) {
						c.Next()
					}
					return
				}
			} else {
				// PUT: клиент (sprof Update) шлёт ВЕСЬ объект, поэтому проверяем только
				// поля, которые реально меняются относительно текущей строки БД.
				// Иначе любой не-админ получил бы 403 на сохранение (у него нет W
				// на большинстве полей, но он их не меняет) — Б3.
				var oldRow map[string]interface{}
				if c.Request.Method == http.MethodPut {
					id := c.Param("id")
					if s, ok := payload["id"].(string); ok && s != "" {
						id = s
					}
					oldRow = fetchOldRowFn(m, ctx, entity, id)
				}
				var denied []string
				for field, newVal := range payload {
					if field == "id" {
						continue // id не проверяется при записи
					}
					if _, known := entity.Fields[field]; !known {
						continue
					}
					if oldRow != nil && !fieldChanged(oldRow, field, newVal) {
						continue // значение не меняется — не блокируем
					}
					if !m.matrix.CanWrite(ctx, entity.Key, field, uc.Roles) {
						denied = append(denied, field)
					}
				}
				if len(denied) > 0 {
					if !deny(http.StatusForbidden, "нет права записи на поля: "+strings.Join(denied, ",")) {
						c.Next()
					}
					return
				}
			}
			c.Next()
		case c.Request.Method == http.MethodDelete:
			if !m.matrix.CanDelete(ctx, entity.Key, uc.Roles) {
				if !deny(http.StatusForbidden, "нет права на удаление "+entity.Key) {
					c.Next()
				}
				return
			}
			c.Next()
		default:
			c.Next()
		}
	}
}

// maskResponse маскирует JSON-ответ согласно матрице доступа.
// ВАЖНО: пишем в w.ResponseWriter (реальный writer), а не в c.Writer —
// c.Writer в этот момент уже подменён на bodyCaptureWriter, и запись в него
// повторно буферизует ответ (клиент получит пустое тело).
func (m *Middleware) maskResponse(c *gin.Context, w *bodyCaptureWriter, entity *Entity, uc UserCtx) {
	if c.Writer.Status() != http.StatusOK {
		_, _ = w.ResponseWriter.Write(w.buf.Bytes())
		return
	}
	ct := c.Writer.Header().Get("Content-Type")
	if !strings.Contains(ct, "json") {
		_, _ = w.ResponseWriter.Write(w.buf.Bytes())
		return
	}

	var v interface{}
	if err := json.Unmarshal(w.buf.Bytes(), &v); err != nil {
		_, _ = w.ResponseWriter.Write(w.buf.Bytes())
		return
	}

	// FTSP/список: {"data": [...]} или {"items": [...], "count": ...}
	if obj, ok := v.(map[string]interface{}); ok {
		if data, ok := obj["data"]; ok {
			obj["data"] = maskValue(c, data, entity, m.matrix, uc)
			v = obj
		} else if items, ok := obj["items"]; ok {
			obj["items"] = maskValue(c, items, entity, m.matrix, uc)
			v = obj
		} else {
			v = maskValue(c, v, entity, m.matrix, uc)
		}
	} else {
		v = maskValue(c, v, entity, m.matrix, uc)
	}

	out, err := json.Marshal(v)
	if err != nil {
		_, _ = w.ResponseWriter.Write(w.buf.Bytes())
		return
	}
	_, _ = w.ResponseWriter.Write(out)
}

// maskValue рекурсивно вырезает поля без права чтения.
func maskValue(c *gin.Context, v interface{}, entity *Entity, matrix *MatrixStore, uc UserCtx) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, val := range t {
			if relKey, ok := entity.Relations[k]; ok {
				if relEnt := GetEntity(relKey); relEnt != nil {
					t[k] = maskValue(c, val, relEnt, matrix, uc)
				}
				continue
			}
			if _, known := entity.Fields[k]; known {
				if !matrix.CanRead(c.Request.Context(), entity.Key, k, uc.Roles) {
					delete(t, k)
				}
			}
		}
		return t
	case []interface{}:
		for i := range t {
			t[i] = maskValue(c, t[i], entity, matrix, uc)
		}
		return t
	default:
		return v
	}
}

// readFormJSON достаёт JSON объекта из multipart-поля `form` (формат sprof).
func readFormJSON(c *gin.Context) map[string]interface{} {
	_ = c.Request.ParseMultipartForm(32 << 20)
	form := c.Request.MultipartForm
	if form == nil {
		return nil
	}
	vals := form.Value["form"]
	if len(vals) == 0 {
		return nil
	}
	payload := map[string]interface{}{}
	_ = json.Unmarshal([]byte(vals[0]), &payload)
	return payload
}

// fetchOldRowFn - точка расширения для тестов (Б3): в проде читает текущую
// строку из БД (fetchOldRow), в тестах подменяется фейковой строкой.
var fetchOldRowFn = func(m *Middleware, ctx context.Context, entity *Entity, id string) map[string]interface{} {
	return m.fetchOldRow(ctx, entity, id)
}

// fieldChanged - изменилось ли значение поля относительно текущей строки БД.
// json-имя поля -> snake_case колонка (для связей — суффикс _id). Временные
// значения сравниваются по инстансу (клиент шлёт ISO, БД — timestamptz).
func fieldChanged(oldRow map[string]interface{}, field string, newVal interface{}) bool {
	col := strcase.ToSnake(field)
	oldVal, ok := oldRow[col]
	if !ok {
		oldVal, ok = oldRow[col+"_id"]
	}
	if !ok {
		return true // поля нет в текущей строке — считаем изменённым
	}
	oldStr, newStr := stringify(oldVal), stringify(newVal)
	if oldStr == newStr {
		return false
	}
	ot, oerr := parseTimeLoose(oldStr)
	nt, nerr := parseTimeLoose(newStr)
	if oerr && nerr && ot.Equal(nt) {
		return false
	}
	return true
}

// parseTimeLoose парсит время в форматах БД (timestamptz) и клиента (ISO).
func parseTimeLoose(s string) (time.Time, bool) {
	layouts := []string{
		time.RFC3339Nano, time.RFC3339,
		"2006-01-02 15:04:05.999999-07:00", "2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05", "2006-01-02",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// Audit - middleware журнала корректуры (этап 3).
// После успешного create/update/delete пишет строки audit_log (поле за полем).
// ВАЖНО: старое состояние строки читаем ДО выполнения handler'а (c.Next),
// иначе diff всегда будет пустым (строка уже обновлена).
func (m *Middleware) Audit() gin.HandlerFunc {
	return func(c *gin.Context) {
		entity := resolveEntity(c.Request.URL.Path)
		if entity == nil {
			c.Next()
			return
		}
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodDelete {
			c.Next()
			return
		}
		if strings.HasSuffix(c.Request.URL.Path, "/ftsp") {
			c.Next()
			return
		}

		payload := readFormJSON(c)
		uc := RolesFromRequest(c.Request.Context(), m.helper, c.Request)
		ctx := c.Request.Context()

		// Старое состояние строки — ДО правки
		var oldRow map[string]interface{}
		if method == http.MethodPut {
			id := c.Param("id")
			if s, ok := payload["id"].(string); ok && s != "" {
				id = s
			}
			oldRow = m.fetchOldRow(ctx, entity, id)
		}

		// Для POST перехватываем ответ, чтобы получить id созданного объекта:
		// клиент может не слать id в payload (сервер генерирует его сам), и без
		// этого object_id в журнале остаётся пустым.
		var capture *bodyCaptureWriter
		if method == http.MethodPost {
			capture = &bodyCaptureWriter{ResponseWriter: c.Writer, buf: &bytes.Buffer{}}
			c.Writer = capture
		}

		c.Next()
		if c.Writer.Status() >= 400 {
			if capture != nil {
				_, _ = capture.ResponseWriter.Write(capture.buf.Bytes())
			}
			return
		}

		createdID := ""
		if capture != nil {
			var resp map[string]interface{}
			if err := json.Unmarshal(capture.buf.Bytes(), &resp); err == nil {
				if s, ok := resp["id"].(string); ok {
					createdID = s
				}
			}
			_, _ = capture.ResponseWriter.Write(capture.buf.Bytes())
		}

		id := c.Param("id")
		objectID := payload["id"]
		if s, ok := objectID.(string); ok {
			id = s
		}
		if createdID != "" {
			id = createdID
		}
		// Если id так и не получили (POST без id в payload, sprob Create без
		// RETURNING → в ответе id=null) — ищем созданную строку по полям payload.
		if id == "" && method == http.MethodPost {
			id = m.findCreatedID(ctx, entity, payload)
		}

		switch method {
		case http.MethodPost:
			m.auditCreate(ctx, entity, id, payload, uc)
		case http.MethodPut:
			m.auditUpdate(ctx, entity, id, oldRow, payload, uc)
		case http.MethodDelete:
			m.auditDelete(ctx, entity, id, payload, uc)
		}
	}
}
