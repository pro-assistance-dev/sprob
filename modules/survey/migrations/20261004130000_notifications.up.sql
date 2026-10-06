-- Уведомления: правила (при каких условиях кому отправлять email) и журнал отправок.
-- FK-constraints не вешаем (общий стиль сервиса: связи — по uuid-колонкам).

create table if not exists notifications (
    id uuid default uuid_generate_v4() not null primary key,
    publication_id uuid,
    name varchar,
    -- список адресов получателей
    emails jsonb,
    subject text,
    body text,
    -- onResponse (пока единственный)
    trigger varchar default 'onResponse',
    enabled boolean default true,
    sort_order integer default 0,
    created_at timestamp default now()
);

create index if not exists notifications_publication_id_idx on notifications (publication_id);

-- Условие отправки: ответ на вопрос (field_code) удовлетворяет operator + value.
create table if not exists notification_rules (
    id uuid default uuid_generate_v4() not null primary key,
    notification_id uuid,
    field_code varchar,
    operator varchar,
    value varchar,
    item_order integer default 0
);

create index if not exists notification_rules_notification_id_idx on notification_rules (notification_id);

-- Журнал отправок: кому, когда, успех/ошибка.
create table if not exists notification_logs (
    id uuid default uuid_generate_v4() not null primary key,
    notification_id uuid,
    publication_id uuid,
    response_id uuid,
    to_address varchar,
    subject varchar,
    status varchar,
    error text,
    created_at timestamp default now()
);

create index if not exists notification_logs_notification_id_idx on notification_logs (notification_id);
create index if not exists notification_logs_publication_id_idx on notification_logs (publication_id);
