package entrances

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/buildings/models"
)

func (s *Service) GetAll(c context.Context) (models.EntrancesWithCount, error) {
	return R.GetAll(c)
}
