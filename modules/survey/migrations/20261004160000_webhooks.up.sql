-- Вебхуки после ответа: внешняя интеграция (POST JSON + HMAC-подпись).
-- Условия вызова — те же операторы, что у уведомлений (webhook_rules).

create table if not exists webhooks (
    id uuid default uuid_generate_v4() not null primary key,
    publication_id uuid,
    name varchar,
    url varchar,
    -- HMAC-ключ подписи тела (X-Signature: sha256=<hex>)
    secret varchar,
    enabled boolean default true,
    sort_order integer default 0,
    created_at timestamp default now()
);

create index if not exists webhooks_publication_id_idx on webhooks (publication_id);

create table if not exists webhook_rules (
    id uuid default uuid_generate_v4() not null primary key,
    webhook_id uuid,
    field_code varchar,
    operator varchar,
    value varchar,
    item_order integer default 0
);

create index if not exists webhook_rules_webhook_id_idx on webhook_rules (webhook_id);

create table if not exists webhook_logs (
    id uuid default uuid_generate_v4() not null primary key,
    webhook_id uuid,
    publication_id uuid,
    response_id uuid,
    url varchar,
    status_code integer default 0,
    status varchar,
    error text,
    created_at timestamp default now()
);

create index if not exists webhook_logs_publication_id_idx on webhook_logs (publication_id);
