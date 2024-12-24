package chats

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/chats/models"

	"github.com/uptrace/bun"
)

func (r *Repository) db() *bun.DB {
	return r.helper.DB.DB
}

func (r *Repository) Create(c context.Context, chats *models.Chat) (err error) {
	_, err = r.db().NewInsert().Model(chats).Exec(c)
	return err
}

func (r *Repository) Get(c context.Context, id string) (*models.Chat, error) {
	item := models.Chat{}
	err := r.helper.DB.IDB(c).NewSelect().
		Model(&item).
		Where("?TableAlias.id = ?", id).Scan(c)
	if err != nil {
		return nil, err
	}
	return &item, err
}

func (r *Repository) GetAll(c context.Context) (items models.ChatsWithCount, err error) {
	items.Chats = make(models.Chats, 0)
	query := r.helper.DB.IDB(c).NewSelect().
		Model(&items.Chats)
	r.helper.SQL.ExtractFTSP(c).HandleQuery(query)

	items.Count, err = query.ScanAndCount(c)
	return items, err
}

func (r *Repository) Update(c context.Context, item *models.Chat) (err error) {
	_, err = r.helper.DB.IDB(c).NewUpdate().Model(item).Where("id = ?", item.ID).Exec(c)
	return err
}

func (r *Repository) Delete(c context.Context, id *string) (err error) {
	_, err = r.db().NewDelete().Model(&models.Chat{}).Where("id = ?", id).Exec(c)
	return err
}
