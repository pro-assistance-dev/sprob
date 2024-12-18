package models

import "errors"

type AuthData struct {
	Email    string `json:"email"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Phone    string `json:"phone"`

	LoginBy string
	Value   string
}

func (item *AuthData) SetLoginBy() error {
	notEmptyFields := 0
	if len(item.Email) != 0 {
		notEmptyFields++
		item.LoginBy = "email"
		item.Value = item.Email
	}

	if len(item.Login) != 0 {
		notEmptyFields++
		if notEmptyFields > 1 {
			return errors.New("too many non-empty fields")
		}
		item.LoginBy = "phone"
		item.Value = item.Login
	}

	if len(item.Phone) != 0 {
		notEmptyFields++
		if notEmptyFields > 1 {
			return errors.New("too many non-empty fields")
		}
		item.LoginBy = "phone"
		item.Value = item.Phone
	}

	return nil
}
