create table if not exists humans (
  id uuid default uuid_generate_v4() not null primary key,
  name text, 
  surname text, 
  patronymic text,
  date_birth timestamp
);
