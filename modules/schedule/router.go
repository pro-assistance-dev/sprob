// Модуль schedule (sprob) — универсальный календарь-расписание.
//
// Регистрирует авто-CRUD (baseR.InitR) для сущностей:
//
//	/schedule-days     — дни календаря (привязка к владельцу: eventId/doctorId через ownerId/ownerType)
//	/schedule-places   — места (залы/кабинеты/аудитории)
//	/schedule-timetables — расписание места на день (секции + слоты); НЕ `schedules`
//	                       (у portal/pros уже есть свои роуты/таблицы `schedules`)
//	/schedule-sessions — секции (интервал с вложенными слотами)
//	/schedule-slots    — слоты (доклады/приёмы/лекции; доменные данные — в payload jsonb)
//
// Подключение в проекте:
//
//	// migrations/main.go
//	scheduleM "github.com/pro-assistance-dev/sprob/modules/schedule/migrations"
//	res = append(res, migrations, chats.Init(), scheduleM.Init())
//
//	// routing/router.go
//	schedule.InitRoutes(api, h)
//
// ⚠️ У проекта не должно быть собственных моделей с теми же именами таблиц/роутов
// (schedules и т.п.) — иначе коллизия роутов. Модуль рассчитан на новые проекты
// (расписание врачей на портале, конференции) или на проекты без своего расписания.
package schedule

import (
	"github.com/pro-assistance-dev/sprob/modules/schedule/models"

	helperPack "github.com/pro-assistance-dev/sprob/helper"
	baseR "github.com/pro-assistance-dev/sprob/routing"

	"github.com/gin-gonic/gin"
)

func InitRoutes(api *gin.RouterGroup, _ *helperPack.Helper) {
	baseR.InitR[models.ScheduleDay](api)
	baseR.InitR[models.SchedulePlace](api)
	baseR.InitR[models.Schedule](api, baseR.WithKey("schedule-timetables"))
	baseR.InitR[models.ScheduleSession](api)
	baseR.InitR[models.ScheduleSlot](api)
}

