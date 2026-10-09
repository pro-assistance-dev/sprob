package inbox

import (
	"context"

	"github.com/pro-assistance-dev/sprob/notify/models"
)

// recipientWhere — адресация «лично или группе»: `"group"` в кавычках — reserved
// word в PostgreSQL, без кавычек синтаксическая ошибка.
const recipientWhere = `user_id = ? OR (user_id = '' AND "group" = ?)`

// GetAll — входящие получателя (свежие сверху) + число непрочитанных.
// Показываем адресованные лично (`user_id`) и групповые (`group`).
func (r *Repository) GetAll(c context.Context, userID string, group string) (models.InboxWithCount, error) {
	items := make(models.InboxItems, 0)
	if err := r.helper.DB.IDB(c).NewSelect().
		Model(&items).
		Where(recipientWhere, userID, group).
		Order("created_at DESC").
		Limit(200).
		Scan(c); err != nil {
		return models.InboxWithCount{}, err
	}
	unread, err := r.helper.DB.IDB(c).NewSelect().
		Model((*models.Inbox)(nil)).
		Where("read_at IS NULL").
		Where(recipientWhere, userID, group).
		Count(c)
	if err != nil {
		return models.InboxWithCount{}, err
	}
	return models.InboxWithCount{Items: items, Unread: unread, Count: len(items)}, nil
}

// MarkRead — пометить одно уведомление прочитанным.
func (r *Repository) MarkRead(c context.Context, id string) error {
	_, err := r.helper.DB.IDB(c).NewUpdate().
		Model((*models.Inbox)(nil)).
		Set("read_at = now()").
		Where("id = ?", id).
		Where("read_at IS NULL").
		Exec(c)
	return err
}

// MarkAllRead — пометить всё прочитанным для получателя.
func (r *Repository) MarkAllRead(c context.Context, userID, group string) error {
	_, err := r.helper.DB.IDB(c).NewUpdate().
		Model((*models.Inbox)(nil)).
		Set("read_at = now()").
		Where("read_at IS NULL").
		Where(recipientWhere, userID, group).
		Exec(c)
	return err
}
