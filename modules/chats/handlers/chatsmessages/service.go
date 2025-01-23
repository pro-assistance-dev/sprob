package chatsmessages

import (
	"context"

	"github.com/pro-assistance-dev/sprob/helpers/util"
	"github.com/pro-assistance-dev/sprob/modules/chats/models"
)

func (s *Service) Create(c context.Context, item *models.ChatMessage[util.WithId]) error {
	err := R.Create(c, item)
	if err != nil {
		return err
	}
	return err
}

func (s *Service) Get(c context.Context, id string) (*models.ChatMessage[util.WithId], error) {
	item, err := R.Get(c, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) GetAll(c context.Context) (models.ChatMessagesWithCount[util.WithId], error) {
	return R.GetAll(c)
}

func (s *Service) Update(c context.Context, item *models.ChatMessage[util.WithId]) error {
	err := R.Update(c, item)
	if err != nil {
		return err
	}
	return err
}

func (s *Service) Delete(c context.Context, id *string) error {
	return R.Delete(c, id)
}
