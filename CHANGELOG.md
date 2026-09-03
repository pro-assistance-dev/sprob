# Changelog — sprob

Семантическое версионирование: `v1.0.2xx` — фиксы/мелочи, `v1.1.x` — новые API/модули.
Правила публикации — `scripts/bump.sh`; синхронизация серверов — по AGENT.md.

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
