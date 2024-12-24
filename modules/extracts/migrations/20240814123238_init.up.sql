create table if not exists extracts (
    id uuid default uuid_generate_v4() not null,
    name varchar,
    schema_id uuid
);

