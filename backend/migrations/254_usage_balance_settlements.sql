-- Durable wallet settlement ledger. No historical balances are changed.
CREATE TABLE IF NOT EXISTS usage_balance_settlements (
    id BIGSERIAL PRIMARY KEY,
    request_id VARCHAR(255) NOT NULL,
    api_key_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    requested_amount NUMERIC(20,8) NOT NULL CHECK (requested_amount > 0),
    charged_amount NUMERIC(20,8) NOT NULL CHECK (charged_amount >= 0),
    absorbed_amount NUMERIC(20,8) NOT NULL CHECK (absorbed_amount >= 0),
    balance_before NUMERIC(20,8) NOT NULL,
    balance_after NUMERIC(20,8) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (request_id, api_key_id),
    CHECK (requested_amount = charged_amount + absorbed_amount),
    CHECK (balance_before - balance_after = charged_amount),
    CHECK (charged_amount <= GREATEST(balance_before, 0))
);
CREATE INDEX IF NOT EXISTS idx_usage_balance_settlements_user_created
    ON usage_balance_settlements (user_id, created_at DESC);
