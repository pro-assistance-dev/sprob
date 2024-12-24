package entrances

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/buildings/models"
)

func (r *Repository) GetAll(c context.Context) (items models.EntrancesWithCount, err error) {
	items.Entrances = make(models.Entrances, 0)
	query := r.helper.DB.IDB(c).NewSelect().
		Model(&items.Entrances)
	r.helper.SQL.ExtractFTSP(c).HandleQuery(query)

	items.Count, err = query.ScanAndCount(c)
	return items, err
}
