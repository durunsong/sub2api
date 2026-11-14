CREATE TABLE IF NOT EXISTS user_subscription_quota_resets (
    id BIGSERIAL PRIMARY KEY,
    user_subscription_id BIGINT NOT NULL REFERENCES user_subscriptions(id) ON DELETE RESTRICT,
    reset_at TIMESTAMPTZ NOT NULL,
    daily BOOLEAN NOT NULL,
    weekly BOOLEAN NOT NULL,
    monthly BOOLEAN NOT NULL,
    CONSTRAINT user_subscription_quota_resets_has_window CHECK (daily OR weekly OR monthly)
);

CREATE INDEX IF NOT EXISTS idx_user_subscription_quota_resets_history
    ON user_subscription_quota_resets (user_subscription_id, reset_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_user_subscription_reset_cards_history
    ON user_subscription_reset_cards (user_subscription_id, consumed_at DESC, id DESC)
    WHERE consumed_at IS NOT NULL;
