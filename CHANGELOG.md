# Changelog — sprob

Семантическое версионирование: `v1.0.2xx` — фиксы/мелочи, `v1.1.x` — новые API/модули.
Правила публикации — `scripts/bump.sh`; синхронизация серверов — по AGENT.md.

## v1.3.0 (04.09.2026) — helpers/analytics (общие агрегаты дашбордов)

> А5.3 (rdkb/TASKS.md): вынесено из rdkb map/hr `handlers/analytics` (код был идентичен),
> чтобы дашборды/агрегаты были переиспользуемы во всех проектах больницы.

- `helpers/analytics.Cache` — TTL-кэш агрегатов: `Get/Set/GetOrLoad/Reset`
  (проект держит экземпляр, журналы запрашивает мимо кэша).
- `helpers/analytics.LabelValue` + `SeriesRows(section, items)` / `SeriesValues(items)` —
  серии для ответов дашбордов и плоских выгрузок.
- map/hr переведены на пакет (локальные копии удалены).


## v1.2.0 (04.09.2026) — модуль access (матрица RACI + аудит + JWKS)

> С4.1 (rdkb/TASKS.md): движок access вынесен из `rdkb/map/server/access` в общий модуль
> `modules/access`, чтобы матрица доступа/журналы были во всех проектах больницы.
> Реестр сущностей — ПРОЕКТНЫЙ (`access.Register`), данные FM-системы остались в map.

### Новое: модуль `modules/access`

- `models`: `Role`, `AccessMatrix`, `AuditLog`, `AuthLog` (таблицы `roles`, `access_matrix`,
  `user_roles`, `audit_log`, `auth_log`); миграция модуля — схема `IF NOT EXISTS`
  (`modules/access/migrations`, подключается через `accessM.Init()` в списке миграций).
- Реестр: `Entity`/`FieldInfo` + `Register(list)`/`GetEntity`/`Entities` — движок общий,
  состав охраняемых сущностей задаёт проект.
- `NewMiddleware(h)` / `NewHandler(h, matrix)` — прежний API map: `AccessControl()`
  (enforcement + маскирование ответа), `Audit()` (журнал корректуры), `Matrix()`,
  `RolesFromRequest`/`UserCtx`/`VerifyTokenNoExp` (JWKS keycloak, HS256 fallback).
- Конфиг env: `ACCESS_ENFORCE`, `JWT_VERIFY`, `JWT_VERIFY_FAIL_OPEN`, `JWT_JWKS_URL`,
  `TOKEN_SECRET` (как в map); новые: `ACCESS_ADMIN_ROLE` (default `R00_ADMIN`),
  `ACCESS_APP_CLIENT` (default `map-app`).
- Аддитивно: map продолжает работать без изменений поведения (реестр регистрируется
  при старте); поведенческие тесты middleware перенесены и зелёные.


## v1.1.0 (03.09.2026) — тестируемость: конструкторы, testkit, Init→*Handler

> Из анализа `archive/analysis-sprob-di-2026-09-03.md` / `analysis-sprob-testability-2026-09-03.md`
> (DI-фреймворк не нужен; точечные правки, Т7 TECH_DEBT). Аддитивно; старые вызовы
> (`Init*`, `Init(h)` как statement, `basehandler.SetHelper`) работают как раньше.

### Новые публичные API

- `basehandler.NewR[T](h)` / `NewS[T](h, r)` / `NewH[T](h)` — конструкторы с явным helper;
  legacy `InitR/InitS/InitH/Init` — обёртки над ними (глобал `Helper` — default).
- `routing.WithHelper(h)` — авто-CRUD `InitR[T](api, …)` монтируется без глобала
  `basehandler.Helper` (тесты, изоляция).
- `handlers/*` и `modules/*/handlers/*`: `Init(h)` **возвращает `*Handler`** — цепочка
  handler→service→repository на полях экземпляра, методы не обращаются к пакетным глобалам.
  Глобалы `H/S/R` заполняются для совместимости (удаление — следующий major).
- `testkit` — тестовый харнесс: `NewSQLiteHelper(t, models…)` (sqlite in-memory + bun +
  авто-CREATE TABLE), `Token(h, userID, domainIDs…)` (JWT, секрет `TestSecret`).

### Изоляция состояния

- `middleware`: FTSP-кэш — экземпляр `FTSPStore` в `Middleware` (вместо пакетного глобала);
  удалён мёртвый `middleware/queries.go` (`Query`/`queriesMap`, ссылок не было).
- `metabase`: кэш карточек — поле `Handler.cards`.

### Чистка

- Удалён мёртвый `modules/buildings` (никем не использовался; вызов в роутинге был
  закомментирован) и мусорный `scrap/` (PDF/out.json) — Т5.1/Т5.3.
- Миграции: устранена коллизия версии `20240814123236` (forms/settings; settings →
  `20240814123303`); проверка в `bump.sh` теперь сравнивает числовые версии
  (легальна только пара `up`+`down`) — Т6.2.

### Тесты

- `routing`: авто-CRUD роут на sqlite без `SetHelper` (`WithHelper`);
- `middleware`: изоляция FTSP-кэша между экземплярами;
- `testkit`: создание таблиц, изоляция БД, валидация токена.
