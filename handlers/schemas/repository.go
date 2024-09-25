package schemas

import (
	"context"

	"github.com/pro-assistance/pro-assister/helpers/project"
)

func (r *Repository) Create(c context.Context, item *project.Schema) (err error) {
	_, err = r.helper.DB.IDB(c).NewInsert().Model(item).Exec(c)
	return err
}

func (r *Repository) GetAll(c context.Context) (items project.SchemasWithCount, err error) {
	items.Schemas = make(project.Schemas, 0)
	q := r.helper.DB.IDB(c).NewSelect().Model(&items.Schemas).
		Relation("Fields")
	r.helper.SQL.ExtractFTSP(c).HandleQuery(q)
	items.Count, err = q.ScanAndCount(c)
	return items, err
}

func (r *Repository) Get(c context.Context, id string) (*project.Schema, error) {
	item := project.Schema{}
	err := r.helper.DB.IDB(c).NewSelect().Model(&item).
		Where("?TableAlias.id = ?", id).
		Scan(c)
	return &item, err
}

func (r *Repository) Delete(c context.Context, id string) (err error) {
	_, err = r.helper.DB.IDB(c).NewDelete().Model(&project.Schema{}).Where("id = ?", id).Exec(c)
	return err
}

func (r *Repository) Update(c context.Context, item *project.Schema) (err error) {
	_, err = r.helper.DB.IDB(c).NewUpdate().Model(item).Where("id = ?", item.ID).Exec(c)
	return err
}
