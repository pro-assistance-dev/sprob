create table if not exists passports (
    id uuid default uuid_generate_v4() not null,
    name varchar,
    schema_id uuid,

  	num varchar NULL,
  	seria varchar NULL,
  	division varchar NULL,
  	division_code varchar NULL,
  	citzenship varchar NULL,
    item_date timestamp without time zone
);

create table if not exists passport_scans (
    id uuid default uuid_generate_v4() not null,
    name varchar,
    file_info_id uuid,
    item_order number
);
