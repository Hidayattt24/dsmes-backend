CREATE TABLE ai_conversations (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id  UUID         NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL DEFAULT 'Percakapan Baru',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_ai_conversations_patient ON ai_conversations(patient_id);
CREATE INDEX idx_ai_conversations_deleted ON ai_conversations(deleted_at);

CREATE TABLE ai_messages (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID        NOT NULL REFERENCES ai_conversations(id) ON DELETE CASCADE,
    patient_id      UUID        NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    role            VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    message         TEXT        NOT NULL DEFAULT '',
    token_count     INT         NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_ai_messages_conversation ON ai_messages(conversation_id);
CREATE INDEX idx_ai_messages_patient ON ai_messages(patient_id);
CREATE INDEX idx_ai_messages_created ON ai_messages(created_at);
CREATE INDEX idx_ai_messages_deleted ON ai_messages(deleted_at);

CREATE TABLE ai_prompt_logs (
    id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id        UUID         NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    conversation_id   UUID         REFERENCES ai_conversations(id) ON DELETE SET NULL,
    generated_prompt  TEXT         NOT NULL DEFAULT '',
    model             VARCHAR(100) NOT NULL DEFAULT '',
    execution_time_ms INT          NOT NULL DEFAULT 0,
    status            VARCHAR(20)  NOT NULL DEFAULT 'success',
    error_message     TEXT         NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_prompt_logs_patient ON ai_prompt_logs(patient_id);
