-- Темы оформления формы (показываются при заполнении) + привязка публикации к теме.
-- Токены — карта CSS-переменных (jsonb); рантайм навешивает их на корень опроса.

create table if not exists themes (
    id uuid default uuid_generate_v4() not null primary key,
    name varchar,
    slug varchar unique,
    tokens jsonb,
    dark boolean default false,
    sort_order integer default 0,
    created_at timestamp default now()
);

alter table publications add column if not exists theme_id uuid;
create index if not exists publications_theme_id_idx on publications (theme_id);

-- Несколько готовых тем (идемпотентно по slug).
insert into themes (id, name, slug, tokens, dark, sort_order)
select 'b0000000-0000-4000-8000-000000000001', 'Классика', 'classic',
       '{"--brand-surface":"#ffffff","--brand-surface-sub":"#f5f6f8","--font-color":"#1a1a1a","--button-primary-background":"#476db5","--button-primary-color":"#ffffff","--border-radius-default":"8px"}'::jsonb,
       false, 0
where not exists (select 1 from themes where slug = 'classic');

insert into themes (id, name, slug, tokens, dark, sort_order)
select 'b0000000-0000-4000-8000-000000000002', 'Медицина', 'medical',
       '{"--brand-surface":"#ffffff","--brand-surface-sub":"#eef6f3","--font-color":"#16302a","--button-primary-background":"#2f9e7a","--button-primary-color":"#ffffff","--border-radius-default":"10px"}'::jsonb,
       false, 1
where not exists (select 1 from themes where slug = 'medical');

insert into themes (id, name, slug, tokens, dark, sort_order)
select 'b0000000-0000-4000-8000-000000000003', 'Тёмная', 'dark',
       '{"--brand-surface":"#1e1f22","--brand-surface-sub":"#2a2c30","--font-color":"#f2f3f5","--button-primary-background":"#6f8ff5","--button-primary-color":"#0f1012","--border-radius-default":"10px"}'::jsonb,
       true, 2
where not exists (select 1 from themes where slug = 'dark');

insert into themes (id, name, slug, tokens, dark, sort_order)
select 'b0000000-0000-4000-8000-000000000004', 'Тёплая', 'warm',
       '{"--brand-surface":"#fffdf8","--brand-surface-sub":"#fdf1e3","--font-color":"#3a2a17","--button-primary-background":"#e08b3c","--button-primary-color":"#ffffff","--border-radius-default":"14px"}'::jsonb,
       false, 3
where not exists (select 1 from themes where slug = 'warm');
