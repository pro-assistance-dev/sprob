-- ФИКС: модель AnswerVariant имеет поле ShowOther (без bun-тега) → колонка
-- `answer_variants.show_other`, но init-миграция создавала только `show_more_questions`,
-- а `other` — только `show_others` (мн. число). Итог: любой запрос формы
-- (GET /api/forms, FTSP, вложенное чтение) падал с
-- `column answer_variants.show_other does not exist` (SQLSTATE=42703).
--
-- Раньше это чинилось локальными миграциями в клиентах (incident:
-- 20260829150000_fix_forms_show_other). Правильное место — модуль форм: теперь
-- колонка появляется у всех потребителей (incident, survey, portal, ...).
-- `show_others` (мн. число) НЕ удаляем — обратная совместимость.
ALTER TABLE answer_variants
    ADD COLUMN IF NOT EXISTS show_other boolean DEFAULT false;
