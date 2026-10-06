package survey

import (
	"github.com/pro-assistance-dev/sprob/helper"
	"github.com/pro-assistance-dev/sprob/middleware"
	"github.com/pro-assistance-dev/sprob/modules/survey/handlers/invites"
	"github.com/pro-assistance-dev/sprob/modules/survey/handlers/notifications"
	"github.com/pro-assistance-dev/sprob/modules/survey/handlers/public"
	"github.com/pro-assistance-dev/sprob/modules/survey/handlers/reports"
	"github.com/pro-assistance-dev/sprob/modules/survey/handlers/webhooks"
	"github.com/pro-assistance-dev/sprob/modules/survey/models"
	baseR "github.com/pro-assistance-dev/sprob/routing"

	invitesR "github.com/pro-assistance-dev/sprob/modules/survey/routing/invites"
	notificationsR "github.com/pro-assistance-dev/sprob/modules/survey/routing/notifications"
	publicR "github.com/pro-assistance-dev/sprob/modules/survey/routing/public"
	reportsR "github.com/pro-assistance-dev/sprob/modules/survey/routing/reports"
	webhooksR "github.com/pro-assistance-dev/sprob/modules/survey/routing/webhooks"

	"github.com/gin-gonic/gin"
)

// InitRoutes собирает роуты ОПРОСНИКА: CRUD+FTSP доменных моделей
// (публикации/параметры/ответы/темы/уведомления/вебхуки), свои ручки
// (аналитика/уведомления/приглашения/вебхуки) и публичный рантайм.
//
// Подключение в проекте:
//
//	// migrations/main.go
//	res = append(res, forms.Init(), surveyM.Init())
//
//	// routing/router.go
//	survey.InitRoutes(api, apiNoToken, h)
//
// ⚠️ Модуль НЕ подключается глобально в sprob/routing: домен опционален,
// иначе все сервисы получили бы чужые роуты (/publications, /themes, ...).
// api — группа с JWT (админка), apiNoToken — без JWT (респонденты).
func InitRoutes(api, apiNoToken *gin.RouterGroup, h *helper.Helper) {
	m := middleware.CreateMiddleware(h)

	// Публичный рантайм опросника — БЕЗ авторизации (внешние респонденты):
	// GET /api/public/survey/:slug, POST /api/public/survey/:slug/responses.
	publicR.Init(apiNoToken.Group("/public/survey"), public.Init(h))

	api.Use(m.InjectFTSP())

	// Публикации форм (ссылка, окно приёма, анонимность, лимит) — CRUD + FTSP.
	baseR.InitR[models.Publication](api)
	// Аналитика/выгрузка по публикации: /api/publications/:id/summary|export|responses.
	reportsR.Init(api.Group("/publications"), reports.Init(h))
	// Уведомления публикации («кому и при каких условиях»): /api/publications/:id/notifications.
	notificationsR.Init(api.Group("/publications"), notifications.Init(h))
	// Рассылка приглашений (персональные ссылки по email): /api/publications/:id/invites/send.
	invitesR.Init(api.Group("/publications"), invites.Init(h))
	// Вебхуки публикации (внешние интеграции): /api/publications/:id/webhooks.
	webhooksR.Init(api.Group("/publications"), webhooks.Init(h))
	// Параметры анкеты (URL → метаданные/префилл) — CRUD + FTSP.
	baseR.InitR[models.SurveyParam](api)
	// Ответы на публикации (обёртка над form_fills) — CRUD + FTSP.
	baseR.InitR[models.Response](api)
	// Приглашения (персональные ссылки для неанонимных публикаций) — CRUD + FTSP.
	baseR.InitR[models.Invite](api)
	// Уведомления (кому и при каких условиях слать email после ответа) — CRUD + FTSP.
	baseR.InitR[models.Notification](api)
	baseR.InitR[models.NotificationRule](api)
	// Журнал отправок уведомлений (только чтение в UI) — CRUD + FTSP.
	baseR.InitR[models.NotificationLog](api)
	// Темы оформления формы (CRUD + FTSP).
	baseR.InitR[models.Theme](api)
	// Вебхуки после ответа (внешние интеграции) — CRUD + FTSP + журнал.
	baseR.InitR[models.Webhook](api)
	baseR.InitR[models.WebhookRule](api)
	baseR.InitR[models.WebhookLog](api)
}
