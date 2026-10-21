# SPROB — Go-фреймворк для бэкендов (basehandler, baseR, helper, middleware)

Общая библиотека-фреймворк для ВСЕХ бэкендов экосистемы: **rdkb (hr/map/food/leiter/incident), portal, pros, ferma**.
Даёт «скелет» сервера: конфиг (viper), helper-сервисы, авто-CRUD + FTSP, миграции, логирование, JWT, почта, чаты.

- Модуль: `github.com/pro-assistance-dev/sprob` (Go 1.25)
- Версии в проектах: **v1.3.0** в основной массе, **v1.4.0** — rdkb map/incident
  (общий `codegen`); таблица — в корневом `AGENT.md` воркспейса.
  Обновлять `go get`'ом в каждом проекте + запись в TASKS.md проекта + таблица версий.
- Релиз новой версии: `make update` → `cmd/scripts/update_assister.sh`: автоинкремент `v1.0.x` + commit + tag + push
- Минор (`v1.x.0`) — новые API/пакеты (`docs` → CHANGELOG); патч (`v1.0.x`) — фиксы.

---

## 🤖 Правила работы агента (обязательные)

Единые для всех проектов `~/prog/work` — **[`../RULES.md`](../RULES.md)** (канон, обязателен).
Копий в проектах нет: правки — только в `RULES.md`, здесь синхронизировать ничего не нужно.
Шаг 0 — `~/prog/work/status.sh` + `bin/ucheck`; карта «задача → команда» — `bin/README.md`.

## 🔧 helper.Helper — что внутри (helper/helper.go)

| Поле | Назначение |
| --- | --- |
| `DB` | Bun+pgdriver, пул, лог запросов; `Run`/`DoAction` для миграций |
| `HTTP` | gin-обёртка: ListenAndServe, HandleError, JSON |
| `Token` | JWT HS256 (`TOKEN_SECRET`/`_ACCESS_MINUTES`/`_REFRESH_HOURS`) |
| `Broker` | SSE-брокер (`Subscribe`/`SendEvent`) — realtime-обновления |
| `Email` | go-simple-mail (SMTP user/password/server/port/auth) |
| `Uploader` / `PDF` / `Templater` | локальный аплоадер (`UPLOAD_PATH`) / PDF / DOCX-шаблоны (`TEMPLATES_PATH`) |
| `SQL` | построитель + FTSP-инъекция (`InjectFTSP2`) |
| `Validator` / `Cron` | validator/v10 / robfig/cron/v3 (фоновые задачи, напр. datasync hr) |
| `Project` | схемы (`InitSchemas`, `Name`, `Root`) |
| `Social` / `Metabase` / `Logger` | IG/YT/VK-ключи / Metabase-клиент / Logrus+rotatelogs (`loggerhelper/`) |

## 🛡 middleware/ (порядок: `InjectRequestInfo` = claims + FTSP)

- `InjectFTSP` — парсит `{ftsp:{f,s,p,t}}`, кэш по QID, инъекция в SQL (пагинация/фильтры/сортировки).
- `InjectClaims` — `user_id`, `domain_ids` из JWT в контекст (`Claim.FromContext(ctx)`).

## 🗄 models/ — общие сущности (таблицы создаёт сам sprob)

`UserAccount` (auth: login/register/restore), `Human`, `Contact`, `Email`, `Phone`, `Website`,
`PostAddress`, `Address`, `FileInfo`, `ValueType`, `Menu`/`SubMenu`, `FTSPQuery`/`FTSPPreset`,
`SearchElement`/`SearchElementMeta`/`SearchModel`/`SearchResult` (глобальный поиск,
`helpers/search`), `BaseModel`, `WithID`.

## 🧩 modules/ — готовые под-приложения (у каждого своя миграция)

- **`chats`** — чат: `Chat`/`ChatMessage`/`ChatMember`, миграции, WS-роутинг
  (pros: `/ws/connect/:chatId/:userId/...`).
- **`schedule`** — универсальный календарь-расписание (29.08.2026): `ScheduleDay`/`SchedulePlace`/
  `Schedule`/`ScheduleSession`/`ScheduleSlot` (таблицы `schedule_days|places|timetables|sessions|
  slots` — НЕ `schedules`), миграция, авто-CRUD `InitRoutes(api, h)` → `/schedule-days`,
  `/schedule-places`, `/schedule-timetables`, `/schedule-sessions`, `/schedule-slots`;
  слоты — `payload jsonb`. Рассчитан на НОВЫЕ проекты (врачи портала, конференции);
  фронт — модуль `schedule` в sprof. ⚠️ Не подключать туда, где уже есть свои
  `schedules`/`perfoms` (pros) — коллизия роутов.
- **`links`** (18.10.2026) — управляемые ссылки/QR: `Link`/`LinkDomain`/`LinkHit`,
  публичная обёртка `GET /r/:code` (302 на цель), пакетный выпуск `POST /api/link-registry/batch`,
  QR на сервере (`GET /api/link-registry/:id/qr?format=svg|png`), свой энкодер в `modules/links/qr`
  (byte mode, ECC M, проверен реальным декодером — OpenCV, `QR_DECODER=…`).
  Цель собирает **резолвер потребителя** (`links.WithResolver`); домен обёртки — данные (`LinkDomain`).
  ⚠️ Админка живёт на `/api/link-registry`, а НЕ `/api/links`: `baseR` регистрирует `/links`
  (множественное от `Link`) и перехватывает свой путь.
- **`documents`, `forms`, `extracts`, `settings`** — документы, формы, выписки, настройки.

### ⚠️ Bun: писать поля со своими `bun:"col"`-именами — только прямым SQL

Грабля 18.10 (стоила ~2 ч): поле с `bun:"link_code"` (колонка не совпадает с именем поля)
через `NewInsert().Model()` уходило **пустым** — `code` в БД был `''`, ссылка получалась
`//r/`, вторая вставка падала на unique. Не помогли ни `Column(...)`, ни `Value(...)`,
ни `ExcludeColumn`, ни `bun.Safe` (Bun кэширует схему структуры). Работает `db.QueryRowContext`
с явным SQL. Вывод: **критичное поле с переименованной колонкой — проверять запросом после
первой вставки**, а не «на глаз по коду». `?`-плейсхолдеры: драйвер здесь go-pg, `$1` падает
с `there is no parameter $1`.

## 📁 Структура

`config/` (viper: `.env` DB_/TOKEN_/SERVER_/EMAIL_/NAME/ROOT/UPLOAD_PATH/METABASE_/INSTAGRAM_/
YOUTUBE_/VK_ + local.yaml), `helper/` (агрегатор + `Run`), `helpers/` (broker, cron, db, email,
http, logger, metabase, pdf, project, search, social, sql, templater, token, uploader, util,
validator), `middleware/`, `handlers/` (auth, basehandler, contacts, emails, fileinfos,
ftsppresets, humans, menus, metabase, phones, schemas, search, usersaccounts, valuetypes),
`routing/` (`Init`, `InitR`), `models/`, `migrations/`, `modules/`, `testkit/` (sqlite/JWT),
**`codegen/` + `cmd/generate-ts/`** (генератор TS-классов из Go-моделей — общий для
проектов, см. ниже), `cmd/scripts/` (golangci.sh, update_assister.sh), `CHANGELOG.md`.

## 🧬 codegen — генератор TS-классов (v1.4.0+)

Общий кодогенератор: строит TS-классы-«оболочки» из Go-моделей сервиса
(поля по json/bun-тегам, связи, `GetClassName`). Используется проектами через
`go run` из каталога сервера:

```bash
cd <service>/server && go run github.com/pro-assistance-dev/sprob/cmd/generate-ts \
  -service <name> -models . -out ../client/src/classes/generated
```

- `codegen` — пакет разбора Go AST → `Schema`/`Field` (таблицы, связи, порядок полей).
- `cmd/generate-ts` — CLI (`-service` заголовок, `-models` путь, `-out` каталог).
- Проекты держат сгенерированное в git и проверяют синхронность гейтом
  (`make check-ts`; в rdkb — CI-джоб `codegen`). Правка логики — **только в sprob**,
  сервисы обновляют версию.

## 🚀 make

```bash
make lint    # cmd/scripts/golangci.sh
make update  # релиз новой версии v1.0.x (commit + tag + push)
make test    # go test ./...
```
