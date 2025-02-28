package basehandler

import (
	"context"
)

func (r *Repository[TSingle, TPlural, TPluralWithCount]) Create(c context.Context, item *TSingle) (err error) {
	_, err = r.helper.DB.IDB(c).NewInsert().Model(item).Exec(c)
	return err
}

// func (r *Repository[TSingle, TPlural, TPluralWithCount]) GetAll(c context.Context) (items TPluralWithCount, err error) {
// 	items.Items = make(models.Chats[util.WithId], 0)
// 	query := r.helper.DB.IDB(c).NewSelect().
// 		Model(&items.Chats)
//
// 	r.helper.SQL.ExtractFTSP(c).HandleQuery(query)
//
// 	items.Count, err = query.ScanAndCount(c)
// 	return items, err
// }
//
// func (r *Repository[TSingle, TPlural, TPluralWithCount]) Get(c context.Context, id string) (item TSingle, err error) {
// 	// item := TSingle{}
// 	err = r.helper.DB.IDB(c).NewSelect().
// 		Model(&item).
// 		Where("?TableAlias.id = ?", id).Scan(c)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &item, err
// }
//
// func (r *Repository[TSingle, TPlural, TPluralWithCount]) Delete(c context.Context, id string) (err error) {
// 	_, err = r.helper.DB.IDB(c).NewDelete().Model(&models.Chat[util.WithId]{}).Where("id = ?", id).Exec(c)
// 	return err
// }
//
// func (r *Repository[TSingle, TPlural, TPluralWithCount]) Update(c context.Context, item *models.Chat[util.WithId]) (err error) {
// 	_, err = r.helper.DB.IDB(c).NewUpdate().Model(item).Where("id = ?", item.ID).Exec(c)
// 	return err
// }
//
// func (r *Repository[TSingle, TPlural, TPluralWithCount]) UpdateMany(c context.Context, item models.Chats[util.WithId]) (err error) {
// 	_, err = r.helper.DB.IDB(c).NewUpdate().Model(item).Exec(c)
// 	return err
// }
