-- ============================================================================
-- Migration: 000023_create_health_facilities.up.sql
-- Module:    Health Facility (Puskesmas) master data
-- Created:   2026-08-29
-- ============================================================================

-- ── Health facilities master table ───────────────────────────────────────────
CREATE TABLE IF NOT EXISTS health_facilities (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(150) NOT NULL UNIQUE,
    address     VARCHAR(255),
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_health_facilities_deleted
    ON health_facilities(deleted_at) WHERE deleted_at IS NOT NULL;

-- ── Assign a facility to a staff account ─────────────────────────────────────
ALTER TABLE staff_accounts ADD COLUMN IF NOT EXISTS health_facility_id UUID
    REFERENCES health_facilities(id);

CREATE INDEX IF NOT EXISTS idx_staff_health_facility
    ON staff_accounts(health_facility_id);

-- ── Seed initial Puskesmas regions ───────────────────────────────────────────
INSERT INTO health_facilities (name) VALUES
    ('Ulee Kareng'),
    ('Darussalam')
ON CONFLICT (name) DO NOTHING;
