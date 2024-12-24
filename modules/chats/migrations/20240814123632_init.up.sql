create table if not exists chats (
    id uuid default uuid_generate_v4() not null primary key,
    name varchar,
    created_on timestamp default CURRENT_TIMESTAMP  not null,
    chat_messages varchar,
);

create table if not exists chatsmessages (
    id uuid default uuid_generate_v4() not null primary key,
    user_id    uuid references users,
    chat varchar,
    chat_id    uuid references chats,
    message varchar,
    type varchar,
    created_on timestamp default CURRENT_TIMESTAMP  not null,
);

create table if not exists chatsusers (
    id uuid default uuid_generate_v4() not null primary key,
    user_id    uuid references users,
    chat varchar,
    chat_id    uuid references chats,
    join_time timestamp default CURRENT_TIMESTAMP  not null,
    exit_time timestamp default CURRENT_TIMESTAMP  not null,
);
