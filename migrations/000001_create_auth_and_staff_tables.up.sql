-- ============================================================================
-- Migration: 000001_create_auth_and_staff_tables.up.sql
-- Module:    Auth + Staff
-- Created:   2026-07-02
-- Run with:  migrate -path ./migrations -database "$DATABASE_URL" up
-- ============================================================================

-- ── Enum types ───────────────────────────────────────────────────────────────
-- All enums are defined once here; subsequent migrations reference them.

CREATE TYPE staff_role_enum          AS ENUM ('admin', 'staff');
CREATE TYPE account_status_enum      AS ENUM ('aktif', 'nonaktif');
CREATE TYPE owner_type_enum          AS ENUM ('staff', 'patient');

-- ── Staff accounts ───────────────────────────────────────────────────────────
CREATE TABLE staff_accounts (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name          VARCHAR(150) NOT NULL,
    username           VARCHAR(50)  UNIQUE NOT NULL,
    email              VARCHAR(150) UNIQUE NOT NULL,
    password_hash      VARCHAR(255) NOT NULL,
    whatsapp_number    VARCHAR(20),
    role               staff_role_enum NOT NULL,
    status             account_status_enum NOT NULL DEFAULT 'aktif',
    position_title     VARCHAR(100),
    short_bio          TEXT,
    profile_photo_url  TEXT,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at         TIMESTAMPTZ
);

CREATE INDEX idx_staff_role   ON staff_accounts(role);
CREATE INDEX idx_staff_status ON staff_accounts(status);
CREATE INDEX idx_staff_deleted ON staff_accounts(deleted_at) WHERE deleted_at IS NOT NULL;

-- ── Password reset tokens (polymorphic — shared by staff & patients) ─────────
CREATE TABLE password_reset_tokens (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_type  owner_type_enum NOT NULL,
    owner_id    UUID        NOT NULL,
    email       VARCHAR(150) NOT NULL,
    otp_code    VARCHAR(10)  NOT NULL,
    is_used     BOOLEAN      NOT NULL DEFAULT false,
    expires_at  TIMESTAMPTZ  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_prt_email_otp ON password_reset_tokens(email, otp_code) WHERE is_used = false;

-- ── Auth sessions (polymorphic — shared by staff & patients) ─────────────────
CREATE TABLE auth_sessions (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_type      owner_type_enum NOT NULL,
    owner_id        UUID         NOT NULL,
    device_info     VARCHAR(255),
    refresh_token   VARCHAR(500) NOT NULL UNIQUE,
    expires_at      TIMESTAMPTZ  NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_owner ON auth_sessions(owner_type, owner_id);
CREATE INDEX idx_sessions_token ON auth_sessions(refresh_token);

-- Seed initial admin and staff accounts (password is 'password123')
INSERT INTO staff_accounts (id, full_name, username, email, password_hash, role, status)
VALUES 
    ('e6c464c2-5509-4c17-a068-07e8ef6e61f2', 'System Administrator', 'admin', 'admin@dsmes.com', '$2a$10$w850aK9B74hC4s/5Z1yOSeKpeKkI2wL3121X77.i.J7cO7N0l/Jb.', 'admin', 'aktif'),
    ('b5220c4c-70e6-42d7-a5eb-0b5c15668e1a', 'Monitoring Staff', 'staff', 'staff@dsmes.com', '$2a$10$w850aK9B74hC4s/5Z1yOSeKpeKkI2wL3121X77.i.J7cO7N0l/Jb.', 'staff', 'aktif')
ON CONFLICT DO NOTHING;

