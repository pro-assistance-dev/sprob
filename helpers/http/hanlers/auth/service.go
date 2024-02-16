package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/pro-assistance/pro-assister/models"
)

func (s *Service) Register(c context.Context, email string, password string) (uuid.NullUUID, error) {
	item := &models.UserAccount{}
	item.Email = email
	item.Password = password
	item.HashPassword()

	err := R.Create(c, item)
	if err != nil {
		return uuid.NullUUID{}, err
	}

	return item.ID, err
}

func (s *Service) Login(c context.Context, email string, password string) (uuid.NullUUID, error) {
	item, err := R.GetByEmail(c, email)
	if err != nil {
		return uuid.NullUUID{}, err
	}
	if !item.CompareWithHashPassword(password) {
		return uuid.NullUUID{}, errors.New("wrong email or password")
	}
	return item.ID, err
}
