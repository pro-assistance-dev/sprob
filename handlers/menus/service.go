package menus

import (
	"context"

	"github.com/pro-assistance-dev/sprob/models"
)

func (s *Service) Create(c context.Context, item *models.Menu) error {
	err := s.r.Create(c, item)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Update(c context.Context, item *models.Menu) error {
	err := s.r.Update(c, item)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) GetAll(c context.Context) (models.MenusWithCount, error) {
	return s.r.GetAll(c)
}

func (s *Service) Get(c context.Context, slug string) (*models.Menu, error) {
	item, err := s.r.Get(c, slug)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Delete(c context.Context, id string) error {
	return s.r.Delete(c, id)
}
