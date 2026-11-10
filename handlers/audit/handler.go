package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Item — строка журнала для клиента.
type Item struct {
	ID        string `json:"id" bun:"id"`
	CreatedAt string `json:"createdAt" bun:"created_at"`
	UserName  string `json:"userName" bun:"user_name"`
	RoleCode  string `json:"roleCode" bun:"role_code"`
	Entity    string `json:"entity" bun:"entity"`
	EntityRus string `json:"entityRus" bun:"-"`
	ObjectID  string `json:"objectId" bun:"object_id"`
	// ObjectTitle — ЧЕЛОВЕКОЧИТАЕМОЕ имя объекта («Выхристюк Юлия», «Отдел»).
	// ⚠️ id обычному пользователю ничего не говорит (10.11): в журнале нужно
	// имя, а id оставить только для ссылки.
	ObjectTitle string `json:"objectTitle" bun:"-"`
	Code        string `json:"code" bun:"code"`
	Field       string `json:"field" bun:"field"`
	FieldRus    string `json:"fieldRus" bun:"-"`
	OldValue    string `json:"oldValue" bun:"old_value"`
	NewValue    string `json:"newValue" bun:"new_value"`
	Operation   string `json:"operation" bun:"operation"`

	// Delta — ЧЕЛОВЕКОЧИТАЕМАЯ разница «было → стало» по полям.
	// Пусто у create (сравнивать не с чем) и у старых записей без `old_value`.
	Delta []DeltaItem `json:"delta" bun:"-"`
	// CanRevert — можно ли откатить к состоянию из этой записи.
	CanRevert bool `json:"canRevert" bun:"-"`
	// IsRevert — это запись О САМОМ ОТКАТЕ (её пишет ручка revert).
	IsRevert bool `json:"isRevert" bun:"-"`
}

// Page — страница журнала.
type Page struct {
	Count int     `json:"count"`
	Items []*Item `json:"items"`
}

// Query — параметры выборки.
type Query struct {
	Search    string
	Entity    string
	Operation string
	ObjectID  string
	// UserName — фильтр по конкретному пользователю (10.11: «хочу посмотреть
	// изменения конкретного пользователя»).
	UserName string
	Offset   int
	Limit    int
}

// Handler — ручки чтения и отк��та.
type Handler struct {
	audit *Audit
}

var handler *Handler

// Register — зарегистрировать ручки журнала.
//
//	GET  <prefix>?search=&entity=&operation=&objectId=&offset=&limit=
//	GET  <prefix>/operations
//	POST <prefix>/revert   {"id":"…","mode":"single|cascade"}
func Register(r *gin.RouterGroup, a *Audit) {
	handler = &Handler{audit: a}
	r.GET("", handler.Get)
	r.GET("/operations", handler.Operations)
	r.GET("/users", handler.Users)
	r.POST("/revert", handler.Revert)
}

// Get — чтение журнала.
func (h *Handler) Get(c *gin.Context) {
	q := Query{
		Search:    strings.TrimSpace(c.Query("search")),
		Entity:    strings.TrimSpace(c.Query("entity")),
		Operation: strings.TrimSpace(c.Query("operation")),
		ObjectID:  strings.TrimSpace(c.Query("objectId")),
		UserName:  strings.TrimSpace(c.Query("userName")),
	}
	q.Offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))
	q.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "25"))
	if q.Limit <= 0 || q.Limit > 200 {
		q.Limit = 25
	}
	if q.Offset < 0 {
		q.Offset = 0
	}

	ctx := c.Request.Context()
	a := h.audit
	// ⚠️ Кэш имён объектов живёт в пределах ОДНОГО ответа: в списке одни и те
	// же объекты повторяются десятки раз, а данные между запросами меняются.
	a.resetTitles()

	// Общие условия — в одном месте: и `count`, и страница фильтруются одинаково.
	applyWhere := func(sq *bun.SelectQuery) *bun.SelectQuery {
		if q.Entity != "" {
			sq = sq.Where("audit_log.entity = ?", q.Entity)
		}
		if q.Operation != "" {
			sq = sq.Where("audit_log.operation = ?", q.Operation)
		}
		if q.ObjectID != "" {
			sq = sq.Where("audit_log.object_id = ?", q.ObjectID)
		}
		if q.UserName != "" {
			sq = sq.Where("audit_log.user_name = ?", q.UserName)
		}
		if q.Search != "" {
			like := "%" + strings.ToLower(q.Search) + "%"
			// ⚠️ Ищем и по ПОДПИСЯМ сущностей — человек ищет «Документ», а в
			// базе лежит `worker-documents`.
			entities := matchKeys(a.options.EntityLabels, q.Search)
			sq = sq.WhereGroup("(", func(w *bun.SelectQuery) *bun.SelectQuery {
				w = w.Where("lower(audit_log.user_name) LIKE ?", like).
					WhereOr("lower(audit_log.entity) LIKE ?", like).
					WhereOr("lower(coalesce(audit_log.field, '')) LIKE ?", like).
					WhereOr("lower(coalesce(audit_log.old_value, '')) LIKE ?", like).
					WhereOr("lower(coalesce(audit_log.new_value, '')) LIKE ?", like).
					WhereOr("lower(coalesce(audit_log.code, '')) LIKE ?", like)
				if len(entities) > 0 {
					w = w.WhereOr("audit_log.entity IN (?)", bun.In(entities))
				}
				return w
			})
		}
		return sq
	}

	// ⚠️ `count` — ОТДЕЛЬНОЙ выборкой без `ColumnExpr`-колонок: рядом с ними
	// `count(*)` требует GROUP BY по всем (SQLSTATE 42803).
	var count int
	if err := applyWhere(a.helper.DB.IDB(ctx).NewSelect().TableExpr("audit_log AS audit_log")).
		ColumnExpr("count(*)").
		Scan(ctx, &count); err != nil {
		a.helper.HTTP.HandleError(c, err)
		return
	}

	rows := make([]*Item, 0)
	// ⚠️ Колонки выбираем ЯВНО: `Model(...)` + `Scan` в `Item` падает на
	// `user_id` («does not have column»). Дату отдаём строкой.
	sq := applyWhere(a.helper.DB.IDB(ctx).NewSelect().TableExpr("audit_log AS audit_log")).
		ColumnExpr("audit_log.id::text AS id").
		ColumnExpr("to_char(audit_log.created_at, 'YYYY-MM-DD HH24:MI:SS') AS created_at").
		ColumnExpr("coalesce(audit_log.user_name, '') AS user_name").
		ColumnExpr("coalesce(audit_log.role_code, '') AS role_code").
		ColumnExpr("audit_log.entity AS entity").
		ColumnExpr("coalesce(audit_log.object_id::text, '') AS object_id").
		ColumnExpr("coalesce(audit_log.code, '') AS code").
		ColumnExpr("coalesce(audit_log.field, '') AS field").
		ColumnExpr("coalesce(audit_log.old_value, '') AS old_value").
		ColumnExpr("coalesce(audit_log.new_value, '') AS new_value").
		ColumnExpr("audit_log.operation AS operation").
		Order("audit_log.created_at DESC").
		Limit(q.Limit).
		Offset(q.Offset)
	if err := sq.Scan(ctx, &rows); err != nil {
		a.helper.HTTP.HandleError(c, err)
		return
	}

	for _, it := range rows {
		if label, ok := a.options.EntityLabels[it.Entity]; ok {
			it.EntityRus = label
		} else {
			it.EntityRus = it.Entity
		}
		it.FieldRus = fieldLabel(it.Field)
		it.ObjectTitle = a.objectTitle(ctx, it)
		it.IsRevert = strings.Contains(it.NewValue, `"revertedFrom"`)
		if it.IsRevert {
			// У отката дельты нет: это служебная запись, а не правка полей.
			it.Delta = []DeltaItem{}
		} else {
			it.Delta = BuildDelta(it.OldValue, it.NewValue)
		}
		_, tableOK := a.options.RevertTables[it.Entity]
		it.CanRevert = tableOK && strings.TrimSpace(it.OldValue) != "" && it.ObjectID != ""
	}
	c.JSON(http.StatusOK, Page{Count: count, Items: rows})
}

// Operations — список операций для фильтра (с подписями).
func (h *Handler) Operations(c *gin.Context) {
	type option struct {
		Value string `json:"value"`
		Label string `json:"label"`
	}
	// Порядок фиксирован: иначе фильтр «прыгает» между запросами.
	out := []option{
		{Value: "create", Label: "Создание"},
		{Value: "update", Label: "Изменение"},
		{Value: "delete", Label: "Удаление"},
	}
	c.JSON(http.StatusOK, out)
}

// Users — список пользователей, которые есть в журнале (для фильтра).
//
//	GET <prefix>/users
//
// ⚠️ Из САМОГО журнала, а не из справочника пользователей: в фильтре нужны
// только те, кто реально что-то менял (иначе — сотни пустых пунктов).
func (h *Handler) Users(c *gin.Context) {
	type userRow struct {
		UserName string `bun:"user_name" json:"userName"`
		Count    int    `bun:"cnt" json:"count"`
	}
	rows := make([]userRow, 0)
	if err := h.audit.helper.DB.IDB(c.Request.Context()).NewRaw(`
		SELECT coalesce(user_name, '') AS user_name, count(*) AS cnt
		FROM audit_log
		WHERE coalesce(user_name, '') <> ''
		GROUP BY 1
		ORDER BY 2 DESC`).Scan(c.Request.Context(), &rows); err != nil {
		h.audit.helper.HTTP.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}

// matchKeys — ключи словаря, чья ПОДПИСЬ или КЛЮЧ содержит запрос.
func matchKeys(dict map[string]string, search string) []string {
	needle := strings.ToLower(search)
	out := make([]string, 0)
	for k, v := range dict {
		if strings.Contains(strings.ToLower(v), needle) || strings.Contains(strings.ToLower(k), needle) {
			out = append(out, k)
		}
	}
	return out
}

// auditRowData — одна запись журнала (для отката).
type auditRowData struct {
	Entity    string  `bun:"entity"`
	ObjectID  string  `bun:"object_id"`
	OldValue  *string `bun:"old_value"`
	NewValue  *string `bun:"new_value"`
	Operation string  `bun:"operation"`
}

func (h *Handler) auditRow(ctx context.Context, id uuid.UUID) (auditRowData, error) {
	var row auditRowData
	err := h.audit.helper.DB.IDB(ctx).NewSelect().
		TableExpr("audit_log").
		ColumnExpr("entity").
		ColumnExpr("coalesce(object_id::text, '') AS object_id").
		ColumnExpr("old_value").
		ColumnExpr("new_value").
		ColumnExpr("operation").
		Where("id = ?", id).
		Scan(ctx, &row)
	return row, err
}

// tableColumns — множество колонок таблицы.
func (a *Audit) tableColumns(ctx context.Context, table string) (map[string]bool, error) {
	rows := make([]struct {
		ColumnName string `bun:"column_name"`
	}, 0)
	if err := a.helper.DB.IDB(ctx).NewRaw(`
		SELECT column_name FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = ?`, table).Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(rows))
	for _, r := range rows {
		out[r.ColumnName] = true
	}
	return out, nil
}

// actorName — кто выполняет действие (из токена).
func actorName(c *gin.Context) string {
	if v, ok := c.Get("userName"); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return ""
}

// jsonUnmarshal — небольшой помощник (чтобы не тянуть encoding/json в
// несколько файлов ради одной строки).
func jsonUnmarshal(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}
