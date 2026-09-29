package sql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance-dev/sprob/helpers/project"
	"github.com/pro-assistance-dev/sprob/helpers/sql/filter"
	"github.com/pro-assistance-dev/sprob/helpers/sql/paginator"
	"github.com/pro-assistance-dev/sprob/helpers/sql/sorter"
	"github.com/pro-assistance-dev/sprob/helpers/sql/tree"
	"github.com/uptrace/bun"
)

type FTSP struct {
	Col   string               `json:"col"`
	Value string               `json:"value"`
	F     filter.FilterModels  `json:"f"`
	T     tree.TreeModel       `json:"t"`
	S     sorter.SortModels    `json:"s"`
	P     *paginator.Paginator `json:"p"`
}

func (i *FTSP) HandleQuery(query *bun.SelectQuery) {
	if i == nil {
		return
	}
	i.distinctOn(query)
	i.P.CreatePagination(query)
	i.F.CreateFilter(query)
	i.S.CreateOrder(query)
	i.T.CreateTree(query)
}

type ftspKey struct{}

type FTSPQuery struct {
	QID  string `json:"qid"`
	FTSP FTSP   `json:"ftsp"`
}

func (i *FTSP) distinctOn(query *bun.SelectQuery) {
	if len(i.S) > 0 {
		t := project.SchemasLib.GetSchema(i.S[0].Model)
		sortCol := t.GetColName(i.S[0].Col)
		query.DistinctOn(fmt.Sprintf("%s.%s, %s.id", t.GetTableName(), sortCol, t.GetTableName()))
	}
}

func (i *SQL) InjectFTSP2(r *http.Request, f *FTSP) {
	*r = *r.WithContext(context.WithValue(r.Context(), ftspKey{}, f))
}

func (i *SQL) InjectFTSP(c *gin.Context) error {
	ftsp := &FTSPQuery{}
	if err := ftsp.FromForm(c); err != nil {
		return err
	}
	r := c.Request

	*r = *r.WithContext(context.WithValue(r.Context(), ftspKey{}, ftsp.FTSP))
	return nil
}

func (i *SQL) ExtractFTSP(ctx context.Context) *FTSP {
	if i, ok := ctx.Value(ftspKey{}).(*FTSP); ok {
		return i
	}
	return nil
}

func (i *FTSPQuery) FromForm(c *gin.Context) error {
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	// ⚠️ Без проверки `form.Value["form"][0]` паниковал на запросе без поля
	// `form` (индекс за границей среза) — клиенту уходил 500 вместо 400.
	values := form.Value["form"]
	if len(values) == 0 || values[0] == "" {
		return errors.New("пустое поле `form` в запросе")
	}
	if err := json.Unmarshal([]byte(values[0]), i); err != nil {
		return err
	}
	// Ф4.3: неизвестная модель/поле — 400 с текстом, а не паника в глубине SQL.
	return i.FTSP.Validate()
}

// ErrUnknownField — клиент сослался на поле/модель, которых нет в схеме.
type ErrUnknownField struct {
	Model string
	Field string
}

func (e ErrUnknownField) Error() string {
	if e.Field == "" {
		return "неизвестная модель: " + e.Model
	}
	return fmt.Sprintf("неизвестное поле %q у модели %q", e.Field, e.Model)
}

// BadRequest — маркер для `helpers/http.StatusForError`: это ошибка ЗАПРОСА
// (клиент назвал несуществующее поле/модель), значит ответ — 400, не 500.
func (e ErrUnknownField) BadRequest() bool { return true }

// Validate проверяет, что все имена моделей и колонок из запроса РАЗРЕШАЮТСЯ
// схемой проекта.
//
// ⚠️ ЗАЧЕМ (Ф4.3). Раньше неизвестное имя доезжало до `SchemasLib.GetSchema`
// и падало там nil-pointer'ом (или в `GetColName` — на `FieldsMap[...] = nil`),
// то есть клиент получал 500 «ошибка на сервере» вместо внятного 400. Это
// затрудняло и отладку клиента, и разбор «фильтр молча не работает».
//
// `col`/`value` (прямой поиск) НЕ проверяем по схеме: они передаются как
// значения, а не как имена колонок (см. `ftsp.FilterQuery`).
func (i *FTSP) Validate() error {
	for _, fm := range i.F {
		if fm == nil {
			continue
		}
		if fm.Model == "" {
			return ErrUnknownField{Field: fm.Col}
		}
		schema := project.SchemasLib.GetSchema(fm.Model)
		if schema == nil {
			return ErrUnknownField{Model: fm.Model}
		}
		// Join-фильтры адресуются к колонке ПРИСОЕДИНЯЕМОЙ модели.
		if fm.JoinTableModel != "" {
			join := project.SchemasLib.GetSchema(fm.JoinTableModel)
			if join == nil {
				return ErrUnknownField{Model: fm.JoinTableModel}
			}
			if fm.Col != "" && join.GetField(fm.Col) == nil {
				return ErrUnknownField{Model: fm.JoinTableModel, Field: fm.Col}
			}
			continue
		}
		if fm.Col != "" && schema.GetField(fm.Col) == nil {
			return ErrUnknownField{Model: fm.Model, Field: fm.Col}
		}
	}
	for _, sm := range i.S {
		if sm == nil {
			continue
		}
		schema := project.SchemasLib.GetSchema(sm.Model)
		if schema == nil {
			return ErrUnknownField{Model: sm.Model}
		}
		if sm.Col != "" && schema.GetField(sm.Col) == nil {
			return ErrUnknownField{Model: sm.Model, Field: sm.Col}
		}
	}
	if i.T.Model != "" && project.SchemasLib.GetSchema(i.T.Model) == nil {
		return ErrUnknownField{Model: i.T.Model}
	}
	if i.P != nil && i.P.CursorMode && i.P.Cursor.Model != "" {
		if project.SchemasLib.GetSchema(i.P.Cursor.Model) == nil {
			return ErrUnknownField{Model: i.P.Cursor.Model}
		}
	}
	return nil
}
