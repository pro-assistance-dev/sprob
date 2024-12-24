create table if not exists forms (
    id uuid default uuid_generate_v4() not null,
    name character varying not null,
    item_order integer default 0
);

create table if not exists form_sections (
    id uuid default uuid_generate_v4() not null,
    name character varying not null,
    form_id uuid default uuid_generate_v4(),
    item_order integer default 0
);

create table if not exists fields (
    id uuid default uuid_generate_v4() not null,
    name character varying not null,
    short_name character varying,
    code character varying,

    item_order integer,
    comment character varying,
    required boolean,
    value_type_id uuid,
    form_section_id uuid,
    parent_id uuid
);

create table if not exists answer_variants (
    id uuid default uuid_generate_v4() not null,
    name character varying not null,
    item_order integer,
    field_id uuid,
    score integer,
    is_matrix_y boolean,
    show_more_questions boolean default false

);



create table if not exists form_fills (
    id uuid default uuid_generate_v4() not null,
    form_id uuid not null,
    num character varying,
    filling_percentage int,
    created_at timestamp without time zone default current_timestamp not null
);

create table if not exists field_fills (
    id uuid default uuid_generate_v4() not null,
    value_string character varying,
    value_other character varying,
    value_number numeric,
    value_date date,

    form_fill_id uuid,
    filled boolean default false,

    field_id uuid not null,

    answer_variant_id uuid
);

create table if not exists selected_answer_variants (
    id uuid default uuid_generate_v4() not null,
    answer_variant_id uuid not null,
    answer_variant_id_y uuid,
    field_fill_id uuid
);

create table if not exists value_types (
    id uuid default uuid_generate_v4() not null,
    name character varying
);

insert into public.value_types (id, "name") values
('9f61f302-6821-40b9-94bc-78dedf955a11'::uuid, 'string'),
('6fd80c4c-ece3-479c-94b6-005ccebcfe73'::uuid, 'set'),
('fc00cc5a-f7a5-4974-ad57-9432656d5e0e'::uuid, 'radio'),
('47affcc5-5d32-4b1f-bf07-33382ed06cda'::uuid, 'number'),
('efdd456c-091b-49d9-ac32-d0d345f88e64'::uuid, 'date'),
('9fa59e5f-b5f4-4dd0-821f-b9aa6eb25a10'::uuid, 'file'),
('6fe180c8-c40e-4d7a-8b0a-9d9e22ae9c61'::uuid, 'text'),
('841bb273-e78c-409c-b442-598e3de6a2b3'::uuid, 'matrixRadio'),
('841bb273-e78c-409c-b442-598e3de6a2b4'::uuid, 'matrixSet');


