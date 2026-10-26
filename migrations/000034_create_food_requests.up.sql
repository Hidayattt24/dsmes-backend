CREATE TABLE IF NOT EXISTS food_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    serving_size VARCHAR(255) NOT NULL,
    estimated_calories NUMERIC(10,2) DEFAULT 0,
    meal_type VARCHAR(50) DEFAULT '',
    notes TEXT DEFAULT '',
    image_url TEXT DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    admin_notes TEXT DEFAULT '',
    reviewed_by UUID REFERENCES staff_accounts(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    approved_food_id UUID REFERENCES foods(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_food_requests_patient_id ON food_requests(patient_id);
CREATE INDEX IF NOT EXISTS idx_food_requests_status ON food_requests(status);
CREATE INDEX IF NOT EXISTS idx_food_requests_created_at ON food_requests(created_at DESC);
