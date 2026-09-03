package valuetypes

import (
	"context"

	"github.com/pro-assistance-dev/sprob/models"
)

func (s *Service) GetAll(c context.Context) (items models.ValueTypes, err error) {
	return s.r.GetAll(c)
}

func (s *Service) Get(c context.Context, id string) (item *models.ValueType, err error) {
	return s.r.Get(c, id)
}
