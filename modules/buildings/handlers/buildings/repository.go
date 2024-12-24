package buildings

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/buildings/models"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func (r *Repository) db() *bun.DB {
	return r.helper.DB.DB
}

func (r *Repository) Create(c context.Context, building *models.Building) (err error) {
	_, err = r.db().NewInsert().Model(building).Exec(c)
	return err
}

func (r *Repository) Get(c context.Context, id string) (*models.Building, error) {
	item := models.Building{}
	err := r.helper.DB.IDB(c).NewSelect().
		Model(&item).
		Relation("Floors").
		Relation("Entrances").
		OrderExpr("NULLIF(regexp_replace(number, '\\D','','g'), '')::numeric").
		Order("name").
		Scan(c)
	return &item, err
}

func (r *Repository) GetAll(c context.Context) (items models.BuildingsWithCount, err error) {
	items.Buildings = make(models.Buildings, 0)
	query := r.helper.DB.IDB(c).NewSelect().
		Model(&items.Buildings)
	r.helper.SQL.ExtractFTSP(c).HandleQuery(query)

	items.Count, err = query.ScanAndCount(c)
	return items, err
}

func (r *Repository) GetByFloorID(c context.Context, floorID string) (item models.Building, err error) {
	err = r.db().NewSelect().Model(&item).
		Relation("Floors").
		Where("exists (select * from floors where floors.id = ? and floors.building_id = building.id)", floorID).Scan(c)
	return item, err
}

func (r *Repository) GetByID(c context.Context, id string) (item models.Building, err error) {
	err = r.db().NewSelect().Model(&item).
		Relation("Floors").
		Relation("Entrances").
		Where("id = ?", id).Scan(c)
	return item, err
}

func (r *Repository) Delete(c context.Context, id *string) (err error) {
	_, err = r.db().NewDelete().Model(&models.Building{}).Where("id = ?", id).Exec(c)
	return err
}

// ! TODO Стас посмотри, плз
func (r *Repository) Update(c context.Context, building *models.Building) (err error) {
	_, err = r.db().NewUpdate().Model(building).Where("id = ?", building.ID).Exec(c)
	if err != nil {
		return err
	}
	floor := new([]models.Floor)
	err = r.db().NewSelect().Model(floor).Where("building_id = ?", building.ID).Scan(c)
	if err != nil {
		return err
	}
	for j := 0; j < len(*floor); j++ {
		found := false
		for i := 0; i < len(building.Floors); i++ {
			if building.Floors[i].ID == (*floor)[j].ID {
				found = true
			}
		}
		if !found {
			_, err = r.db().NewDelete().Model(floor).Where("id = ?", (*floor)[j].ID).Exec(c)
			if err != nil {
				return err
			}
		}
	}
	for _, floors := range building.Floors {
		_, err = r.db().NewUpdate().Model(floors).Where("id = ?", floors.ID).Exec(c)
		if err != nil {
			return err
		}
		if floors.ID.UUID == uuid.Nil {
			_, err = r.db().NewInsert().Model(floors).Exec(c)
			if err != nil {
				return err
			}
		}
	}

	entrance := new([]models.Entrance)
	err = r.db().NewSelect().Model(entrance).Where("building_id = ?", building.ID).Scan(c)
	if err != nil {
		return err
	}
	for j := 0; j < len(*entrance); j++ {
		found := false
		for i := 0; i < len(building.Entrances); i++ {
			if building.Entrances[i].ID == (*entrance)[j].ID {
				found = true
			}
		}
		if !found {
			_, err = r.db().NewDelete().Model(entrance).Where("id = ?", (*entrance)[j].ID).Exec(c)
		}
	}
	for _, entrances := range building.Entrances {
		_, err = r.db().NewUpdate().Model(entrances).Where("id = ?", entrances.ID).Exec(c)
		if entrances.ID.UUID == uuid.Nil {
			_, err = r.db().NewInsert().Model(entrances).Exec(c)
		}
	}

	return err
}
