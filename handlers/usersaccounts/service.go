package usersaccounts

import (
	"context"

	"github.com/pro-assistance-dev/sprob/models"
)

func (s *Service) Create(c context.Context, item *models.UserAccount) error {
	err := item.HashPassword()
	if err != nil {
		return err
	}
	return R.Create(c, item)
}

func (s *Service) GetAll(c context.Context) (models.EmailsWithCount, error) {
	return R.GetAll(c)
}

func (s *Service) Get(c context.Context, id string) (*models.UserAccount, error) {
	return R.Get(c, id)
}

func (s *Service) Update(c context.Context, item *models.UserAccount) error {
	err := item.HashPassword()
	if err != nil {
		return err
	}
	return R.Update(c, item)
}

func (s *Service) Delete(c context.Context, id string) error {
	return R.Delete(c, id)
}
