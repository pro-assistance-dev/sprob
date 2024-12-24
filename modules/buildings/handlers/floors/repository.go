package floors

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/buildings/models"
)

func (r *Repository) GetAll(c context.Context) (items models.FloorsWithCount, err error) {
	items.Floors = make(models.Floors, 0)
	query := r.helper.DB.IDB(c).NewSelect().
		Model(&items.Floors)
	r.helper.SQL.ExtractFTSP(c).HandleQuery(query)

	items.Count, err = query.ScanAndCount(c)
	return items, err
}
