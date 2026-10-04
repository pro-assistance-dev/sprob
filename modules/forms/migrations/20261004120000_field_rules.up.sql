-- Условия показа вопросов («если, то»): field_rules.
--
-- Правило привязано к вопросу (field_id) и описывает условие на ДРУГОЙ вопрос
-- (depends_on_field_id): оператор + значение. Несколько правил поля = AND.
-- Без правил вопрос показывается всегда.

create table if not exists field_rules (
    id uuid default uuid_generate_v4() not null,
    field_id uuid,
    depends_on_field_id uuid,
    depends_on_field_code varchar,
    operator varchar,
    value varchar,
    item_order integer default 0
);

create index if not exists field_rules_field_id_idx on field_rules (field_id);
