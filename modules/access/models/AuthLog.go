package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// AuthLog - журнал входа-выхода в систему (Задача 5, map.md):
// кто/когда/какая роль/учётка, IP, user-agent; operation: login / logout / heartbeat.
//
// Записи пишутся с клиента (события auth login/logout -> POST /api/auth/log):
//   - login          - вход (явный вход через keycloak или восстановление сессии при старте);
//   - logout         - выход; reason: user (кнопка «Выйти») / tab-close (закрытие вкладки) /
//     token-expired (истёк и не обновился токен);
//   - heartbeat      - периодический сигнал активности (раз в 5 минут) - по нему видно
//     «живые» сессии, даже если закрытие вкладки не дошло до сервера.
type AuthLog struct {
	bun.BaseModel `bun:"auth_log,alias:auth_log"`
	Relationable

	CreatedAt time.Time     `json:"createdAt" bun:"created_at"`
	UserID    uuid.NullUUID `json:"userId,omitempty" bun:"user_id"`
	UserName  string        `json:"userName,omitempty" bun:"user_name"`
	RoleCode  string        `json:"roleCode,omitempty" bun:"role_code"`
	Operation string        `json:"operation,omitempty" bun:"operation"` // login / logout / heartbeat
	Reason    string        `json:"reason,omitempty" bun:"reason"`       // logout: user / tab-close / token-expired
	IP        string        `json:"ip,omitempty" bun:"ip"`
	UserAgent string        `json:"userAgent,omitempty" bun:"user_agent"`
}

type AuthLogs []*AuthLog

type AuthLogsWithCount struct {
	AuthLogs AuthLogs `json:"items"`
	Count    int      `json:"count"`
}
