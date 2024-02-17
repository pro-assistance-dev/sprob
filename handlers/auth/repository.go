package auth

import (
	"context"

	"github.com/pro-assistance/pro-assister/models"
)

func (r *Repository) Create(c context.Context, item *models.UserAccount) error {
	_, err := r.helper.DB.IDB(c).NewInsert().Model(item).Exec(c)
	return err
}

func (r *Repository) GetByEmail(c context.Context, email string) (*models.UserAccount, error) {
	item := models.UserAccount{}
	err := r.helper.DB.IDB(c).NewSelect().Model(&item).
		Where("?TableAlias.email = ?", email).
		Scan(c)
	return &item, err
}

func (r *Repository) GetByUUID(c context.Context, uid string) (*models.UserAccount, error) {
	item := models.UserAccount{}
	err := r.helper.DB.IDB(c).NewSelect().Model(&item).
		Where("?TableAlias.uuid = ?", uid).
		Scan(c)
	return &item, err
}

func (r *Repository) UpdateUUID(c context.Context, userID string) (err error) {
	_, err = r.helper.DB.IDB(c).NewUpdate().
		Model((*models.UserAccount)(nil)).
		Set("check_uuid = uuid_generate_v4()").
		Where("?TableAlias.id = ?", userID).
		Exec(c)
	return err
}

func (r *Repository) UpdatePassword(c context.Context, id string, password string) (err error) {
	_, err = r.helper.DB.IDB(c).NewUpdate().
		Model((*models.UserAccount)(nil)).
		Set("password = ?", password).
		Where("?TableAlias.id = ?", id).
		Exec(c)
	return err
}
