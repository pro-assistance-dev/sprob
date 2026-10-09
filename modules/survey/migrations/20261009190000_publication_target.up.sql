-- Полиморфная привязка публикации (формы) к сущности экосистемы:
-- target_type = 'event'|'worker'|'incident'|…, target_id = id сущности.
-- Модуль survey не знает моделей проектов — связь обезличенная (тип + uuid).
alter table publications add column if not exists target_type varchar;
alter table publications add column if not exists target_id uuid;

create index if not exists publications_target_idx on publications (target_type, target_id);
