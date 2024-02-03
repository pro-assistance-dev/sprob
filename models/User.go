package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"users,select:users_view,alias:users_view"`

	ID       uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Email    string        `json:"email"`
	UUID     uuid.UUID     `bun:"type:uuid,nullzero,notnull,default:uuid_generate_v4()"  json:"uuid"` // для восстановления пароля - обеспечивает уникальность страницы на фронте
	Phone    string        `json:"phone"`
	Password string        `json:"password"`
	IsActive bool          `json:"isActive"`
}

type Users []*User
