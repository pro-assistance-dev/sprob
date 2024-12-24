create table if not exists buildings (
    id uuid default uuid_generate_v4() not null primary key,
    name varchar,
    address varchar,
    number varchar,
    map_node_name varchar
);

create table if not exists entrances (
    id uuid default uuid_generate_v4() not null primary key,
    name varchar,
    number varchar,
    map_node_name varchar,
    building_id uuid REFERENCES buildings(id) ON UPDATE CASCADE ON DELETE CASCADE
);

create table if not exists floors (
    id uuid default uuid_generate_v4() not null primary key,
    name varchar,
    number varchar,
    building_id uuid REFERENCES buildings(id) ON UPDATE CASCADE ON DELETE CASCADE
);
