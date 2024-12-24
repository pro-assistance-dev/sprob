package chatsmessages

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/chats/models"
)

func (s *Service) GetAll(c context.Context) (models.ChatMessagesWithCount, error) {
	return R.GetAll(c)
}
