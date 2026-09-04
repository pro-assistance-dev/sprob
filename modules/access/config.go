package access

import "os"

// Конфигурация движка (env, значения по умолчанию — наследие rdkb/map, С4.1).
// Проекты переопределяют при необходимости (например, hr: ACCESS_ADMIN_ROLE=ADMIN).

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// AdminRole - роль с полным доступом ко всем полям (по умолчанию R00_ADMIN из map).
var AdminRole = envOr("ACCESS_ADMIN_ROLE", "R00_ADMIN")

// appClient - clientId приложения в keycloak: роли проекта назначаются как клиентские
// роли именно этого клиента (resource_access.<appClient>.roles).
var appClient = envOr("ACCESS_APP_CLIENT", "map-app")
