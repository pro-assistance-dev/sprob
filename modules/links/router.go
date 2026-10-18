package links

import (
	"github.com/gin-gonic/gin"
	"github.com/pro-assistance-dev/sprob/helper"
	"github.com/pro-assistance-dev/sprob/middleware"
	"github.com/pro-assistance-dev/sprob/modules/links/handlers/links"
	"github.com/pro-assistance-dev/sprob/modules/links/models"
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
func InitRoutes(r *gin.Engine, api, apiNoToken *gin.RouterGroup, h *helper.Helper, opts ...links.Option) *links.Handler {
	handler := links.New(h, opts...)

	// Публичная обёртка — до всяких мидлвар с токеном.
	r.GET("/r/:code", handler.Redirect)

	m := middleware.CreateMiddleware(h)
	api.Use(m.InjectFTSP())

	baseR.InitR[models.Link](api)
	baseR.InitR[models.LinkDomain](api)
	baseR.InitR[models.LinkHit](api)

	group := api.Group("/links")
	group.POST("/batch", handler.BatchEnsure)
	group.GET("/:id/qr", handler.QrCode)
	group.GET("/:id/hits", handler.Hints)

	api.PUT("/link-domains/default", handler.SwitchDomain)
	api.DELETE("/link-hits", handler.PurgeHitsHandler)

	return handler
}
