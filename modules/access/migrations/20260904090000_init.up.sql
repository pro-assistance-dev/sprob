-- Модуль access (sprob): матрица доступа (RACI), роли, журнал корректуры (audit_log),
-- журнал входа-выхода (auth_log), назначение ролей (user_roles).
-- Схема повторяет таблицы rdkb/map (вынесены 04.09, С4.1): у map таблицы уже есть —
-- здесь CREATE ... IF NOT EXISTS, чтобы новые проекты получали схему автоматически.
-- ⚠️ Наполнение (роли проекта, access_matrix, сиды) — ПРОЕКТНОЕ, в модуль не входит.

CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code varchar NOT NULL UNIQUE,
    name varchar NOT NULL,
    description text NULL,
    periodicity text NULL,
    consequences text NULL,
    responsible text NULL,
    status varchar NOT NULL DEFAULT 'действует',
    disbanded_at timestamptz NULL
);

--bun:split

CREATE TABLE IF NOT EXISTS access_matrix (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity varchar NOT NULL,
    field varchar NOT NULL,
    role_code varchar NOT NULL,
    access varchar NOT NULL DEFAULT '',
    owner_role varchar NULL,
    periodicity varchar NULL,
    consequence varchar NULL,
    UNIQUE (entity, field, role_code)
);

CREATE INDEX IF NOT EXISTS access_matrix_entity_idx ON access_matrix (entity, field);

--bun:split

CREATE TABLE IF NOT EXISTS user_roles (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL,
    role_code varchar NOT NULL,
    UNIQUE (user_id, role_code)
);

CREATE INDEX IF NOT EXISTS user_roles_user_idx ON user_roles (user_id);

--bun:split

CREATE TABLE IF NOT EXISTS audit_log (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at timestamptz NOT NULL DEFAULT now(),
    user_id uuid NULL,
    user_name varchar NULL,
    role_code varchar NULL,
    entity varchar NULL,
    object_id uuid NULL,
    code varchar NULL,
    field varchar NULL,
    old_value text NULL,
    new_value text NULL,
    operation varchar NULL
);

CREATE INDEX IF NOT EXISTS audit_log_entity_idx ON audit_log (entity, created_at DESC);
CREATE INDEX IF NOT EXISTS audit_log_user_idx ON audit_log (user_id, created_at DESC);

--bun:split

CREATE TABLE IF NOT EXISTS auth_log (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at timestamptz NOT NULL DEFAULT now(),
    user_id uuid NULL,
    user_name varchar NULL,
    role_code varchar NULL,
    operation varchar NOT NULL,
    reason varchar NULL,
    ip varchar NULL,
    user_agent varchar NULL
);

CREATE INDEX IF NOT EXISTS auth_log_created_idx ON auth_log (created_at DESC);
CREATE INDEX IF NOT EXISTS auth_log_user_idx ON auth_log (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS auth_log_operation_idx ON auth_log (operation, created_at DESC);
