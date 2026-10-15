-- 未活跃用户余额券群发：批次 + 逐收件人明细（每位收件人一张独立余额兑换码）。
-- 邮件模板走通知邮件事件 redeem.inactive_coupon（/admin/settings 可编辑），批次不保存模板。

CREATE TABLE IF NOT EXISTS redeem_email_batches (
    id                   BIGSERIAL PRIMARY KEY,
    status               VARCHAR(20) NOT NULL DEFAULT 'pending'
                         CHECK (status IN ('pending', 'running', 'succeeded', 'partial_failed', 'failed')),
    segments             TEXT[] NOT NULL
                         CHECK (cardinality(segments) > 0
                                AND segments <@ ARRAY['inactive_1_2m', 'inactive_2_3m', 'inactive_3m_plus', 'never_used']::TEXT[]),
    reference_time       TIMESTAMPTZ NOT NULL,
    coupon_value         NUMERIC(20, 8) NOT NULL CHECK (coupon_value > 0),
    code_expires_at      TIMESTAMPTZ,
    total_count          INT NOT NULL CHECK (total_count > 0),
    pending_count        INT NOT NULL DEFAULT 0 CHECK (pending_count >= 0),
    sending_count        INT NOT NULL DEFAULT 0 CHECK (sending_count >= 0),
    sent_count           INT NOT NULL DEFAULT 0 CHECK (sent_count >= 0),
    failed_count         INT NOT NULL DEFAULT 0 CHECK (failed_count >= 0),
    skipped_count        INT NOT NULL DEFAULT 0 CHECK (skipped_count >= 0),
    created_by           BIGINT NOT NULL REFERENCES users(id),
    idempotency_key_hash CHAR(64) NOT NULL,
    request_hash         CHAR(64) NOT NULL,
    started_at           TIMESTAMPTZ,
    finished_at          TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (created_by, idempotency_key_hash),
    CHECK (pending_count + sending_count + sent_count + failed_count + skipped_count = total_count)
);

CREATE INDEX IF NOT EXISTS idx_redeem_email_batches_created_at
    ON redeem_email_batches (created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS redeem_email_items (
    id               BIGSERIAL PRIMARY KEY,
    batch_id         BIGINT NOT NULL REFERENCES redeem_email_batches(id) ON DELETE CASCADE,
    position         INT NOT NULL CHECK (position >= 0),
    user_id          BIGINT REFERENCES users(id) ON DELETE SET NULL,
    recipient_name   VARCHAR(255) NOT NULL DEFAULT '',
    recipient_email  VARCHAR(255) NOT NULL,
    redeem_code_id   BIGINT UNIQUE REFERENCES redeem_codes(id) ON DELETE SET NULL,
    redeem_code      VARCHAR(32) NOT NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'pending'
                     CHECK (status IN ('pending', 'sending', 'sent', 'failed', 'skipped')),
    attempt_count    INT NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    attempt_token    UUID,
    lease_expires_at TIMESTAMPTZ,
    sent_at          TIMESTAMPTZ,
    last_attempt_at  TIMESTAMPTZ,
    error_code       VARCHAR(64),
    error_message    TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (batch_id, position),
    UNIQUE (batch_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_redeem_email_items_claim
    ON redeem_email_items (batch_id, position)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_redeem_email_items_sending_lease
    ON redeem_email_items (lease_expires_at)
    WHERE status = 'sending';

CREATE INDEX IF NOT EXISTS idx_redeem_email_items_batch_status
    ON redeem_email_items (batch_id, status, position);
