# TECH_DEBT.md — план работ по техническому долгу (sprob)

> Отдельный план на основе аудита кода (обновлён 03.09.2026), по образцу `portal/TECH_DEBT.md`.
> sprob — Go-библиотека (`github.com/pro-assistance-dev/sprob`), используется
> в rdkb (5 сервисов), portal, pros, ferma. Долг здесь = долг во всех проектах сразу.
> Статусы: `☐` todo · `🔄` в работе. Приоритеты: 🔴 Критично · 🟠 Высокий · 🟡 Средний · ⚪ Низкий.
> Закрытые пункты (Т1–Т4, Т5.1, Т5.3, Т6.2, Т6.5, Т7.1, Т7.2) и история реестра — в
> [`../archive/tech-debt-sprob-2026-09-03.md`](../archive/tech-debt-sprob-2026-09-03.md).

---

## 🟡 Т5. Мёртвый/неиспользуемый код

- [ ] 1. `modules/extracts`, `modules/documents`, `modules/settings` — кто реально использует.
     **Проверка 03.09**: ни один проект не импортирует (импорты только внутри sprob,
     `routing/router.go`); во фронтах вызовов роутов нет. Модули инертно регистрируются через
     `routing.Init` во всех проектах. Удаление — отдельное решение (роут-контракт, версия).
     `modules/buildings` (мёртвый, никем не использовался) и `scrap/` (мусор: PDF/out.json,
     ссылок нет) — **удалены 03.09** → архив

## 🟡 Т6. Процесс публикации и версии

- [ ] 1. Публиковать тег с описанием изменений (семантическая версия +
     CHANGELOG)

## 🟠 Т7. Тестируемость: убрать глобалы-мосты (план 03.09)

> Из анализа [`../archive/analysis-sprob-di-2026-09-03.md`](../archive/analysis-sprob-di-2026-09-03.md)
> и [`../archive/analysis-sprob-testability-2026-09-03.md`](../archive/analysis-sprob-testability-2026-09-03.md):
> DI-фреймворк не нужен; точечные правки. Каждый шаг аддитивен.

- [x] 1. **basehandler/routing**: конструкторы `NewR[T](h)`/`NewS[T](h, r)`/`NewH[T](h)`
  - опция `routing.WithHelper(h)` — авто-CRUD монтируется без `basehandler.Helper`
    (глобал остаётся default-ом для legacy `Init*`)
- [x] 2. **middleware**: пакетные кэши `ftspStore`/`queriesMap` — в поля `Middleware`
     (изоляция FTSP-состояния на тест); `queries.go` — мёртвый, удалить
- [ ] 3. **sprob/testkit** (по образцу `basehandler/repository_test.go`): `NewSQLiteHelper(t)`,
     роут-харнесс, тестовый JWT — потом раскатка на клиенты (корневой О9)
- [ ] 4. **handlers/\* sprob**: `Init(h)` возвращает `*Handler` (прецедент — portal auditlog),
     глобалы `H/S/R` — на вылет в 2 этапа (2-й — со следующей major-версией)
- [ ] 5. **Сервисы клиентов**: repository на границе service→repo как интерфейс/поле,
     доменная логика — чистыми методами (канон для нового кода; задачи по сервисам —
     в TECH_DEBT проектов)

> Т7.1/Т7.2 выполнены 03.09 (коммит b695950, тесты: авто-CRUD роут на sqlite без
> SetHelper — routing/initR_test.go; изоляция FTSP-кэша — middleware/ftsp_test.go) → архив

> CI-остаток golangci (Т2.3) ведётся в `rdkb/TECH_DEBT.md` Т8.2.
