package inns

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/inns/models"
)

func (s *Service) Create(c context.Context, item *models.Inn) error {
	err := R.Create(c, item)
	if err != nil {
		return err
	}
	return err
}

func (s *Service) GetAll(c context.Context) (models.InnsWithCount, error) {
	return R.GetAll(c)
}

func (s *Service) Get(c context.Context, id string) (*models.Inn, error) {
	item, err := R.Get(c, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Update(c context.Context, item *models.Inn) error {
	err := R.Update(c, item)
	if err != nil {
		return err
	}
	return err
}

func (s *Service) Delete(c context.Context, id *string) error {
	return R.Delete(c, id)
}

func (s *Service) UpdateMany(c context.Context, item models.Inns) error {
	err := R.UpdateMany(c, item)
	if err != nil {
		return err
	}
	return err
}
