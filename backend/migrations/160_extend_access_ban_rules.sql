-- Extend global ip_bans into multi-type access ban rules (IP, UA, IP+UA, email suffix).

ALTER TABLE ip_bans
    ADD COLUMN IF NOT EXISTS rule_type VARCHAR(20) NOT NULL DEFAULT 'ip';

ALTER TABLE ip_bans
    ADD COLUMN IF NOT EXISTS ua_pattern VARCHAR(255);

ALTER TABLE ip_bans
    ALTER COLUMN pattern TYPE VARCHAR(255);

DROP INDEX IF EXISTS idx_ip_bans_pattern_active_unique;

CREATE UNIQUE INDEX IF NOT EXISTS idx_ip_bans_rule_pattern_active_unique
    ON ip_bans (rule_type, pattern, COALESCE(ua_pattern, ''))
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_ip_bans_rule_type ON ip_bans(rule_type);

COMMENT ON COLUMN ip_bans.rule_type IS 'Rule type: ip, ua, ip_ua, email_suffix';
COMMENT ON COLUMN ip_bans.ua_pattern IS 'User-Agent match pattern for ua / ip_ua rules';
COMMENT ON COLUMN ip_bans.pattern IS 'Primary match pattern: IP/CIDR, UA substring, or email suffix';
