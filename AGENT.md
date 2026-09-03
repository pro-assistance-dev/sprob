# SPROB — Go-фреймворк для бэкендов (basehandler, baseR, helper, middleware)

Общая библиотека-фреймворк для ВСЕХ бэкендов экосистемы: **rdkb (hr/map/food/leiter/incident), portal, pros, ferma**.
Даёт «скелет» сервера: конфиг (viper), helper-сервисы, авто-CRUD + FTSP, миграции, логирование, JWT, почта, чаты.

- Модуль: `github.com/pro-assistance-dev/sprob` (Go 1.23.4)
- Версии в проектах: **все на `v1.0.251`** (01.09; таблица — в корневом `AGENT.md` воркспейса).
  Обновлять `go get`'ом в каждом проекте + запись в TASKS.md проекта + таблица версий.
- Релиз новой версии: `make update` → `cmd/scripts/update_assister.sh`: автоинкремент `v1.0.x` + commit + tag + push

---

## 🤖 Правила работы агента (обязательные)

Общие для всех проектов `~/prog/work`; канон — корневой [`../AGENT.md`](../AGENT.md).
Блок синхронизируется по проектам (rdkb, portal, pros, ferma, sprob, sprof, rdkb-proxy).

### 🚫 Аналитический паралич — запрещён

- План — не больше **5–10 строк**. После этого — **немедленно писать код**.
- **Запрещены** файлы планирования (`mind.md`, `plan.md`, `analysis.md`) и любые логи
  размышлений на диске. Планирование — короткими пунктами в `TASKS.md` или в голове.
- Планирование дольше ~5 минут — **стоп и писать код**: состояние уже в git.

### 🔄 Инкрементальная разработка

- Задача → **минимальный работающий шаг** → проверка → следующий шаг.
- Шаг №1 — **самая простая версия**, без «запасных» фич; расширение — только после
  того, как шаг №1 заработал.

### 🚫 Анти-зацикливание: прогресс = изменения в git

- Один и тот же файл не перечитывать **больше 2 раз подряд** без изменений в коде.
- **3+ шага подряд без изменённых файлов** — стоп: сверить `git status`/`git diff`,
  сделать самое маленькое следующее действие; если его нет — спросить пользователя.
- Не повторять одно и то же действие «на всякий случай» в надежде на другой результат.

### 🔄 После рестарта сессии

- **Не читать** старые логи/дампы — их не должно быть.
- **Проверить** `git status` и `git diff` — это актуальное состояние.
- **Продолжать** с места остановки, не «пере-исследуя» уже известное.

### 📦 Коммиты — чекпоинты

- Коммитить **после каждого рабочего шага**.
- Если >30 минут без коммита — остановиться и закоммитить.

### 🔧 Окружение — проверка в начале

- Права, инструменты, контейнеры проверить **один раз в начале** работы.
- В процессе не перепроверять и не тратить шаги на уже проверенное.

### 📝 Код > документация

- Сначала код, потом документация.

## 📋 Как облегчить работу агенту (обязательные)

Канон — корневой [`../AGENT.md`](../AGENT.md) (раздел «Как облегчить работу агенту»)
и [`../IMPROVEMENTS.md`](../IMPROVEMENTS.md); блок синхронизируется по проектам.

### 📦 Архивация выполненного

- Задачи со статусом `✅` из `TASKS.md` **не удаляются, а переносятся в архив**: `../archive/tasks-YYYY-MM-DD.md` (если за день много задач) или `../archive/tasks-<project>-YYYY-MM-DD.md` (проектные реестры).
- **Запрещено** оставлять в `TASKS.md` строки с `✅` дольше конца текущей сессии:
  в реестрах живут только `☐`/`🔄`/`⏸`.
- В рабочих файлах (`AGENT.md`, `TASKS.md`) — только актуальное состояние: без
  устаревших версий, удалённых сервисов, исправленных граблей, выполненных задач.
- Архив в основном контексте **не читается** — только точечный поиск по дате/теме.

### 🧠 Кадзен (1% улучшения)

- После любой задачи — **предложи 1–2 микро-улучшения** в затронутом коде/документации
  (не новая фича: мёртвый комментарий, переименование, `TODO`, фикс опечатки,
  вынос повторяющегося кода в константу).
- **Не делай автоматически** — сначала спроси: «Вижу тут [микро-улучшение]. Сделать?»
  «да» — в рамках того же коммита; «нет» — не возвращаешься.

### ✅ Чеклист конца сессии

- Всё закоммичено во всех затронутых репозиториях (включая `AGENT.md`/`TASKS.md`).
- Незакрытые задачи — в `TASKS.md`; `✅`-строки — в `../archive/`; в рабочих файлах
  не осталось выполненного и устаревшего.
- Версии sprob/sprof (если менялись) — в таблице корневого `../AGENT.md` + запись в `TASKS.md`.
- Мусор инструментов (`.aider*`, `*.history`, дампы диалогов, `mind*.md`) — не в индексе git.

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
| `buildings` | Здания/этажи/входы (используется ТОЛЬКО внутри sprob — `routing/router.go`; проекты не импортируют, см. `TECH_DEBT.md` Т5) |
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
├── modules/           # chats, buildings, schedule, documents, forms, extracts, settings
└── cmd/scripts/       # golangci.sh (lint), update_assister.sh (релиз v1.0.x)
```

## 🚀 make

```bash
make lint    # cmd/scripts/golangci.sh
make update  # релиз новой версии v1.0.x (commit + tag + push)
make test    # go test ./...
```
