package fileinfos

import (
	"context"
	"pro-assister/models"

	"github.com/google/uuid"
)

func (r *Repository) Create(c context.Context, item *models.FileInfo) (err error) {
	_, err = R.helper.DB.IDB(c).NewInsert().Model(item).Exec(c)
	return err
}

func (r *Repository) Get(c context.Context, id string) (*models.FileInfo, error) {
	item := models.FileInfo{}
	err := R.helper.DB.IDB(c).NewSelect().Model(&item).
		Where("file_infos.id = ?", id).Scan(c)
	return &item, err
}

func (r *Repository) Update(c context.Context, item *models.FileInfo) (err error) {
	_, err = R.helper.DB.IDB(c).NewUpdate().Model(item).Where("id = ?", item.ID).Exec(c)
	return err
}

func (r *Repository) Upsert(c context.Context, item *models.FileInfo) (err error) {
	_, err = R.helper.DB.IDB(c).NewInsert().On("conflict (id) do update").
		Set("original_name = EXCLUDED.original_name").
		Set("file_system_path = EXCLUDED.file_system_path").
		Model(item).
		Exec(c)
	return err
}

func (r *Repository) CreateMany(c context.Context, items models.FileInfos) (err error) {
	_, err = R.helper.DB.IDB(c).NewInsert().Model(&items).Exec(c)
	return err
}

func (r *Repository) UpsertMany(c context.Context, items models.FileInfos) (err error) {
	_, err = R.helper.DB.IDB(c).NewInsert().On("conflict (id) do update").
		Set("original_name = EXCLUDED.original_name").
		Set("file_system_path = EXCLUDED.file_system_path").
		Model(&items).
		Exec(c)
	return err
}

func (r *Repository) Delete(c context.Context, id uuid.NullUUID) (err error) {
	return err
}
