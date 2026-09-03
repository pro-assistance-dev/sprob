create table if not exists color_themes (
    id uuid default uuid_generate_v4() not null primary key,
    name varchar,
    label varchar
);

insert into public.color_themes (id, "name", label) values
('6fd80c4c-ece3-479c-94b6-005ccebcfe73'::uuid, 'blueLight','Синяя'),
('9f61f302-6821-40b9-94bc-78dedf955a11'::uuid, 'aquaBlueLight', 'Бирюзовая');
