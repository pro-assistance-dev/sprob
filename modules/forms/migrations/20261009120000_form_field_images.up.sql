-- Изображения в опросе: шапка формы и картинка к каждому вопросу.
-- Храним путь, отданный файловым сервисом (/api/static/<fileSystemPath>) —
-- сами файлы живут в file_infos + статике, здесь только ссылка.
alter table forms add column if not exists image_url varchar;
alter table fields add column if not exists image_url varchar;
