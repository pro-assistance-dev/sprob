-- In-app «входящие»: уведомления, адресованные пользователю (`user_id`) или
-- группе/роли (`group`), хранятся до прочтения (`read_at`) — для колокольчика/бейджа.
create table if not exists notify_inbox (
    id uuid default uuid_generate_v4() not null primary key,
    user_id varchar,
    "group" varchar,
    title varchar,
    body text,
    url varchar,
    rule_id uuid,
    event varchar,
    entity varchar,
    item_id varchar,
    read_at timestamp without time zone,
    created_at timestamp without time zone not null default current_timestamp
);

create index if not exists notify_inbox_user_idx on notify_inbox (user_id, read_at);
