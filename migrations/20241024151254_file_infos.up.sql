create table if not exists file_infos (
  id uuid default uuid_generate_v4() not null primary key,
  original_name text, 
  file_system_path text
);
