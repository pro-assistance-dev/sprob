package chatsusers

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/chats/models"
)

func (s *Service) GetAll(c context.Context) (models.ChatsUsersWithCount, error) {
	return R.GetAll(c)
}
