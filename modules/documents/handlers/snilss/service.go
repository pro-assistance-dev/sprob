package snilss

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/documents/models"
)

func (s *Service) Create(c context.Context, item *models.Snils) error {
	err := R.Create(c, item)
	if err != nil {
		return err
	}
	return err
}

func (s *Service) GetAll(c context.Context) (models.SnilssWithCount, error) {
	return R.GetAll(c)
}

func (s *Service) Get(c context.Context, id string) (*models.Snils, error) {
	item, err := R.Get(c, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Update(c context.Context, item *models.Snils) error {
	err := R.Update(c, item)
	if err != nil {
		return err
	}
	return err
}

func (s *Service) Delete(c context.Context, id *string) error {
	return R.Delete(c, id)
}

func (s *Service) UpdateMany(c context.Context, item models.Snilss) error {
	err := R.UpdateMany(c, item)
	if err != nil {
		return err
	}
	return err
}