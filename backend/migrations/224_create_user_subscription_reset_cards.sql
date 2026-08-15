CREATE TABLE IF NOT EXISTS user_subscription_reset_cards (
    id BIGSERIAL PRIMARY KEY,
    user_subscription_id BIGINT NOT NULL REFERENCES user_subscriptions(id) ON DELETE RESTRICT,
    validity_days INTEGER NOT NULL CHECK (validity_days BETWEEN 1 AND 36500),
    source_type VARCHAR(32) NOT NULL,
    source_reference VARCHAR(255) NOT NULL,
    source_sequence INTEGER NOT NULL DEFAULT 1 CHECK (source_sequence > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    consumed_at TIMESTAMPTZ,
    consume_idempotency_key VARCHAR(128),
    voided_at TIMESTAMPTZ,
    CONSTRAINT user_subscription_reset_cards_consumed_after_created CHECK
        (consumed_at IS NULL OR consumed_at >= created_at),
    CONSTRAINT user_subscription_reset_cards_source_uniq UNIQUE
        (user_subscription_id, source_type, source_reference, source_sequence)
);

CREATE INDEX IF NOT EXISTS idx_user_subscription_reset_cards_available
    ON user_subscription_reset_cards (user_subscription_id, validity_days, id)
    WHERE consumed_at IS NULL AND voided_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_subscription_reset_cards_consume_idempotency
    ON user_subscription_reset_cards (user_subscription_id, consume_idempotency_key)
    WHERE consume_idempotency_key IS NOT NULL;
