package chatsusers

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/chats/models"
)

func (r *Repository) GetAll(c context.Context) (items models.ChatsUsersWithCount, err error) {
	items.ChatsUsers = make(models.ChatsUsers, 0)
	query := r.helper.DB.IDB(c).NewSelect().
		Model(&items.ChatsUsers)
	r.helper.SQL.ExtractFTSP(c).HandleQuery(query)

	items.Count, err = query.ScanAndCount(c)
	return items, err
}
