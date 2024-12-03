create table if not exists users_accounts (
  id uuid default uuid_generate_v4() not null primary key,
  uuid uuid default uuid_generate_v4() not null, 
  email text, 
  login text, 
  phone text,
  password text
);
