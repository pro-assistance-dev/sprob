package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/pro-assistance-dev/sprob/models"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Register(c context.Context, email string, password string, itemID uuid.NullUUID) (*models.UserAccount, bool, error) {
	item := &models.UserAccount{}
	item.Email = email
	item.Password = password
	item.ItemID = itemID

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
			return nil, false, err
		}
		go s.helper.Email.SendEmail([]string{item.Email}, "Восстановление пароля", mail)
		// if err != nil {
		// 	return uuid.NullUUID{}, false, err
		// }
		return nil, true, nil
	}

	err := item.HashPassword()
	if err != nil {
		return nil, false, err
	}
	err = R.Create(c, item)
	if err != nil {
		return nil, false, err
	}

	err = s.SendConfirmEmailMail(item.ID.UUID.String(), item.Email)
	if err != nil {
		return nil, false, err
	}
	return item, false, err
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
	go s.helper.Email.SendEmail([]string{email}, "Подтверждение email", mail)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (s *Service) Login(c context.Context, authData *models.AuthData) (*models.UserAccount, error) {
	err := authData.SetLoginBy()
	if err != nil {
		return nil, err
	}
	item, err := R.Get(c, authData.LoginBy, authData.Value)
	if (err != nil && err.Error() == sql.ErrNoRows.Error()) || !item.PasswordEqWithHashed(authData.Password) {
		return nil, errors.New("неверный логин или пароль")
	}
	if err != nil {
		return nil, err
	}
	return item, err
}

func (h *Service) ConfirmEmail(c context.Context, id string) error {
	return R.ConfirmEmail(c, id)
}

func (h *Service) EmailIsConfirm(c context.Context, email string) error {
	item, err := R.EmailIsConfirm(c, email)
	if err == nil {
		return nil
	}
	if item != nil && item.ID.Valid {
		_ = h.SendConfirmEmailMail(item.ID.UUID.String(), email)
	}
	return err
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
