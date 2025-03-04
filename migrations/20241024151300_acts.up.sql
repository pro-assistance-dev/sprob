create table acts (
  id uuid default uuid_generate_v4() not null primary key,
  user_id uuid,
  client_id uuid,
  item_id uuid,
  created_at timestamp,
  updated_at timestamp
)
