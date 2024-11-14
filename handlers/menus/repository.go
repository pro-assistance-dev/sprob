package menus

import (
	"context"

	"github.com/pro-assistance-dev/sprob/models"
	"github.com/uptrace/bun"
	// _ "github.com/go-pg/pg/v10/orm"
)

func (r *Repository) Create(c context.Context, item *models.Menu) (err error) {
	_, err = r.helper.DB.IDB(c).NewInsert().Model(item).Exec(c)
	return err
}

func (r *Repository) GetAll(c context.Context) (item models.MenusWithCount, err error) {
	item.Menus = make(models.Menus, 0)
	query := r.helper.DB.IDB(c).NewSelect().Model(&item.Menus).
		Relation("Icon").
		Relation("SubMenus", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("sub_menus.hide != true").Order("sub_menus.sub_menu_order")
		}).
		Relation("SubMenus.Icon").
		Order("menus.menu_order").
		Where("?TableAlias.hide != true")
	item.Count, err = query.ScanAndCount(c)
	return item, err
}

func (r *Repository) Get(c context.Context, slug string) (*models.Menu, error) {
	item := models.Menu{}
	err := r.helper.DB.IDB(c).NewSelect().Model(&item).
		Where("?TableAlias.id = ?", slug).
		Scan(c)
	return &item, err
}

func (r *Repository) Delete(c context.Context, id string) (err error) {
	_, err = r.helper.DB.IDB(c).NewDelete().Model(&models.Menu{}).Where("id = ?", id).Exec(c)
	return err
}

func (r *Repository) Update(c context.Context, item *models.Menu) (err error) {
	_, err = r.helper.DB.IDB(c).NewUpdate().Model(item).Where("id = ?", item.ID).Exec(c)
	return err
}
