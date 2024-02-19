package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/pro-assistance/pro-assister/models"
)

func (s *Service) Register(c context.Context, email string, password string) (uuid.NullUUID, error) {
	item := &models.UserAccount{}
	item.Email = email
	item.Password = password
	err := item.HashPassword()
	if err != nil {
		return uuid.NullUUID{}, err
	}
	err = R.Create(c, item)
	if err != nil {
		return uuid.NullUUID{}, err
	}

	return item.ID, err
}

func (s *Service) Login(c context.Context, email string, password string) (uuid.NullUUID, error) {
	item, err := R.GetByEmail(c, email)
	if (err != nil && err.Error() == sql.ErrNoRows.Error()) || !item.CompareWithHashPassword(password) {
		return uuid.NullUUID{}, errors.New("неверный логин или пароль")
	}
	if err != nil {
		return uuid.NullUUID{}, err
	}
	return item.ID, err
}

func (h *Service) CheckUUID(c context.Context, id string, uid string) error {
	userAccount, err := R.GetByUUID(c, uid)
	if userAccount == nil || err != nil {
		return err
	}
	return R.UpdateUUID(c, id)
}

func (h *Service) UpdatePassword(c context.Context, id string, password string) error {
	return R.UpdatePassword(c, id, password)
}
