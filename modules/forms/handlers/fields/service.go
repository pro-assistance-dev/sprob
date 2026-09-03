package fields

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/forms/models"

	"github.com/google/uuid"
)

func (s *Service) Create(c context.Context, item *models.Field) error {
	return s.r.Create(c, item)
}

func (s *Service) GetAll(c context.Context) (models.FieldsWithCount, error) {
	items, err := s.r.GetAll(c)
	if err != nil {
		return items, err
	}
	return items, nil
}

func (s *Service) Get(c context.Context, id string) (*models.Field, error) {
	item, err := s.r.Get(c, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Update(c context.Context, item *models.Field) error {
	return s.r.Update(c, item)
}

func (s *Service) GetAnthropometryFields(c context.Context) (models.Fields, error) {
	return s.r.GetAnthropometryFields(c)
}

func (s *Service) Delete(c context.Context, id string) error {
	return s.r.Delete(c, id)
}

func (s *Service) UpsertMany(c context.Context, items models.Fields) error {
	if len(items) == 0 {
		return nil
	}

	err := s.r.upsertMany(c, items)
	if err != nil {
		return err
	}
	// registerPropertyRadioService := FieldFillvariants.CreateService(s.helper)
	// err = registerPropertyRadioService.UpsertMany(items.GetRegisterPropertyRadios())
	// if err != nil {
	// 	return err
	// }
	// err = registerPropertyRadioService.DeleteMany(items.GetRegisterPropertyRadioForDelete())
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (s *Service) DeleteMany(c context.Context, idPool []uuid.UUID) error {
	if len(idPool) == 0 {
		return nil
	}
	return s.r.deleteMany(c, idPool)
}
