create table if not exists passports (
    id uuid default uuid_generate_v4() not null,
    name varchar,
  	num varchar,
  	seria varchar,
  	division varchar,
  	division_code varchar,
  	citzenship varchar,
    adress varchar,
    item_date timestamp without time zone
);

create table if not exists passport_scans (
    id uuid default uuid_generate_v4() not null,
    name varchar,
    file_info_id uuid,
    item_order numeric
);
