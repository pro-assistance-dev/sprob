-- Фикс FK (29.08.2026): в первой версии миграции 20260829000000 FK у schedule_sessions
-- и schedule_slots ссылались на таблицу `schedules` (legacy portal/pros) вместо
-- `schedule_timetables`. Для БД, где миграция уже применена, пересоздаём FK
-- идемпотентно; для свежих БД (с исправленной 20260829000000) — drop+add безвредны.

alter table schedule_slots drop constraint if exists schedule_slots_schedule_id_fkey;
alter table schedule_slots
    add constraint schedule_slots_schedule_id_fkey
    foreign key (schedule_id) references schedule_timetables (id) on update cascade on delete cascade;

alter table schedule_sessions drop constraint if exists schedule_sessions_schedule_id_fkey;
alter table schedule_sessions
    add constraint schedule_sessions_schedule_id_fkey
    foreign key (schedule_id) references schedule_timetables (id) on update cascade on delete cascade;
