package valuetypes

import (
	"context"

	"github.com/pro-assistance-dev/sprob/models"
)

func (s *Service) GetAll(c context.Context) (items models.ValueTypes, err error) {
	return R.GetAll(c)
}

func (s *Service) Get(c context.Context, id string) (item *models.ValueType, err error) {
	return R.Get(c, id)
}
