package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

type UserAccount struct {
	bun.BaseModel `bun:"users_accounts,alias:users_accounts"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	UUID          uuid.UUID     `bun:"type:uuid,nullzero,notnull,default:uuid_generate_v4()"  json:"uuid"` // для восстановления пароля - обеспечивает уникальность страницы на фронте

	Email        string        `json:"email"`
	Login        string        `json:"login"`
	Password     string        `json:"password"`
	ItemID       uuid.NullUUID `bun:"type:uuid" json:"itemId"`
	Phone        string        `json:"phone"`
	ConfirmEmail bool          `json:"confirmEmail"`
}

type UsersAccounts []*UserAccount

type UsersAccountsWithCount struct {
	UsersAccounts UsersAccounts `json:"items"`
	Count         int           `json:"count"`
}

func (item *UserAccount) CompareWithUUID(externalUUID string) bool {
	return item.UUID.String() == externalUUID
}

func (item *UserAccount) HashPassword() error {
	hash, err := bcrypt.GenerateFromPassword([]byte(item.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	item.Password = string(hash)
	return nil
}

func (item *UserAccount) PasswordEqWithHashed(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(item.Password), []byte(password)) == nil
}
