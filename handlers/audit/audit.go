// Package audit — журнал изменений (корректуры) для сервисов sprob.
//
// ⚠️ ЗАЧЕМ ОБЩИЙ ПАКЕТ (10.11.2026). Журнал появился в hr, и владелец прямо
// сказал: «это функционал, который должен распространиться на все сервисы
// рдкб — и портал тоже; я не должен указывать на одни и те же задачи».
// Поэтому вся механика — здесь, а не в проекте:
//
//   - запись изменений (`Use` — middleware на движок, ДО `baseR.Init`!);
//   - снимок «до» из БД (без него нет ни дельты, ни отката);
//   - человекочитаемая дельта «было → стало» (`BuildDelta`);
//   - два отката (`single` — только это изменение, `cascade` — это и всё после);
//   - чтение журнала с фильтрами (`Handler`).
//
// Проекту остаётся ОДНА строка подключения:
//
//	audit.Use(r, helper)               // в routing.Init, ДО baseR.Init
//	audit.Register(api.Group("/audit-log"), helper) // ручки чтения/отката
//
// ⚠️ `Use` именно на движке (`r`) и ДО регистрации роутов: gin применяет
// `group.Use()` только к роутам, объявленным ПОСЛЕ вызова, а sprob регистрирует
// свои (`humans`, `schemas`, `file-infos`…) ВНУТРИ `baseR.Init`. Мидлвар на
// группе `api` их не видел: правка ФИО отвечала 200, а в журнале было пусто
// (поймано 08.11 на hr).
//
// Таблица `audit_log` создаётся миграцией проекта; её структура одинакова у
// всех (см. `models.AuditLog`).
package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/pro-assistance-dev/sprob/helper"
	accessPack "github.com/pro-assistance-dev/sprob/modules/access"
)

// AuditLog — строка журнала. Таблица ОБЩАЯ для всех сервисов.
type AuditLog struct {
	bun.BaseModel `bun:"audit_log,alias:audit_log"`

	ID        uuid.UUID     `bun:"id,type:uuid,pk" json:"id"`
	CreatedAt time.Time     `bun:"created_at,type:timestamptz" json:"createdAt"`
	UserID    uuid.NullUUID `bun:"user_id,type:uuid" json:"userId"`
	UserName  string        `bun:"user_name" json:"userName"`
	RoleCode  string        `bun:"role_code" json:"roleCode"`
	Entity    string        `bun:"entity" json:"entity"`
	ObjectID  uuid.NullUUID `bun:"object_id,type:uuid" json:"objectId"`
	Code      string        `bun:"code" json:"code"`
	Field     string        `bun:"field" json:"field"`
	OldValue  *string       `bun:"old_value" json:"oldValue"`
	NewValue  *string       `bun:"new_value" json:"newValue"`
	Operation string        `bun:"operation" json:"operation"`
}

// Options — настройки журнала для конкретного сервиса.
//
// ⚠️ Сервисы различаются только этим: какие ручки не писать (`Skip`), какие
// сущности имеют русские подписи (`EntityLabels`) и какие откатываются
// (`RevertTables`). Механику дублировать не нужно.
type Options struct {
	// Skip — сегменты пути, которые НЕ журналируются (служебные: `auth`,
	// `analytics`, `search`, `m2m`…). Обязательный минимум добавляется сам
	// (`audit-log`, чтобы просмотр не писал запись о просмотре).
	Skip map[string]bool
	// EntityLabels — русские подписи сущностей (первый сегмент пути →
	// «Сотрудник»). Нет подписи — показываем сырое имя.
	EntityLabels map[string]string
	// RevertTables — сущности, для которых возможен откат: имя ручки → таблица.
	// ⚠️ Только простые таблицы: сущности со вложенными объектами откатываются
	// ПО ЧАСТЯМ, чтобы не переписывать связи.
	RevertTables map[string]string
}

// Audit — middleware и настройки.
type Audit struct {
	helper  *helper.Helper
	options Options
	// titles — кэш «имя объекта» в пределах одного запроса списка журнала.
	titles map[titleKey]string
}

// NewAudit — журнал с настройками проекта.
func NewAudit(h *helper.Helper, opts ...Options) *Audit {
	a := &Audit{helper: h}
	if len(opts) > 0 {
		a.options = opts[0]
	}
	return a
}

// Use — подключить middleware на ДВИЖОК (не на группу!).
//
// ⚠️ Вызывать ДО `baseR.Init`: см. комментарий пакета.
func Use(r *gin.Engine, h *helper.Helper, opts ...Options) *Audit {
	a := NewAudit(h, opts...)
	r.Use(a.Middleware())
	return a
}

// UserFromToken — identity из токена (общий движок access sprob).
func (a *Audit) UserFromToken(r *http.Request) (userID, userName, role string) {
	uc := accessPack.RolesFromRequest(r.Context(), a.helper, r)
	userID = uc.UserID
	userName = uc.UserName
	if len(uc.Roles) > 0 {
		role = uc.Roles[0]
	}
	return
}

// resolveEntity — первый сегмент пути после /api/.
func resolveEntity(path string) string {
	trimmed := strings.TrimPrefix(path, "/api/")
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" {
		return ""
	}
	return strings.SplitN(trimmed, "/", 2)[0]
}

// readPayload — тело запроса как объект: JSON или multipart-поле `form`.
//
// ⚠️ Раньше читался ТОЛЬКО multipart (`sprob`-формат), а `application/json`
// давал пустой payload: в журнале не было ни `ids` пакета, ни `code`, ни
// снимка изменений — записи выглядели одинаковыми (07.11).
func readPayload(c *gin.Context) map[string]interface{} {
	payload := map[string]interface{}{}

	if strings.HasPrefix(c.ContentType(), "application/json") {
		var raw []byte
		if c.Request.GetBody != nil {
			if r, err := c.Request.GetBody(); err == nil && r != nil {
				raw, _ = io.ReadAll(r)
				_ = r.Close()
			}
		}
		if len(raw) == 0 && c.Request.Body != nil {
			raw, _ = io.ReadAll(c.Request.Body)
		}
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &payload)
			// ⚠️ Тело ДОЛЖНО остаться читаемым для хендлера: возвращаем его из
			// буфера, иначе ручка отвечает 400 «некорректное тело запроса»
			// (07.11: bulk-accept перестал работать, как только аудит начал
			// читать JSON).
			c.Request.Body = io.NopCloser(bytes.NewReader(raw))
			return payload
		}
		return nil
	}

	_ = c.Request.ParseMultipartForm(32 << 20)
	form := c.Request.MultipartForm
	if form == nil {
		return nil
	}
	vals := form.Value["form"]
	if len(vals) == 0 {
		return nil
	}
	_ = json.Unmarshal([]byte(vals[0]), &payload)
	return payload
}

// builtinSkip — служебные ручки, которые не журналируются НИКОГДА.
//
// ⚠️ `audit-log` — иначе открытие страницы журнала писало бы запись о том, что
// журнал открыли.
var builtinSkip = map[string]bool{
	"auth": true, "analytics": true, "search": true, "m2m": true,
	"dashboard": true, "uploads": true, "access-matrix": true, "tree": true,
	"audit-log": true,
}

// skipEntities — служебные ручки, не попадающие в журнал.
func (a *Audit) skipEntities(entity string) bool {
	return builtinSkip[entity] || a.options.Skip[entity]
}

// buildItem собирает строку журнала.
//
// ⚠️ ID генерируем ЯВНО: поле `ID uuid.UUID` с `pk` bun пишет ЗНАЧЕНИЕМ, и без
// этого в БД уходит нулевой UUID — вторая же запись билась о `audit_log_pkey`
// (23505), и журнал терял все строки после первой.
func buildItem(method, entity, objectID, code string, newVal *string, userID, userName, role string) *AuditLog {
	op := "update"
	if method == http.MethodPost {
		op = "create"
	}
	if method == http.MethodDelete {
		op = "delete"
	}

	item := &AuditLog{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		Entity:    entity,
		Operation: op,
		Code:      code,
		NewValue:  newVal,
		UserName:  userName,
		RoleCode:  role,
	}
	if userID != "" {
		if id, err := uuid.Parse(userID); err == nil {
			item.UserID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}
	if objectID != "" {
		if id, err := uuid.Parse(objectID); err == nil {
			item.ObjectID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}
	return item
}

// Middleware — запись строки журнала на create/update/delete.
func (a *Audit) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodDelete {
			c.Next()
			return
		}
		entity := resolveEntity(c.Request.URL.Path)
		if entity == "" || a.skipEntities(entity) || strings.HasSuffix(c.Request.URL.Path, "/ftsp") {
			c.Next()
			return
		}

		payload := readPayload(c)

		// ⚠️ Состояние ДО читаем ДО `c.Next()`: после него строка уже изменена,
		// и «старое» значение взять неоткуда. Без этого в журнале был только
		// `new_value` — «что отправили», а не «было → стало», и откат был
		// невозможен (08.11).
		var before map[string]interface{}
		if method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
			before = a.snapshot(c.Request.Context(), entity, objectIDFrom(c, payload))
		}

		c.Next()
		if c.Writer.Status() >= 400 {
			return
		}

		userID, userName, role := a.UserFromToken(c.Request)
		objectID := c.Param("id")
		if s, ok := payload["id"].(string); ok && s != "" {
			objectID = s
		}
		code := ""
		for _, k := range []string{"code", "fullCode", "number", "name"} {
			if v, ok := payload[k].(string); ok && v != "" {
				code = v
				break
			}
		}
		var newVal *string
		if b, err := json.Marshal(payload); err == nil && len(b) > 0 {
			s := string(b)
			if len(s) > 2000 {
				s = s[:2000] + "..."
			}
			newVal = &s
		}
		item := buildItem(method, entity, objectID, code, newVal, userID, userName, role)
		// «Было» — снимок строки ДО запроса (для update/delete). Из него
		// строится дельта в журнале и работает откат.
		item.OldValue = encodeSnapshot(before)
		// Пакетные операции (`POST …/bulk-accept`, `{ids:[…]}`) без этого
		// выглядели как «Создание» без объекта: по такой записи не понять,
		// ЧТО именно приняли (07.11).
		if ids := payloadIDs(payload); len(ids) > 0 {
			if item.ObjectID.Valid == false && len(ids[0]) == 36 {
				if id, err := uuid.Parse(ids[0]); err == nil {
					item.ObjectID = uuid.NullUUID{UUID: id, Valid: true}
				}
			}
			item.Code = pluralObjects(len(ids))
		}

		if _, err := a.helper.DB.IDB(c.Request.Context()).NewInsert().Model(item).Exec(c.Request.Context()); err != nil {
			// потери записей не должны быть незаметны
			log.Printf("[audit] writeAudit: %v (entity=%s op=%s)", err, entity, item.Operation)
		}
	}
}

// payloadIDs — список id из тела (`ids`) для пакетных ручек.
func payloadIDs(payload map[string]interface{}) []string {
	raw, ok := payload["ids"].([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

// pluralObjects — «пакет: N объектов» для колонки «Код».
func pluralObjects(n int) string {
	switch {
	case n%10 == 1 && n%100 != 11:
		return fmt.Sprintf("пакет: %d объект", n)
	case n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20):
		return fmt.Sprintf("пакет: %d объекта", n)
	default:
		return fmt.Sprintf("пакет: %d объектов", n)
	}
}

// UserFromTokenH — identity из токена через ОБЩИЙ движок access (sprob).
//
// ⚠️ Вынесено из проекта (10.11): раньше жило в `hr/access`, и после переезда
// журнала в sprob модуль `hr/access` исчез — функция нужна и журналу, и auth.
func UserFromTokenH(h *helper.Helper, r *http.Request) (userID, userName, role string) {
	uc := accessPack.RolesFromRequest(r.Context(), h, r)
	userID = uc.UserID
	userName = uc.UserName
	if len(uc.Roles) > 0 {
		role = uc.Roles[0]
	}
	return
}
