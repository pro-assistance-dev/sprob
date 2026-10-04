## v1.5.0 (04.10.2026) — feat(forms): условия показа вопросов (field_rules)

> ЗАЧЕМ. В конструкторе форм (survey) нужны ветвления: «покажи вопрос Б,
> только если на вопрос А ответили X». Раньше дерево формы такого не знало.

- Модель `forms/models.FieldRule`: `field_id` (показываемый вопрос),
  `depends_on_field_id`/`depends_on_field_code` (вопрос-условие), `operator`
  (`equals`/`notEquals`/`contains`/`notContains`/`answered`/`notAnswered`/
  `greater`/`less`), `value`. Несколько правил поля — AND.
- `Field.FieldRules` (has-many) + `SetIDForChildren` проставляет id правилам.
- Сохранение/удаление дерева формы включает `field_rules`
  (`insertChildren`/`deleteChildren`); `GetAll`/`Get` грузят правила.
- Миграция `20261004120000_field_rules`.

## v1.4.9 (03.10.2026) — feat(sql): OR-группы в FTSP-фильтрах

> ЗАЧЕМ. Нужно «поиск одним полем по нескольким колонкам» (ФИО ИЛИ код
> сотрудника). Раньше все фильтры соединялись только через AND — OR не было.

- `filter.FilterModel.Group` (JSON `group`): фильтры с ОДИНАКОВЫМ ненулевым
  `group` объединяются через `OR` (в одной скобке); между группами и остальными
  — `AND`. `group=0` (дефолт) — прежнее поведение.
- Рефактор: `constructWhere` → `whereFragment()` (возвращает `schema.QueryWithArgs`),
  чтобы один фрагмент можно было положить как в `Where`, так и в `WhereGroup`.
- Тест: `TestCreateFilter_OrGroup` (OR внутри группы, AND между).

## v1.4.8 (03.10.2026) — fix(sql): LIKE-фильтр ломался на колонке с таблицей (42P01)

> ЗАЧЕМ. В hr «поиск по ФИО», в portal «поиск по названию новости» отдавали
> 500: `missing FROM-clause entry for table "lower(regexp_replace(news"`
> (SQLSTATE=42P01).

- `filter.FilterModel.constructWhere` собирал выражение
  `lower(regexp_replace(<table>.<col>, ...))` и отдавал его через
  `schema.UnsafeIdent`. Bun видит ТОЧКУ внутри выражения и разбирает его как
  `таблица.колонка` — в SQL уезжал алиас `lower(regexp_replace(news`, Postgres
  отвечал 42P01.
- Теперь колонка передаётся ОТДЕЛЬНЫМ `?` с `bun.Ident`: точка попадает только
  в `Ident`, где bun корректно квотит `"news"."title"`.
- Тест-регрессия: `TestConstructWhere_LikeQuotesDottedColumn` (падает на старом
  коде, проходит на новом).

## v1.4.8 — fix(search): SQL-инъекция через строку поиска

- `handlers/search.Repository.Search` склеивал `searchModel.Query` (ввод
  пользователя) прямо в SQL (`'%' + Query + '%'`) — апостроф ломал запрос, а
  подстановка могла выполнить произвольный SQL. Значения поиска теперь уходят
  ПАРАМЕТРАМИ (`?`), как в остальных фильтрах.

## v1.4.5 (02.10.2026) — refactor: чистка кода (lint 8 → 0)

> ЗАЧЕМ. Изначальная задача «чисти код»: в библиотеке накопились мёртвые пакеты,
> закомментированные легаси-блоки и замечания линтера.

- Удалён мёртвый пакет `helpers/sql/tree/mocks` (584 строки; ссылался только
  закомментированный импорт в тесте).
- Вырезаны закомментированные легаси-блоки: `FilterModel` (join-v1), `treemodel`
  (`parseJSONToTreeModel`), `SearchGroup`, `helpers/project/schema`, `templater`,
  `db/actions`, `middleware.CheckPermission`.
- Убраны неиспользуемые параметры/результаты: `renderField(s, f)` → `renderField(f)`,
  `parseIntDefault(...) (int, bool)` → `int`, тест-хелпер `newJWKSServer(t, kid, …)`
  → `newJWKSServer(t, …)`; промотированные поля `key.PublicKey.N` → `key.N`.
- Устаревший `parser.ParseDir`/`ast.Package` в `codegen` заменён на
  `parser.ParseFile` + группировку файлов по имени пакета (Go 1.25).
  Проверено на rdkb/map: генератор дал 56 TS-классов, diff к закоммиченным — пустой.
- Итог: `golangci-lint run ./...` — **0 issues**; `go vet`/`go test ./...` — зелёные.

## v1.4.4 (02.10.2026) — fix(email): письма-рассылки больше не уходят в спам

> ЗАЧЕМ. На pros часть писем последней рассылки попала в спам. Разбор
> `helpers/email`: письмо собиралось без обязательных для массовой рассылки
> заголовков, а тема с кириллицей шла в заголовке СЫРОЙ.

- Добавлены заголовки `Date`, `Message-ID`, `List-Unsubscribe`,
  `List-Unsubscribe-Post` (one-click). Gmail/Yandex требуют их для bulk-почты:
  без `Date`/`Message-ID` и с одной ссылкой «отписаться» в теле письмо почти
  гарантированно уходит в спам/промо.
- **Тема кодируется RFC 2047** (`mime.QEncoding`) — раньше кириллица в `Subject`
  шла буквально (кракозябры у части клиентов + спам-признак).
- `From` — с отображаемым именем («АНО «Просодействие» <addr>»), а не голый
  punycode-адрес.
- Заголовки собираются в детерминированном порядке (раньше — итерация по map,
  порядок случаен).
- Удалён мёртвый `(*request).ToBytes` (ручная сборка multipart/base64) — не
  использовался нигде; сборка вложений — через `mime/multipart`.
- Тесты: `main_test.go` — наличие анти-спам заголовков, кодирование темы,
  стабильный порядок.

## v1.4.3 (28.09.2026) — fix(http): невалидный/истёкший токен → 401, а не 500

> ЗАЧЕМ. Заведено на pros: в логах `Error #01: Token is expired` /
> `signature is invalid` отдавались как **HTTP 500**, хотя это ошибка клиента.
>
> Причина: `HandleError` сравнивал строку ТОЧНО (`== "Token is expired"`), а jwt-go
> оборачивает ошибку в `*jwt.ValidationError` и добавляет длительность:
> «`Token is expired by 1h2m3s`» — не совпадало НИКОГДА. Так же пролетали
> «`signature is invalid`», «`token is unverifiable`» и др.

- `IsAuthError(err)` — классификация: сначала типизированно (`errors.As` по
  `*jwt.ValidationError` — надёжно, не зависит от текста: пробивает и обёртки),
  затем текстовая страховка по списку фрагментов (регистронезависимо).
- `StatusForError(err)` — 401 для ошибок токена, 500 для остального.
- `HandleError` использует их вместо точного сравнения строки.
- Тесты: `TestIsAuthError_JWTValidation` (13 случаев, включая реальные сообщения
  из логов и НЕ-auth ошибки), `TestStatusForError`.


## v1.4.2 (28.09.2026) — feat(baseR): серверный поиск опций (К2/Т8)

> ЗАЧЕМ. `GET <entity>/options/<label>/<value>` отдаёт справочник ЦЕЛИКОМ: у крупных
> (сотрудники РДКБ — 2250) это секунды и выпадашка на 2000 строк, в которой не найти
> нужное. Клиенты выкручивались самодельными обработчиками (rdkb map — `/workers/search`).

- `GET <entity>/options/<label>/<value>?query=<подстрока>&limit=<n>` — поиск на
  СЕРВЕРЕ. Без `query` поведение прежнее (полный справочник).
- Белый список полей — у МОДЕЛИ: `SearchColumns() []string` (интерфейс `Searchable`).
  `query` НИКОГДА не подставляется в SQL как имя колонки — это исключает инъекцию
  через параметр поиска.
- Лимит обязателен: `SearchLimitDefault` (50) и `SearchLimitMax` (500);
  слишком большой `limit` зажимается, а не выполняется.
- `LOWER(col) LIKE LOWER(?)` вместо `ILIKE`: работает и в Postgres (прод), и в
  sqlite (тесты baseR) — `ILIKE` в sqlite — синтаксическая ошибка.
- У модели без `SearchColumns` запрос с `query` → **400** (не 500): клиент должен
  узнать, что справочник не searchable, а не получить тихо полный список.
- Тесты: `TestRepository_OptionsSearch_ByWhitelist`, `_Limit`, `_NotSearchable`,
  `TestHandler_Options_NotSearchable400` (проверены поломкой белого списка).

## v1.4.1 (20.09.2026) — fix(access): массивы не дают ложный 403

> Найдено на rdkb/map: сохранение помещения/контура у роли-инженера падало с 403
> (`нет права записи на поля: formedFromCodes`). Причина — `fetchOldRow` отдаёт
> `text[]` строкой-литералом PG (`"{}"`), а клиент шлёт JSON (`"[]"`); `stringify`
> давал `"{}" != "[]"` → `fieldChanged=true` → deny, хотя поле не менялось.

- `fieldChanged` сравнивает нормализованные значения: PG-литерал массива → JSON
  (`"{a,b}"` → `["a","b"]`, `"{}"` → `[]`), пустые контейнеры (`nil`/`""`/`[]`/
  `{}`/`null`) — один класс «пусто». Время — как и раньше, по инстансу.
- Тесты: `TestFieldChangedArrays` (6 случаев), `TestNormalizeValue`;
  `TestFieldChanged` (регрессия) зелёный.

## v1.4.0 (14.09.2026) — codegen: общий генератор TS-классов из Go-моделей

> Т8.1 (rdkb/TECH_DEBT.md): генератор был сервисной копией (rdkb/map и rdkb/incident
> держали идентичные `cmd/server/project` + `cmd/generate-ts`). Вынесен в sprob,
> чтобы копий не было во всех проектах экосистемы.

- `codegen` — пакет разбора Go-моделей в схемы (AST → Schema/Field): таблицы,
  json/bun-теги, связи (belongs-to/has-many/m2m), порядок объявления полей.
- `cmd/generate-ts` — CLI: `-service <name>` (заголовок файлов), `-models <path>`,
  `-out <dir>`. Запуск: из каталога сервера сервиса
  `go run github.com/pro-assistance-dev/sprob/cmd/generate-ts -service map -models .`.
- Исправления при выносе (были в сервисных копиях):
  * `NewProject(nil)` паниковал (разыменование конфига без проверки) — теперь nil безопасен;
  * `ModelsPath` игнорировался (всегда обход `.`) — теперь учитывается;
  * убраны отладочные `fmt.Println` (засоряли вывод генератора).
- Тесты: `codegen/project_test.go` (парсинг временного модуля, связи схем,
  порядок полей, идемпотентность, регрессия на nil-конфиг).

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
