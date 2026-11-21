-- 扩展余额券邮件收件人范围，保留历史分段与零余额互斥约束。
ALTER TABLE redeem_email_batches
  DROP CONSTRAINT IF EXISTS redeem_email_batches_segments_check;

ALTER TABLE redeem_email_batches
  ADD CONSTRAINT redeem_email_batches_segments_check
  CHECK (
    cardinality(segments) > 0
    AND segments <@ ARRAY['inactive_1_2m', 'inactive_2_3m', 'inactive_3m_plus', 'never_used', 'active_7d', 'active_30d', 'zero_balance']::TEXT[]
    AND (NOT ('zero_balance' = ANY (segments)) OR cardinality(segments) = 1)
  );
