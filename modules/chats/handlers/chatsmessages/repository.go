package chatsmessages

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/chats/models"
)

func (r *Repository) GetAll(c context.Context) (items models.ChatMessagesWithCount, err error) {
	items.ChatMessages = make(models.ChatMessages, 0)
	query := r.helper.DB.IDB(c).NewSelect().
		Model(&items.ChatMessages)
	r.helper.SQL.ExtractFTSP(c).HandleQuery(query)

	items.Count, err = query.ScanAndCount(c)
	return items, err
}
