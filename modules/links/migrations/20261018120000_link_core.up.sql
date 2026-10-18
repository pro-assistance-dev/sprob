create table if not exists link_domains (
    id uuid default uuid_generate_v4() not null primary key,
    host varchar not null unique,
    scheme varchar not null default 'https',
    is_default bool not null default false,
    note varchar
);

-- Активный домен обёрток (решение 18.10): один, остальные — история.
create unique index if not exists link_domains_single_default
    on link_domains ((is_default)) where is_default;

insert into link_domains (host, scheme, is_default, note) values
    ('links.rdkb-mos.ru', 'https', true, 'обёртка ссылок/QR (rdkb)')
on conflict (host) do nothing;

create table if not exists links (
    id uuid default uuid_generate_v4() not null primary key,
    code varchar not null unique,
    title varchar,
    app varchar not null,
    target_type varchar,
    target_id varchar,
    external_url varchar,
    path varchar,
    enabled bool not null default true,
    redirect_kind int not null default 302,
    expires_at timestamptz,
    note varchar,
    hits bigint not null default 0,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index if not exists links_app_target_idx on links (app, target_type, target_id);
create index if not exists links_target_idx on links (target_type, target_id);

create table if not exists link_hits (
    id uuid default uuid_generate_v4() not null primary key,
    link_id uuid references links (id) on delete cascade,
    code varchar,
    ip varchar,
    user_agent varchar,
    referer varchar,
    created_at timestamptz not null default now()
);

create index if not exists link_hits_link_idx on link_hits (link_id, created_at);
create index if not exists link_hits_created_idx on link_hits (created_at);
