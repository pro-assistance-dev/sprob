package survey

import (
	"github.com/pro-assistance-dev/sprob/helper"
	"github.com/pro-assistance-dev/sprob/middleware"
	"github.com/pro-assistance-dev/sprob/modules/survey/handlers/invites"
	"github.com/pro-assistance-dev/sprob/modules/survey/handlers/notifications"
	"github.com/pro-assistance-dev/sprob/modules/survey/handlers/public"
	"github.com/pro-assistance-dev/sprob/modules/survey/handlers/reports"
	"github.com/pro-assistance-dev/sprob/modules/survey/handlers/webhooks"

	invitesR "github.com/pro-assistance-dev/sprob/modules/survey/routing/invites"
	notificationsR "github.com/pro-assistance-dev/sprob/modules/survey/routing/notifications"
	publicR "github.com/pro-assistance-dev/sprob/modules/survey/routing/public"
	reportsR "github.com/pro-assistance-dev/sprob/modules/survey/routing/reports"
	webhooksR "github.com/pro-assistance-dev/sprob/modules/survey/routing/webhooks"

	"github.com/gin-gonic/gin"
)

// InitRoutes собирает СОБСТВЕННЫЕ роуты модуля survey (публичный рантайм,
// аналитика, уведомления, приглашения, вебхуки).
//
// ⚠️ CRUD+FTSP самих доменных моделей (Publication/SurveyParam/Response/Theme/…)
// регистрирует вызывающий через baseR.InitR — здесь НЕ дублируем: импорт
// `sprob/routing` из модуля даёт циклический импорт (routing → survey → routing).
//
// api — группа с JWT (админский контур), apiNoToken — без JWT (респонденты).
func InitRoutes(api, apiNoToken *gin.RouterGroup, h *helper.Helper) {
	m := middleware.CreateMiddleware(h)

	// Публичный рантайм опросника — БЕЗ авторизации (внешние респонденты):
	// GET /api/public/survey/:slug, POST /api/public/survey/:slug/responses.
	publicR.Init(apiNoToken.Group("/public/survey"), public.Init(h))

	api.Use(m.InjectFTSP())

	// Аналитика/выгрузка по публикации: /api/publications/:id/summary|export|responses.
	reportsR.Init(api.Group("/publications"), reports.Init(h))
	// Уведомления публикации («кому и при каких условиях»): /api/publications/:id/notifications.
	notificationsR.Init(api.Group("/publications"), notifications.Init(h))
	// Рассылка приглашений (персональные ссылки по email): /api/publications/:id/invites/send.
	invitesR.Init(api.Group("/publications"), invites.Init(h))
	// Вебхуки публикации (внешние интеграции): /api/publications/:id/webhooks.
	webhooksR.Init(api.Group("/publications"), webhooks.Init(h))
}
