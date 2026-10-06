-- Invite (приглашения) — персональные ссылки для неанонимных публикаций.
-- FK-constraints не вешаем (общий стиль сервиса: связи — по uuid-колонкам).

create table if not exists invites (
    id uuid default uuid_generate_v4() not null primary key,
    publication_id uuid,
    token varchar unique,
    email varchar,
    phone varchar,
    fio varchar,
    used_at timestamp,
    expires_at timestamp,
    created_at timestamp default now()
);

create index if not exists invites_publication_id_idx on invites (publication_id);
create index if not exists invites_token_idx on invites (token);
