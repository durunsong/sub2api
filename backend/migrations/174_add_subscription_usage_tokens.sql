-- Per-subscription raw token counters (independent of group rate multipliers).
-- Mirrored after daily/weekly/monthly USD windows so quota resets clear tokens too.
ALTER TABLE user_subscriptions
    ADD COLUMN IF NOT EXISTS daily_usage_tokens   BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS weekly_usage_tokens  BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS monthly_usage_tokens BIGINT NOT NULL DEFAULT 0;

-- Best-effort backfill for currently active usage windows from raw usage_logs.
-- Uses input/output/cache token columns only (same formula as UsageLog.TotalTokens).
UPDATE user_subscriptions us
SET
    daily_usage_tokens = CASE
        WHEN us.daily_window_start IS NULL THEN 0
        ELSE COALESCE((
            SELECT SUM(
                COALESCE(ul.input_tokens, 0)
                + COALESCE(ul.output_tokens, 0)
                + COALESCE(ul.cache_creation_tokens, 0)
                + COALESCE(ul.cache_read_tokens, 0)
            )
            FROM usage_logs ul
            WHERE ul.subscription_id = us.id
              AND ul.created_at >= us.daily_window_start
        ), 0)
    END,
    weekly_usage_tokens = CASE
        WHEN us.weekly_window_start IS NULL THEN 0
        ELSE COALESCE((
            SELECT SUM(
                COALESCE(ul.input_tokens, 0)
                + COALESCE(ul.output_tokens, 0)
                + COALESCE(ul.cache_creation_tokens, 0)
                + COALESCE(ul.cache_read_tokens, 0)
            )
            FROM usage_logs ul
            WHERE ul.subscription_id = us.id
              AND ul.created_at >= us.weekly_window_start
        ), 0)
    END,
    monthly_usage_tokens = CASE
        WHEN us.monthly_window_start IS NULL THEN 0
        ELSE COALESCE((
            SELECT SUM(
                COALESCE(ul.input_tokens, 0)
                + COALESCE(ul.output_tokens, 0)
                + COALESCE(ul.cache_creation_tokens, 0)
                + COALESCE(ul.cache_read_tokens, 0)
            )
            FROM usage_logs ul
            WHERE ul.subscription_id = us.id
              AND ul.created_at >= us.monthly_window_start
        ), 0)
    END
WHERE us.deleted_at IS NULL
  AND (
      us.daily_window_start IS NOT NULL
      OR us.weekly_window_start IS NOT NULL
      OR us.monthly_window_start IS NOT NULL
  );
