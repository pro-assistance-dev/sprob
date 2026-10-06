-- Survey (опросник): публикация форм, параметризация анкеты, учёт ответов.
--
-- Доменная модель самих форм (forms/form_sections/fields/answer_variants/
-- form_fills/field_fills) живёт в sprob/modules/forms — здесь НЕ дублируем.
-- Эти таблицы добавляют то, чего нет в шаблоне формы: публикацию наружу,
-- параметры из URL (одна анкета на разные отделения) и метаданные ответов.
--
-- FK-constraints не вешаем (общий стиль сервиса: связи — по uuid-колонкам).

create table if not exists publications (
    id uuid default uuid_generate_v4() not null primary key,
    form_id uuid,
    title varchar,
    slug varchar unique,
    -- draft | published | closed
    status varchar default 'draft',
    -- анонимная публикация: без приглашения и без связи с респондентом
    anonym boolean default false,
    opens_at timestamp,
    closes_at timestamp,
    -- 0 — без лимита ответов
    "limit" integer default 0,
    thank_you_text text,
    -- бренд публичного рантайма: rdkb | portal | pros | ferma (встраивание без форка)
    brandbook varchar default 'rdkb',
    created_at timestamp default now(),
    updated_at timestamp default now()
);

create index if not exists publications_slug_idx on publications (slug);
create index if not exists publications_form_id_idx on publications (form_id);

-- SurveyParam — параметр анкеты из URL (`/f/<slug>?department=cardio`).
-- Ключ параметризации: одна форма на разные отделения/подразделения; значение
-- пишется в метаданные ответа, может предзаполнять поле формы (field_code).
create table if not exists survey_params (
    id uuid default uuid_generate_v4() not null primary key,
    publication_id uuid,
    key varchar,
    label varchar,
    -- код поля формы (fields.code), которое предзаполняется (пусто — только метаданные)
    field_code varchar,
    required boolean default false,
    -- не показывать предзаполненное поле респонденту
    hidden boolean default false,
    -- допустимые значения (пусто — любое)
    options jsonb,
    sort_order integer default 0
);

create index if not exists survey_params_publication_id_idx on survey_params (publication_id);

-- Response — ответ на публикацию: обёртка над form_fills + метаданные аналитики.
-- Ответы по полям лежат в field_fills (sprob/modules/forms) — НЕ дублируем.
create table if not exists responses (
    id uuid default uuid_generate_v4() not null primary key,
    publication_id uuid,
    form_fill_id uuid,
    -- значения параметров анкеты из URL (key → value) для выгрузок/фильтров
    params jsonb,
    -- необратимые хэши для антифрода без хранения PII
    respondent_hash varchar,
    ip_hash varchar,
    started_at timestamp default now(),
    finished_at timestamp,
    duration integer default 0,
    created_at timestamp default now()
);

create index if not exists responses_publication_id_idx on responses (publication_id);
create index if not exists responses_form_fill_id_idx on responses (form_fill_id);
create index if not exists responses_created_at_idx on responses (created_at);
