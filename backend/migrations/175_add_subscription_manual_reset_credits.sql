-- Credits granted when a user repurchases a still-active subscription.
-- Consumed atomically by POST /subscriptions/:id/reset-daily.
ALTER TABLE user_subscriptions
    ADD COLUMN IF NOT EXISTS manual_reset_credits INTEGER NOT NULL DEFAULT 0;
