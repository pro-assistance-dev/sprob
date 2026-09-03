# TECH_DEBT.md — план работ по техническому долгу (sprob)

> Отдельный план на основе аудита кода (обновлён 03.09.2026), по образцу `portal/TECH_DEBT.md`.
> sprob — Go-библиотека (`github.com/pro-assistance-dev/sprob`), используется
> в rdkb (5 сервисов), portal, pros, ferma. Долг здесь = долг во всех проектах сразу.
> Статусы: `☐` todo · `🔄` в работе · `✅` done.
> Приоритеты: 🔴 Критично · 🟠 Высокий · 🟡 Средний · ⚪ Низкий.
> Обновлять после каждой сессии.

---

## ✅ Т1. `GetForm` — принимать JSON в API — ЗАКРЫТО 31.08 (v1.0.250)

Детект по Content-Type: `application/json` → парсить тело напрямую,
`multipart/form-data` → текущий путь; возврат `map[string][]*multipart.FileHeader`
сохранён; тесты на оба пути (3 теста, PASS). Опубликован в **v1.0.250**
(вместе с Н11 kebab→camel) и синхронизирован во все 7 серверов.

## ✅ Т2. Включить линтеры (golangci) — ЗАКРЫТО 01.09

`.golangci.yaml` переведён в v2-формат, включён базовый набор: errcheck, govet,
ineffassign, misspell, staticcheck, unconvert, unparam, gocritic. Починено 24+ замечания
(Close() без проверки, exitAfterDefer ×2, singleCaseSwitch ×4, ifElseChain ×2, captLocal ×4,
badCall Join, QF1012, пропущенная проверка ошибки в buildings Update, комментарии).
`commentFormatting` отключён с объяснением — ~230 строк комментированного кода в стиле
`//code` (вернуть при чистке мёртвого кода, Т5). Итог: **golangci-lint run ./... = 0 issues**.
Осталось (Т2.3): подключить `golangci-lint run` в CI rdkb.

## ✅ Т3. TODO «Стас посмотри, плз» и `context.TODO()` — ЗАКРЫТО 01.09

TODO в buildings — это пропущенная проверка ошибки при удалении входов (в отличие от
этажей): добавлен `if err != nil { return err }`. `context.TODO()` ×4 (helpers/db/actions.go,
helper/helper.go) → `context.Background()` — верхнеуровневые CLI/bootstrap-операции,
контекстной цепочки нет.

## ✅ Т4. Тестовое покрытие библиотеки — ЗАКРЫТО 01.09

- GetForm: тесты на оба пути (в рамках Т1, 31.08).
- **baseR**: CRUD-цикл на sqlite in-memory (modernc.org/sqlite + sqlitedialect,
  `handlers/basehandler/repository_test.go`): create → get → options → getAll → update →
  delete + Get unknown → ErrNoRows (01.09).
- schedule: раскладка слотов/конфликты живут в sprof (`src/modules/schedule/`,
  тесты 27/27 из portal Н35-6); в Go-моделях чистой логики нет — тестировать нечего.
- Отчёт покрытия (`go test -cover`) — остаётся на будущее (низкий приоритет).

## 🟡 Т5. Мёртвый/неиспользуемый код

- [ ] 1. `modules/buildings` — используется ТОЛЬКО внутри sprob
     (`routing/router.go`), ни один проект не импортирует. Либо удалить,
     либо вынести в отдельный модуль/проект
- [ ] 2. Проверить `modules/extracts`, `modules/documents`, `modules/settings` —
     кто реально использует (grep по проектам)
- [ ] 3. `scrap/` — актуален ли (используется portal для скрейпа страниц?)

## 🟡 Т6. Процесс публикации и версии

**Сделано**: `scripts/bump.sh` — чеклист публикации (31.08, О5.4).

- [ ] 1. Публиковать тег с описанием изменений (семантическая версия +
     CHANGELOG)
- [ ] 2. Миграции модулей: до публикации проверять уникальность таймстампов
     (в `modules/*/migrations`)

## ✅ Т6.5. ⚠️ Огромный незакоммиченный WIP — РАЗОБРАН 31.08 (О3)

327 файлов = случайный chmod (644→775) — восстановлено; search-рефакторинг
(global_search без миграции) — откачен, сохранён в /tmp/sprob-search-wip.patch;
Н11 (GetSchema kebab→camel) — закоммичен; репозиторий чист.

---

## Прогресс

| Дата | Что сделано |
|------|-------------|
| 31.08.2026 | Создан план; Т1 (GetForm: JSON + multipart, 3 теста — PASS), v1.0.250 опубликован и синхронизирован в 7 серверов; обнаружен чужой незакоммиченный WIP (~200 файлов) — вынесен в Т6.5; Т6.1: scripts/bump.sh |
| 01.09.2026 | Т2 (golangci: базовый набор включён, 0 issues; commentFormatting отключён — комментированный код), Т3 (TODO Стас разобран — пропущенная проверка ошибки; context.TODO→Background), Т4 (baseR CRUD-тест на sqlite; schedule — тесты в sprof 27/27) |
| 03.09.2026 | Сверка реестра с git: статусы/прогресс актуальны (v1.0.251 — последний релиз, 01.09); изменений после 01.09 нет |
