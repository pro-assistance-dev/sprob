# SPROB — Go-фреймворк для бэкендов (basehandler, baseR, helper, middleware)

Общая библиотека-фреймворк для ВСЕХ бэкендов экосистемы: **rdkb (hr/map/food/leiter), portal, pros**.
Даёт «скелет» сервера: конфиг (viper), helper-сервисы, авто-CRUD + FTSP, миграции, логирование, JWT, почта, чаты.

- Модуль: `github.com/pro-assistance-dev/sprob` (Go 1.23.4)
- Версии в проектах: **все на `v1.0.245`** (rdkb hr/map/food/leiter, portal, pros; выровнено 2026-08-26).
  Обновлять `go get`'ом в каждом проекте + запись в TASKS.md + таблица версий в корневом AGENT.md
- Релиз новой версии: `make update` → `cmd/scripts/update_assister.sh`: автоинкремент `v1.0.x` + commit + tag + push

---

## ⚙️ Как подключается в проект (паттерн единый)

```go
// cmd/server/main.go
import (
    "github.com/pro-assistance-dev/sprob/config"
    "github.com/pro-assistance-dev/sprob/helper"
)

func main() {
    c := config.Init()                      // viper: .env / local.yaml → Config
    h := helper.NewHelper(c)                // все сервисы разом
    helper.Run(migrations.Init(), routing.Init) // запуск: миграции → схемы → HTTP
}
```

- `helper.Run(migrations, routerInitFunc)` — флаги: `-mode=migrate -action=migrate|create|createSql|rollback` (проекты: `go run ./cmd/server/main.go -mode=migrate -action=migrate`).
- При каждом старте применяет **собственные миграции sprob** (`migrations/`) + проектные (переданные первым аргументом).
- `helper.Project.InitSchemas()` — генерирует таблицы из моделей (см. `helpers/project`, `cmd/server/project` в map).

## 🔌 Авто-CRUD (routing + handlers/basehandler)

```go
api, apiNoToken := baseR.Init(r, h)          // две группы: api (с auth) и apiNoToken
api.Use(m.InjectFTSP())

baseR.InitR[models.Event](api)               // GET ''/GET /:id/POST/PUT/DELETE + POST /ftsp + GET /options/:label/:value
baseR.InitR[models.Event](api, baseR.WithHandler(custom.H), baseR.WithWS(ws), baseR.WithKey("events"))
```

- Ключ роута — авто: `Pluralize(ИмяТипа) → kebab-case` (`Event` → `events`), переопределяется `WithKey`.
- Опции: `WithWS(ws)` (добавляет WebSocket-группу), `WithHandler(h)` (кастомные ручки, интерфейс `IHandler`).
- `handlers/basehandler/` — `InitH[T]()`: готовые Handler/Service/Repository на дженериках (`Relationable`).

## 🔧 helper.Helper — что внутри (helper/helper.go)

| Поле | Назначение |
|------|------------|
| `DB` | Bun + pgdriver, пул, лог запросов (logrusbun), `Run`/`DoAction` для миграций |
| `HTTP` | gin-обёртка: ListenAndServe, HandleError, JSON-ответы |
| `Token` | JWT (HS256, `TOKEN_SECRET`/`TOKEN_ACCESS_MINUTES`/`TOKEN_REFRESH_HOURS`) |
| `Broker` | SSE-брокер (`Subscribe`, `SendEvent`) — обновления в реальном времени |
| `Email` | go-simple-mail (SMTP: user/password/server/port/auth) |
| `Uploader` | локальный аплоадер (`UPLOAD_PATH`) |
| `PDF` | генерация PDF |
| `SQL` | построитель SQL + FTSP-инъекция (`InjectFTSP2`) |
| `Templater` | шаблоны DOCX (`TEMPLATES_PATH`) |
| `Validator` | валидация (validator/v10) |
| `Cron` | robfig/cron/v3 — фоновые задачи (напр. datasync в hr) |
| `Project` | схемы/проект (`InitSchemas`, `Name`, `Root`) |
| `Social` | Instagram/YouTube/VK API-ключи |
| `Metabase` | клиент Metabase (`METABASE_URL`/`METABASE_API_KEY`) |
| `Logger` | Logrus + lfshook + file-rotatelogs (`loggerhelper/`) |

## 🛡 middleware/ (порядок: `InjectRequestInfo` = claims + FTSP)

- `InjectFTSP` — парсит `{ftsp:{f,s,p,t}}` из form/query, кэширует по QID в памяти, инъекция в SQL-запрос (пагинация/фильтры/сортировки).
- `InjectClaims` — `user_id`, `domain_ids` из JWT в контекст запроса (`Claim.FromContext(ctx)`).

## 🗄 models/ — общие сущности (таблицы создаёт сам sprob)

`UserAccount` (auth: login/register/restore), `Human`, `Contact`, `Email`, `Phone`, `Website`, `PostAddress`, `Address`,
`FileInfo` (файлы), `ValueType`, `Menu`/`SubMenu`, `FTSPQuery`, `FTSPPreset`, `SearchElement`/`SearchElementMeta`/`SearchModel`/`SearchResult`
(глобальный поиск, `helpers/search`), `BaseModel`, `WithID`.

## 🧩 modules/ — готовые под-приложения (у каждого своя миграция)

| Модуль | Что даёт |
|--------|----------|
| `chats` | Чат: модели `Chat`/`ChatMessage`/`ChatMember`, миграции, WS-роутинг (использует pros: `/ws/connect/:chatId/:userId/...`) |
| `buildings` | Здания/этажи/входы (использует map) |
| `schedule` | **Универсальный календарь-расписание** (29.08.2026): модели `ScheduleDay`/`SchedulePlace`/`Schedule`/`ScheduleSession`/`ScheduleSlot` (таблицы `schedule_days`/`schedule_places`/**`schedule_timetables`**/`schedule_sessions`/`schedule_slots` — НЕ `schedules`, у portal/pros уже есть свои), миграция, авто-CRUD (`InitRoutes(api, h)` → `/schedule-days`, `/schedule-places`, `/schedule-timetables`, `/schedule-sessions`, `/schedule-slots`). Слоты имеют `payload jsonb` для доменных данных. Рассчитан на НОВЫЕ проекты (расписание врачей портала, конференции); фронт — модуль `schedule` в sprof. ⚠️ Не подключать туда, где уже есть свои модели `schedules`/`perfoms` (pros) — коллизия роутов. |
| `documents`, `forms`, `extracts`, `settings` | Документы, формы, выписки, настройки |

## 📁 Структура

```
sprob/
├── config/            # viper-конфиг: .env (DB_*, TOKEN_*, SERVER_*, EMAIL_*, NAME, ROOT, UPLOAD_PATH, METABASE_*, INSTAGRAM_*, YOUTUBE_*, VK_*) + local.yaml
├── helper/            # Helper-агрегатор + Run (запуск сервера/миграций)
├── helpers/           # broker, cron, db, email, http, logger, metabase, pdf, project, search, social, sql, templater, token, uploader, util, validator
├── middleware/        # InjectFTSP / InjectClaims / InjectRequestInfo
├── handlers/          # auth, basehandler, contacts, emails, fileinfos, ftsppresets, humans, menus, metabase, phones, schemas, search, usersaccounts, valuetypes
├── routing/           # Init (api/apiNoToken) + InitR (авто-CRUD)
├── models/            # общие Bun-модели
├── migrations/        # общие миграции (выполняются при каждом старте)
├── modules/           # chats, buildings, documents, forms, extracts, settings
└── cmd/scripts/       # golangci.sh (lint), update_assister.sh (релиз v1.0.x)
```

## 🚀 make

```bash
make lint    # cmd/scripts/golangci.sh
make update  # релиз новой версии v1.0.x (commit + tag + push)
make test    # go test ./...
```
