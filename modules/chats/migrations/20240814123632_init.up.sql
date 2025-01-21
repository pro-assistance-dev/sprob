create table if not exists chats (
    id uuid default uuid_generate_v4() not null primary key,
    name varchar,
    created_on timestamp default CURRENT_TIMESTAMP  not null,
    chat_messages varchar
);

create table if not exists chat_messages (
    id uuid default uuid_generate_v4() not null primary key,
    user_id    uuid,
    chat_id    uuid,
    message varchar,
    type varchar,
    created_on timestamp default CURRENT_TIMESTAMP  not null
);

create table if not exists chat_users (
    id uuid default uuid_generate_v4() not null primary key,
    user_id    uuid,
    chat_id    uuid,
    join_time timestamp default CURRENT_TIMESTAMP  not null,
    exit_time timestamp default CURRENT_TIMESTAMP  not null
);
