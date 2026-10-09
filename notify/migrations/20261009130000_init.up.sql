create table if not exists notify_rules (
    id uuid default uuid_generate_v4() not null primary key,
    name varchar not null,
    event varchar not null default '*.*',
    channel varchar not null default 'email',
    enabled boolean not null default true,
    subject text,
    body text,
    meta jsonb,
    item_order integer default 0,
    created_at timestamp without time zone default current_timestamp not null
);

create table if not exists notify_targets (
    id uuid default uuid_generate_v4() not null primary key,
    rule_id uuid not null references notify_rules(id) on delete cascade,
    type varchar not null,
    value varchar not null,
    item_order integer default 0
);

create table if not exists notify_conds (
    id uuid default uuid_generate_v4() not null primary key,
    rule_id uuid not null references notify_rules(id) on delete cascade,
    field varchar not null,
    operator varchar not null,
    value varchar,
    item_order integer default 0
);

create table if not exists notify_outbox (
    id uuid default uuid_generate_v4() not null primary key,
    channel varchar not null,
    to_address varchar not null,
    subject varchar,
    body text not null,
    meta jsonb,
    rule_id uuid,
    event varchar,
    entity varchar,
    item_id varchar,
    attempts integer not null default 0,
    next_attempt_at timestamp without time zone not null default current_timestamp,
    sent_at timestamp without time zone,
    last_error text,
    created_at timestamp without time zone not null default current_timestamp
);

create index if not exists notify_outbox_pending_idx
    on notify_outbox (next_attempt_at) where sent_at is null;

create table if not exists notify_logs (
    id uuid default uuid_generate_v4() not null primary key,
    rule_id uuid,
    event varchar,
    entity varchar,
    item_id varchar,
    channel varchar,
    to_address varchar,
    subject varchar,
    status varchar,
    error text,
    created_at timestamp without time zone default current_timestamp not null
);
