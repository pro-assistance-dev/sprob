package basehandler

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ErrNotSearchable — у модели нет белого списка полей (`SearchColumns`).
var ErrNotSearchable = errors.New("справочник не поддерживает поиск")

const (
	// SearchLimitDefault — сколько опций отдаём, когда клиент не задал limit.
	SearchLimitDefault = 50
	// SearchLimitMax — потолок: защищает от `?limit=100000` (лишняя нагрузка).
	SearchLimitMax = 500
)

// searchLimit приводит запрошенный лимит к допустимому диапазону.
func searchLimit(limit int) int {
	if limit <= 0 {
		return SearchLimitDefault
	}
	if limit > SearchLimitMax {
		return SearchLimitMax
	}
	return limit
}

func (r *Repository[T]) Create(c context.Context, item *T) (err error) {
	_, err = r.helper.DB.IDB(c).NewInsert().Model(item).Exec(c)
	return err
}

type DBItemsWithCount[T any] struct {
	Items []T `json:"items"`
	Count int `json:"count"`
}

type LabelValue struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// Searchable — модель объявляет БЕЛЫЙ СПИСОК полей для серверного поиска опций (К2, Т8).
//
// Зачем белый список: `?query=` нельзя подставлять в SQL напрямую — клиент мог бы
// прислать любое имя колонки. Модель сама говорит, по каким полям её искать.
//
//
//	func (w *Worker) SearchColumns() []string { return []string{"full_name", "email"} }
//
// Пустой список = поиск не поддерживается (сервер вернёт 400 — клиент должен
// знать, что справочник не searchable, а не получать тихо полный список).
type Searchable interface {
	SearchColumns() []string
}

// OptionOrderable — модель задаёт ПОРЯДОК вариантов в /options.
//
// Зачем: по умолчанию список сортируется по подписи (`Order(labelCol)`),
// и это ломается там, где порядок задан данными: у корпусов есть
// `item_order`, и «10-й корпус» встаёт вторым строкой (`Корпус 1, Корпус 10,
// Корпус 2…`). Модель может вернуть колонку порядка (например `item_order`).
type OptionOrderable interface {
	OptionsOrderColumn() string
}

func (r *Repository[T]) Options(c context.Context, labelCol string, valueCol string) ([]*LabelValue, error) {
	return r.optionsQuery(c, labelCol, valueCol, "", 0)
}

// OptionsSearch — то же, что Options, но с серверным поиском по `query` (К2, Т8).
//
// Зачем: справочник сотрудников — 2250 записей; тянуть их целиком в селект
// бессмысленно, если пользователь набирает 3 буквы. Поиск идёт по БЕЛОМУ СПИСКУ
// полей от модели (`SearchColumns`), поэтому `query` не попадает в SQL как SQL.
// Лимит обязателен (0 → SearchLimitDefault).
func (r *Repository[T]) OptionsSearch(
	c context.Context,
	labelCol string,
	valueCol string,
	query string,
	limit int,
) ([]*LabelValue, error) {
	return r.optionsQuery(c, labelCol, valueCol, query, limit)
}

// SearchColumns — белый список полей поиска от модели (пусто — поиск не поддержан).
func (r *Repository[T]) SearchColumns() []string {
	if s, ok := any(*new(T)).(Searchable); ok {
		return s.SearchColumns()
	}
	return nil
}

func (r *Repository[T]) optionsQuery(
	c context.Context,
	labelCol string,
	valueCol string,
	query string,
	limit int,
) ([]*LabelValue, error) {
	items := make([]*LabelValue, 0)

	colExpr := fmt.Sprintf("%s as value, %s as label", valueCol, labelCol)

	q := r.helper.DB.IDB(c).NewSelect().Model((*T)(nil)).ColumnExpr(colExpr)

	// Поиск — только по полям, объявленным моделью (белый список).
	if query != "" {
		cols := r.SearchColumns()
		if len(cols) == 0 {
			return nil, ErrNotSearchable
		}
		conds := make([]string, 0, len(cols))
		args := make([]any, 0, len(cols))
		for _, col := range cols {
			// LOWER(...) LIKE LOWER(?) — регистронезависимо и БЕЗ ILIKE:
			// ILIKE есть только в Postgres, а тесты baseR идут на sqlite
			// (ILIKE там — синтаксическая ошибка). LOWER+LIKE работает в обоих.
			conds = append(conds, fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", col))
			args = append(args, "%"+query+"%")
		}
		q = q.Where("("+strings.Join(conds, " OR ")+")", args...)
	}

	// Порядок: колонка модели (если объявила) → подпись (как было).
	if o, ok := any(*new(T)).(OptionOrderable); ok {
		if col := o.OptionsOrderColumn(); col != "" {
			q = q.Order(col)
		} else {
			q = q.Order(labelCol)
		}
	} else {
		q = q.Order(labelCol)
	}

	q = q.Limit(searchLimit(limit))

	err := q.Scan(c, &items)

	return items, err
}

func (r *Repository[T]) GetAll(c context.Context) (items DBItemsWithCount[T], err error) {
	var i []T

	q := r.helper.DB.IDB(c).NewSelect().Model(&i)

	r.relation(q)
	r.helper.SQL.ExtractFTSP(c).HandleQuery(q)

	items.Count, err = q.ScanAndCount(c)
	items.Items = i

	return items, err
}

func (r *Repository[T]) Get(c context.Context, id string) (item T, err error) {
	q := r.helper.DB.IDB(c).NewSelect().
		Model(&item)

	r.relation(q)

	err = q.Where("?TableAlias.id = ?", id).Scan(c)
	if err != nil {
		return item, err
	}
	return item, err
}

func (r *Repository[T]) Delete(c context.Context, id string) (err error) {
	_, err = r.helper.DB.IDB(c).NewDelete().Model((*T)(nil)).Where("id = ?", id).Exec(c)
	return err
}

func (r *Repository[T]) Update(c context.Context, item *T) (err error) {
	_, err = r.helper.DB.IDB(c).NewUpdate().Model(item).WherePK().Exec(c)
	return err
}

// func (r *Repository[TSingle, TPlural, TPluralWithCount]) UpdateMany(c context.Context, item models.Chats[util.WithId]) (err error) {
// 	_, err = r.helper.DB.IDB(c).NewUpdate().Model(item).Exec(c)
// 	return err
// }
