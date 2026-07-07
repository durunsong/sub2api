-- Global IP/CIDR ban rules for registration and API key gateway access.
CREATE TABLE IF NOT EXISTS ip_bans (
    id          BIGSERIAL PRIMARY KEY,
    pattern     VARCHAR(64)  NOT NULL,
    status      VARCHAR(20)  NOT NULL DEFAULT 'active',
    reason      VARCHAR(255),
    source      VARCHAR(50)  NOT NULL DEFAULT 'manual',
    created_by  BIGINT,
    expires_at  TIMESTAMPTZ,
    last_hit_at TIMESTAMPTZ,
    hit_count   BIGINT       NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ip_bans_pattern_active_unique
    ON ip_bans(pattern)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_ip_bans_status ON ip_bans(status);
CREATE INDEX IF NOT EXISTS idx_ip_bans_expires_at ON ip_bans(expires_at);
CREATE INDEX IF NOT EXISTS idx_ip_bans_last_hit_at ON ip_bans(last_hit_at);
CREATE INDEX IF NOT EXISTS idx_ip_bans_deleted_at ON ip_bans(deleted_at);

COMMENT ON TABLE ip_bans IS 'Global client IP/CIDR ban rules';
COMMENT ON COLUMN ip_bans.pattern IS 'Banned IP or CIDR pattern, e.g. 203.0.113.10 or 203.0.113.0/24';
COMMENT ON COLUMN ip_bans.status IS 'Rule status: active or inactive';
COMMENT ON COLUMN ip_bans.reason IS 'Admin-visible ban reason';
COMMENT ON COLUMN ip_bans.source IS 'Ban source, e.g. manual or automation';
COMMENT ON COLUMN ip_bans.created_by IS 'Admin user id that created the rule';
COMMENT ON COLUMN ip_bans.expires_at IS 'Expiration time; NULL means permanent';
COMMENT ON COLUMN ip_bans.last_hit_at IS 'Last time this rule blocked a request';
COMMENT ON COLUMN ip_bans.hit_count IS 'Number of blocked requests';
