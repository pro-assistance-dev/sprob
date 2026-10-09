package notify

import (
	"github.com/pro-assistance-dev/sprob/helper"
	"github.com/pro-assistance-dev/sprob/middleware"
	"github.com/pro-assistance-dev/sprob/notify/handlers/rules"
	rulesR "github.com/pro-assistance-dev/sprob/notify/routing/rules"

	"github.com/gin-gonic/gin"
)

// InitRoutes монтирует админский CRUD правил уведомлений: /api/notify-rules
// (список + FTSP + карточка). Сам Notifier создаёт потребитель (`notify.New(h)`)
// и публикует события; роуты нужны только для правки правил.
//
// ⚠️ Модуль подключается ПОТРЕБИТЕЛЕМ (как survey/schedule), не глобально:
// домен опционален и добавляет таблицы/роуты.
func InitRoutes(api *gin.RouterGroup, h *helper.Helper) {
	api.Use(middleware.CreateMiddleware(h).InjectFTSP())
	rulesR.Init(api.Group("/notify-rules"), rules.Init(h))
}
