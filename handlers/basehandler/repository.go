package basehandler

import (
	"context"
	"fmt"
)

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

func (r *Repository[T]) Options(c context.Context, labelCol string, valueCol string) ([]*LabelValue, error) {
	items := make([]*LabelValue, 0)

	colExpr := fmt.Sprintf("%s as value, %s as label", valueCol, labelCol)

	err := r.helper.DB.IDB(c).NewSelect().Model((*T)(nil)).ColumnExpr(colExpr).Scan(c, &items)

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
