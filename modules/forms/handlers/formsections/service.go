package formsections

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/forms/models"
)

func (s *Service) Create(c context.Context, item *models.FormSection) error {
	err := s.r.Create(c, item)
	if err != nil {
		return err
	}
	return err
}

func (s *Service) GetAll(c context.Context) (models.FormSectionsWithCount, error) {
	return s.r.GetAll(c)
}

func (s *Service) Get(c context.Context, id string) (*models.FormSection, error) {
	item, err := s.r.Get(c, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Update(c context.Context, item *models.FormSection) error {
	err := s.r.Update(c, item)
	if err != nil {
		return err
	}
	return err
}

func (s *Service) UpdateMany(c context.Context, item models.FormSections) error {
	err := s.r.UpdateMany(c, item)
	if err != nil {
		return err
	}
	return err
}

func (s *Service) Delete(c context.Context, id *string) error {
	return s.r.Delete(c, id)
}
