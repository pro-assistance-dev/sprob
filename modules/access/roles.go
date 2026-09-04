package access

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pro-assistance-dev/sprob/helper"
)

// UserCtx - результат разбора токена: роли и идентификатор пользователя.
type UserCtx struct {
	UserID   string
	UserName string
	Roles    []string
	// TokenValid - подпись токена проверена (JWKS/TOKEN_SECRET) и валидна.
	TokenValid bool
	// TokenInvalid - в заголовке был токен, но его подпись/срок не прошли проверку.
	// Такой запрос НЕ считается аутентифицированным (AccessControl отвечает 401).
	TokenInvalid bool
	// TokenUnverified - JWKS был недоступен и включён fail-open: claims разобраны
	// без проверки подписи (диагностический флаг, см. jwks.go).
	TokenUnverified bool
}

// systemRoles - служебные роли keycloak, не относящиеся к матрице FM.
// В токене они попадают в realm_access.roles / resource_access.account.roles
// и не должны влиять на доступ (в матрице access_matrix их нет).
var systemRoles = map[string]bool{
	"OFFLINE_ACCESS":       true,
	"UMA_AUTHORIZATION":    true,
	"DEFAULT-ROLES-RDKB":   true,
	"MANAGE-ACCOUNT":       true,
	"MANAGE-ACCOUNT-LINKS": true,
	"VIEW-PROFILE":         true,
}

// RolesFromRequest извлекает роли пользователя из запроса:
//  1. JWT в заголовке `token` (keycloak: realm_access.roles; sprob: claims.roles / user_id);
//  2. таблица user_roles по user_id из токена.
//
// Б1: подпись токена проверяется (JWKS keycloak / TOKEN_SECRET для HS256) — см. jwks.go.
// Токен с невалидной подписью → UserCtx{TokenInvalid: true} (claims не разбираются).
// Запросы без токена остаются анонимными (TokenInvalid=false, ролей нет).
func RolesFromRequest(c context.Context, h *helper.Helper, r *http.Request) UserCtx {
	uc := UserCtx{}
	raw := r.Header.Get("token")
	if raw == "" {
		raw = r.Header.Get("Authorization")
		raw = strings.TrimPrefix(raw, "Bearer ")
	}
	if raw == "" {
		return uc
	}

	claims, status := verifyRequestToken(raw)
	switch status {
	case tokenStatusValid:
		fillUserCtx(&uc, claims)
		uc.TokenValid = true
	case tokenStatusUnverified:
		// Fail-open: JWKS недоступен, claims без подписи (как до Б1) + WARNING в лог.
		fillUserCtx(&uc, claims)
		uc.TokenUnverified = true
	case tokenStatusDisabled:
		// JWT_VERIFY=false: аварийный режим — поведение до Б1.
		fillUserCtx(&uc, claims)
	case tokenStatusInvalid:
		// Токен есть, но подпись невалидна — не доверяем claims вообще.
		uc.TokenInvalid = true
		return uc
	case tokenStatusNone:
		return uc
	}

	// Фолбэк: роли из таблицы user_roles
	if uc.UserID != "" && len(uc.Roles) == 0 {
		if roles, err := RolesFromDB(c, h, uc.UserID); err == nil {
			uc.Roles = roles
		}
	}
	return uc
}

// fillUserCtx разбирает claims токена в UserCtx (user_id, имя, роли).
func fillUserCtx(uc *UserCtx, claims map[string]interface{}) {
	if v, ok := claims["user_id"].(string); ok {
		uc.UserID = v
	}
	if v, ok := claims["sub"].(string); ok && uc.UserID == "" {
		uc.UserID = v
	}
	if v, ok := claims["name"].(string); ok {
		uc.UserName = v
	}
	if v, ok := claims["preferred_username"].(string); ok && uc.UserName == "" {
		uc.UserName = v
	}
	if ra, ok := claims["realm_access"].(map[string]interface{}); ok {
		if roles, ok := ra["roles"].([]interface{}); ok {
			for _, role := range roles {
				if s, ok := role.(string); ok {
					uc.Roles = append(uc.Roles, s)
				}
			}
		}
	}
	// Роли клиента (resource_access): в keycloak FM-роли (r03_rem и т.п.) назначены
	// как клиентские роли map-app — в realm_access их нет.
	// Читаем ТОЛЬКО роли нашего клиента (map-app): если пользователю назначена
	// роль в другом приложении (hr-app и т.п.), она не должна давать доступ в map.
	// Фолбэк: если map-app в токене нет вовсе (старые токены/другие клиенты) —
	// собираем роли из всех resource_access клиентов.
	if ra, ok := claims["resource_access"].(map[string]interface{}); ok {
		if app, ok := ra[appClient].(map[string]interface{}); ok {
			if roles, ok := app["roles"].([]interface{}); ok {
				for _, role := range roles {
					if s, ok := role.(string); ok {
						uc.Roles = append(uc.Roles, s)
					}
				}
			}
		} else {
			for _, client := range ra {
				cm, ok := client.(map[string]interface{})
				if !ok {
					continue
				}
				if roles, ok := cm["roles"].([]interface{}); ok {
					for _, role := range roles {
						if s, ok := role.(string); ok {
							uc.Roles = append(uc.Roles, s)
						}
					}
				}
			}
		}
	}
	if roles, ok := claims["roles"].([]interface{}); ok {
		for _, role := range roles {
			if s, ok := role.(string); ok {
				uc.Roles = append(uc.Roles, s)
			}
		}
	}
	// Дубликаты
	seen := map[string]bool{}
	filtered := uc.Roles[:0]
	for _, r := range uc.Roles {
		if !seen[r] {
			seen[r] = true
			filtered = append(filtered, r)
		}
	}
	uc.Roles = filtered
	// Нормализация: keycloak отдаёт роли в нижнем регистре (r03_rem),
	// матрица и реестр используют верхний (R03_REM).
	for i := range uc.Roles {
		uc.Roles[i] = strings.ToUpper(uc.Roles[i])
	}
	// Убираем служебные роли keycloak (offline_access, default-roles-* и т.п.) —
	// в матрице их нет, а на клиенте они не должны считаться «ролью FM».
	filtered = uc.Roles[:0]
	for _, r := range uc.Roles {
		if !systemRoles[r] {
			filtered = append(filtered, r)
		}
	}
	uc.Roles = filtered
}

// RolesFromDB читает назначенные роли пользователя из user_roles.
func RolesFromDB(ctx context.Context, h *helper.Helper, userID string) ([]string, error) {
	var roleCodes []string
	err := h.DB.IDB(ctx).NewSelect().
		Table("user_roles").
		Column("role_code").
		Where("user_id = ?", userID).
		Scan(ctx, &roleCodes)
	if err != nil {
		return nil, err
	}
	for i := range roleCodes {
		roleCodes[i] = strings.ToUpper(roleCodes[i])
	}
	return roleCodes, nil
}

// decodeJWTClaims декодирует payload JWT без проверки подписи.
func decodeJWTClaims(raw string) map[string]interface{} {
	claims := map[string]interface{}{}
	parts := strings.Split(raw, ".")
	if len(parts) < 2 {
		return claims
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return claims
	}
	_ = json.Unmarshal(payload, &claims)
	return claims
}
