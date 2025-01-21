package chatsusers

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/chats/models"
)

func (r *Repository) Create(c context.Context, item *models.ChatUser[any]) (err error) {
	_, err = r.helper.DB.IDB(c).NewInsert().Model(item).Exec(c)
	return err
}

func (r *Repository) GetAll(c context.Context) (items models.ChatsUsersWithCount[any], err error) {
	items.ChatsUsers = make(models.ChatsUsers[any], 0)
	query := r.helper.DB.IDB(c).NewSelect().
		Model(&items.ChatsUsers)

	r.helper.SQL.ExtractFTSP(c).HandleQuery(query)

	items.Count, err = query.ScanAndCount(c)
	return items, err
}

func (r *Repository) Get(c context.Context, id string) (*models.ChatUser[any], error) {
	item := models.ChatUser[any]{}
	err := r.helper.DB.IDB(c).NewSelect().
		Model(&item).
		Where("?TableAlias.id = ?", id).Scan(c)
	if err != nil {
		return nil, err
	}
	return &item, err
}

func (r *Repository) Delete(c context.Context, id *string) (err error) {
	_, err = r.helper.DB.IDB(c).NewDelete().Model(&models.ChatUser[any]{}).Where("id = ?", *id).Exec(c)
	return err
}

func (r *Repository) Update(c context.Context, item *models.ChatUser[any]) (err error) {
	_, err = r.helper.DB.IDB(c).NewUpdate().Model(item).Where("id = ?", item.ID).Exec(c)
	return err
}

func (r *Repository) UpdateMany(c context.Context, item models.ChatsUsers[any]) (err error) {
	_, err = r.helper.DB.IDB(c).NewUpdate().Model(item).Exec(c)
	return err
}
