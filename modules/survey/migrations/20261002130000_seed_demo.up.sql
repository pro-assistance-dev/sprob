-- Seed: демо-форма «Удовлетворённость пациентов» + публикация + параметр анкеты.
-- Для быстрого клик-теста админки/рантайма на чистой БД (dev).
--
-- Публичная ссылка: /f/satisfaction?department=cardio
-- (department — обязательный параметр с допустимыми значениями).
--
-- ⚠️ Таблицы модуля форм (forms/form_sections/fields/answer_variants) объявлены
-- БЕЗ primary key/unique — поэтому `ON CONFLICT (id)` на них падает. Идемпотентность
-- делаем через `where not exists`.

insert into forms (id, name, item_order)
select 'a0000000-0000-4000-8000-000000000001', 'Удовлетворённость пациентов', 0
where not exists (select 1 from forms where id = 'a0000000-0000-4000-8000-000000000001');

insert into form_sections (id, name, form_id, item_order)
select 'a0000000-0000-4000-8000-000000000002', 'Общее', 'a0000000-0000-4000-8000-000000000001', 0
where not exists (select 1 from form_sections where id = 'a0000000-0000-4000-8000-000000000002');

-- Вопрос «Отделение» (radio): value_type_id — из сида модуля форм.
insert into fields (id, name, short_name, code, item_order, required, value_type_id, form_section_id)
select 'a0000000-0000-4000-8000-000000000003', 'Отделение', 'Отделение', 'department', 0, true,
       'fc00cc5a-f7a5-4974-ad57-9432656d5e0e', 'a0000000-0000-4000-8000-000000000002'
where not exists (select 1 from fields where id = 'a0000000-0000-4000-8000-000000000003');

insert into answer_variants (id, name, item_order, field_id, score)
select 'a0000000-0000-4000-8000-000000000004', 'Кардиология', 0, 'a0000000-0000-4000-8000-000000000003', 0
where not exists (select 1 from answer_variants where id = 'a0000000-0000-4000-8000-000000000004');

insert into answer_variants (id, name, item_order, field_id, score)
select 'a0000000-0000-4000-8000-000000000005', 'Хирургия', 1, 'a0000000-0000-4000-8000-000000000003', 0
where not exists (select 1 from answer_variants where id = 'a0000000-0000-4000-8000-000000000005');

-- Вопрос «Оценка» (radio).
insert into fields (id, name, short_name, code, item_order, required, value_type_id, form_section_id)
select 'a0000000-0000-4000-8000-000000000006', 'Оцените качество обслуживания', 'Оценка', 'rate', 1, true,
       'fc00cc5a-f7a5-4974-ad57-9432656d5e0e', 'a0000000-0000-4000-8000-000000000002'
where not exists (select 1 from fields where id = 'a0000000-0000-4000-8000-000000000006');

insert into answer_variants (id, name, item_order, field_id, score)
select 'a0000000-0000-4000-8000-000000000007', 'Отлично', 0, 'a0000000-0000-4000-8000-000000000006', 5
where not exists (select 1 from answer_variants where id = 'a0000000-0000-4000-8000-000000000007');

insert into answer_variants (id, name, item_order, field_id, score)
select 'a0000000-0000-4000-8000-000000000008', 'Хорошо', 1, 'a0000000-0000-4000-8000-000000000006', 4
where not exists (select 1 from answer_variants where id = 'a0000000-0000-4000-8000-000000000008');

insert into answer_variants (id, name, item_order, field_id, score)
select 'a0000000-0000-4000-8000-000000000009', 'Плохо', 2, 'a0000000-0000-4000-8000-000000000006', 1
where not exists (select 1 from answer_variants where id = 'a0000000-0000-4000-8000-000000000009');

insert into publications (id, form_id, title, slug, status, anonym, thank_you_text, brandbook)
select 'a0000000-0000-4000-8000-000000000010', 'a0000000-0000-4000-8000-000000000001',
       'Удовлетворённость пациентов', 'satisfaction', 'published', true,
       'Спасибо! Ваш ответ помогает нам стать лучше.', 'rdkb'
where not exists (select 1 from publications where id = 'a0000000-0000-4000-8000-000000000010');

insert into survey_params (id, publication_id, key, label, required, options, sort_order)
select 'a0000000-0000-4000-8000-000000000011', 'a0000000-0000-4000-8000-000000000010',
       'department', 'Отделение', true, '["cardio","surgery"]', 0
where not exists (select 1 from survey_params where id = 'a0000000-0000-4000-8000-000000000011');
