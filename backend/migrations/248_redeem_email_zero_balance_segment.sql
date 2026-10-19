-- 余额为 0 的收件人筛选与未活跃分段互斥，批次 segments 只存 {zero_balance}。
-- 不修改 246；上线由启动迁移执行。历史未活跃批次仍然合法。

DO $$
DECLARE
  cname text;
BEGIN
  SELECT c.conname INTO cname
  FROM pg_constraint c
  WHERE c.conrelid = 'redeem_email_batches'::regclass
    AND c.contype = 'c'
    AND pg_get_constraintdef(c.oid) ILIKE '%inactive_1_2m%'
    AND pg_get_constraintdef(c.oid) NOT ILIKE '%zero_balance%';
  IF cname IS NOT NULL THEN
    EXECUTE format('ALTER TABLE redeem_email_batches DROP CONSTRAINT %I', cname);
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'redeem_email_batches_segments_check'
  ) THEN
    ALTER TABLE redeem_email_batches
      ADD CONSTRAINT redeem_email_batches_segments_check
      CHECK (
        cardinality(segments) > 0
        AND segments <@ ARRAY['inactive_1_2m', 'inactive_2_3m', 'inactive_3m_plus', 'never_used', 'zero_balance']::TEXT[]
        AND (NOT ('zero_balance' = ANY (segments)) OR cardinality(segments) = 1)
      );
  END IF;
END $$;
