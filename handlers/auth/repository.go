package auth

import (
	"context"

	"github.com/pro-assistance-dev/sprob/models"
	"github.com/uptrace/bun"
)

func (r *Repository) Create(c context.Context, item *models.UserAccount) error {
	_, err := r.helper.DB.IDB(c).NewInsert().Model(item).Exec(c)
	return err
}

func (r *Repository) Get(c context.Context, loginBy string, value string) (*models.UserAccount, error) {
	item := models.UserAccount{}
	err := r.helper.DB.IDB(c).NewSelect().Model(&item).
		Where("lower(?TableAlias.?) = lower(?)", bun.Ident(loginBy), value).
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

func (r *Repository) ConfirmEmail(c context.Context, id string) (err error) {
	_, err = r.helper.DB.IDB(c).NewUpdate().
		Model((*models.UserAccount)(nil)).
		Set("confirm_email = true").
		Where("?TableAlias.id = ?", id).
		Exec(c)
	return err
}

func (r *Repository) UpdateUUID(c context.Context, userID string) (err error) {
	_, err = r.helper.DB.IDB(c).NewUpdate().
		Model((*models.UserAccount)(nil)).
		Set("uuid = uuid_generate_v4()").
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
