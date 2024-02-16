package auth

import (
	"context"
	"mdgkb/tsr-tegister-server-v1/models"
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
