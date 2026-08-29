-- Модуль schedule (sprob): универсальный календарь-расписание.
-- Дни → расписания (место × день, таблица schedule_timetables) → секции/слоты.
-- ⚠️ Не называется `schedules`: у portal/pros уже есть свои таблицы `schedules`.

create table if not exists schedule_days (
    id uuid default uuid_generate_v4() not null primary key,
    item_date date not null,
    description varchar,
    owner_id uuid not null,
    owner_type varchar not null,
    created_at timestamp default CURRENT_TIMESTAMP not null,
    updated_at timestamp default CURRENT_TIMESTAMP not null
);

create index if not exists schedule_days_owner_idx on schedule_days (owner_id, owner_type);

create table if not exists schedule_places (
    id uuid default uuid_generate_v4() not null primary key,
    name varchar not null,
    description varchar,
    color varchar,
    created_at timestamp default CURRENT_TIMESTAMP not null,
    updated_at timestamp default CURRENT_TIMESTAMP not null
);

create table if not exists schedule_timetables (
    id uuid default uuid_generate_v4() not null primary key,
    name varchar,
    description varchar,
    schedule_day_id uuid not null references schedule_days on update cascade on delete cascade,
    place_id uuid not null references schedule_places on update cascade on delete cascade,
    created_at timestamp default CURRENT_TIMESTAMP not null,
    updated_at timestamp default CURRENT_TIMESTAMP not null
);

create index if not exists schedule_timetables_day_idx on schedule_timetables (schedule_day_id);
create index if not exists schedule_timetables_place_idx on schedule_timetables (place_id);

create table if not exists schedule_sessions (
    id uuid default uuid_generate_v4() not null primary key,
    name varchar,
    description varchar,
    start_time varchar not null,
    end_time varchar not null,
    schedule_id uuid not null references schedule_timetables on update cascade on delete cascade,
    created_at timestamp default CURRENT_TIMESTAMP not null,
    updated_at timestamp default CURRENT_TIMESTAMP not null
);

create index if not exists schedule_sessions_schedule_idx on schedule_sessions (schedule_id);

create table if not exists schedule_slots (
    id uuid default uuid_generate_v4() not null primary key,
    name varchar,
    description varchar,
    start_time varchar not null,
    end_time varchar not null,
    schedule_id uuid not null references schedule_timetables on update cascade on delete cascade,
    session_id uuid references schedule_sessions on update cascade on delete set null,
    payload jsonb,
    created_at timestamp default CURRENT_TIMESTAMP not null,
    updated_at timestamp default CURRENT_TIMESTAMP not null
);

create index if not exists schedule_slots_schedule_idx on schedule_slots (schedule_id);
create index if not exists schedule_slots_session_idx on schedule_slots (session_id);
