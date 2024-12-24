create table if not exists snilss (
    id uuid default uuid_generate_v4() not null,
    num varchar,
    file_info_id uuid
);
