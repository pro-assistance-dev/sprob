package buildings

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/buildings/models"
)

func (s *Service) Create(c context.Context, item *models.Building) error {
	err := R.Create(c, item)
	if err != nil {
		return err
	}
	return err
}

func (s *Service) Get(c context.Context, id string) (*models.Building, error) {
	item, err := R.Get(c, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) GetAll(c context.Context) (models.BuildingsWithCount, error) {
	return R.GetAll(c)
}

func (s *Service) Update(c context.Context, item *models.Building) error {
	err := R.Update(c, item)
	if err != nil {
		return err
	}
	return err
}

func (s *Service) Delete(c context.Context, id *string) error {
	return R.Delete(c, id)
}
