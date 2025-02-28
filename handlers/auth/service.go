package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/pro-assistance-dev/sprob/models"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Register(c context.Context, email string, password string) (uuid.NullUUID, bool, error) {
	item := &models.UserAccount{}
	item.Email = email
	item.Password = password

	existingUserAccount, _ := R.Get(c, "email", item.Email)
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
	}

	err := item.HashPassword()
	if err != nil {
		return uuid.NullUUID{}, false, err
	}
	err = R.Create(c, item)
	if err != nil {
		return uuid.NullUUID{}, false, err
	}

	err = s.SendConfirmEmailMail(item.ID.UUID.String(), item.Email)
	if err != nil {
		return uuid.NullUUID{}, false, err
	}
	return item.ID, false, err
}

func (s *Service) SendConfirmEmailMail(id, email string) error {
	emailStruct := struct {
		Host string
		ID   string
	}{
		s.helper.HTTP.Host,
		id,
	}

	mail, err := s.helper.Templater.ParseTemplate(emailStruct, "email/successRegistration.gohtml")
	if err != nil {
		return err
	}
	err = s.helper.Email.SendEmail([]string{email}, "Подтверждение email", mail)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Login(c context.Context, authData *models.AuthData) (uuid.NullUUID, error) {
	err := authData.SetLoginBy()
	if err != nil {
		return uuid.NullUUID{}, err
	}
	item, err := R.Get(c, authData.LoginBy, authData.Value)
	if (err != nil && err.Error() == sql.ErrNoRows.Error()) || !item.PasswordEqWithHashed(authData.Password) {
		return uuid.NullUUID{}, errors.New("неверный логин или пароль")
	}
	if err != nil {
		return uuid.NullUUID{}, err
	}
	return item.ID, err
}

func (h *Service) ConfirmEmail(c context.Context, id string) error {
	return R.ConfirmEmail(c, id)
}

func (h *Service) EmailIsConfirm(c context.Context, email string) error {
	return R.EmailIsConfirm(c, email)
}

func (h *Service) CheckUUID(c context.Context, id string, uid string) error {
	userAccount, err := R.GetByUUID(c, uid)
	if userAccount == nil || err != nil {
		return err
	}
	return R.UpdateUUID(c, id)
}

func (h *Service) UpdatePassword(c context.Context, id string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return R.UpdatePassword(c, id, string(hashedPassword))
}
