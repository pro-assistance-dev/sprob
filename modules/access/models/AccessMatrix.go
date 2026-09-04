package models

import "github.com/uptrace/bun"

// AccessMatrix - строка матрицы доступа (Этап 2, RACI):
// entity.field x role_code -> access (” / R / W / M).
type AccessMatrix struct {
	bun.BaseModel `bun:"access_matrix,alias:access_matrix"`
	Relationable

	Entity    string `json:"entity" bun:"entity"`
	Field     string `json:"field" bun:"field"`
	RoleCode  string `json:"roleCode" bun:"role_code"`
	Access    string `json:"access" bun:"access"`
	OwnerRole string `json:"ownerRole,omitempty" bun:"owner_role"`
	// Периодичность актуализации поля (разово/при изменении/ежемесячно/...)
	Periodicity string `json:"periodicity,omitempty" bun:"periodicity"`
	// Последствие незаполнения (кратко)
	Consequence string `json:"consequence,omitempty" bun:"consequence"`
}

type AccessMatrixes []*AccessMatrix

type AccessMatrixesWithCount struct {
	AccessMatrixes AccessMatrixes `json:"items"`
	Count          int            `json:"count"`
}
