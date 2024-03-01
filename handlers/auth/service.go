package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/pro-assistance/pro-assister/models"
)

func (s *Service) Register(c context.Context, email string, password string) (uuid.NullUUID, bool, error) {
	item := &models.UserAccount{}
	item.Email = email
	item.Password = password

	existingUserAccount, _ := R.GetByEmail(c, item.Email)
	if existingUserAccount.ID.Valid {
		emailStruct := struct {
			RestoreLink string
			Host        string
		}{
			s.helper.HTTP.GetRestorePasswordURL(existingUserAccount.ID.UUID.String(), existingUserAccount.UUID.String()),
			s.helper.HTTP.Host,
		}

		mail, err := s.helper.Templater.ParseTemplate(emailStruct, "email/refreshToDouble.gohtml")
		if err != nil {
			return uuid.NullUUID{}, false, err
		}
		err = s.helper.Email.SendEmail([]string{item.Email}, "Восстановление пароля", mail)
		if err != nil {
			return uuid.NullUUID{}, false, err
		}
		return uuid.NullUUID{}, true, nil
	} else {
		emailStruct := struct {
			Host string
		}{
			s.helper.HTTP.Host,
		}

		mail, err := s.helper.Templater.ParseTemplate(emailStruct, "email/successRegistration.gohtml")
		if err != nil {
			return uuid.NullUUID{}, false, err
		}
		err = s.helper.Email.SendEmail([]string{item.Email}, "Успешная регистрация на сайте", mail)
		if err != nil {
			return uuid.NullUUID{}, false, err
		}
	}

	err := item.HashPassword()
	if err != nil {
		return uuid.NullUUID{}, false, err
	}
	err = R.Create(c, item)
	if err != nil {
		return uuid.NullUUID{}, false, err
	}

	return item.ID, false, err
}

func (s *Service) Login(c context.Context, email string, password string) (uuid.NullUUID, error, error) {
	item, err := R.GetByEmail(c, email)
	if (err != nil && err.Error() == sql.ErrNoRows.Error()) || !item.PasswordEqWithHashed(password) {
		return uuid.NullUUID{}, errors.New("неверный логин или пароль"), err
	}
	if err != nil {
		return uuid.NullUUID{}, err, err
	}
	return item.ID, err, err
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
