package links

import (
	"github.com/gin-gonic/gin"
	"github.com/pro-assistance-dev/sprob/helper"
	"github.com/pro-assistance-dev/sprob/middleware"
	linkshc "github.com/pro-assistance-dev/sprob/modules/links/handlers/links"
	"github.com/pro-assistance-dev/sprob/modules/links/models"
	qr "github.com/pro-assistance-dev/sprob/modules/links/qr"
	baseR "github.com/pro-assistance-dev/sprob/routing"
)

// InitRoutes собирает роуты модуля управляемых ссылок:
//
//	GET  /r/:code                      — ПУБЛИЧНАЯ обёртка (302 на цель, без токена);
//	CRUD /api/links (+FTSP)            — реестр ссылок;
//	POST /api/links/batch              — пакетный выпуск (до 6000 QR);
//	GET  /api/links/:id/qr             — QR обёртки (svg/png, генерирует сервер);
//	GET  /api/links/:id/hits           — агрегаты переходов;
//	CRUD /api/link-domains, /api/link-hits — домены обёртки и журнал.
//
// Подключение в проекте:
//
//	links.InitRoutes(r, api, apiNoToken, h, links.WithResolver(myResolver))
//
// ⚠️ В sprob/routing глобально НЕ подключается: домен опционален, иначе чужие
// сервисы получили бы /links и публичный /r/*.
//
// Резолвер цели передаётся через `linkshc.WithResolver` (пакет handlers/links).
func InitRoutes(r *gin.Engine, api, apiNoToken *gin.RouterGroup, h *helper.Helper, opts ...linkshc.Option) *linkshc.Handler {
	handler := linkshc.New(h, opts...)

	// Публичная обёртка — до всяких мидлвар с токеном.
	r.GET("/r/:code", handler.Redirect)

	m := middleware.CreateMiddleware(h)
	api.Use(m.InjectFTSP())

	// ⚠️ ПУТЬ АДМИНКИ — /api/link-registry, а не /api/links: авто-CRUD baseR
	// регистрирует /links (множественное от Link), и на общем префиксе два
	// набора ручек мешали друг другу. Свои ручки реестра — на отдельном префиксе.
	//
	// ⚠️ POST "" отдельно НЕ регистрируем (gin падал бы на дубле пути:
	// `handlers are already registered for path '/api/link-registry'`), но и
	// штатный baseR-Create не годится: он вставляет через
	// `bun.NewInsert().Model`, а `link_code` у модели `scanonly` → Postgres
	// отдаёт `null value in column "link_code" ... violates not-null` (23502).
	// Решение: baseR монтируется с НАШИМ handler'ом (WithHandler) — его Create
	// идёт через InsertGenerated (код генерится SQL-вставкой).
	group := api.Group("/link-registry")
	group.POST("/batch", handler.BatchEnsure)
	registry := baseR.InitR[models.Link](api, baseR.WithKey("link-registry"), baseR.WithHandler(handler))
	// Идемпотентное создание/поиск по цели и служебные ручки реестра.
	registry.POST("/ensure", handler.CreateLink)
	registry.GET("/:id/qr", handler.QrCode)
	registry.GET("/:id/hits", handler.Hints)
	baseR.InitR[models.LinkDomain](api)
	baseR.InitR[models.LinkHit](api)

	api.PUT("/link-domains/default", handler.SwitchDomain)
	api.DELETE("/link-hits", handler.PurgeHitsHandler)

	return handler
}

// WithResolver — прокси к опции обработчика: потребитель не импортирует
// внутренний пакет handlers/links ради одной настройки.
func WithResolver(r qr.Resolver) linkshc.Option {
	return linkshc.WithResolver(r)
}
