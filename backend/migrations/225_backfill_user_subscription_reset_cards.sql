INSERT INTO user_subscription_reset_cards
    (user_subscription_id, validity_days, source_type, source_reference, source_sequence, created_at)
SELECT us.id,
       LEAST(GREATEST(g.default_validity_days, 1), 36500),
       'legacy_backfill',
       us.id::text,
       sequence_no,
       NOW()
FROM user_subscriptions us
JOIN groups g ON g.id = us.group_id
CROSS JOIN LATERAL generate_series(1, us.manual_reset_credits) AS sequence_no
WHERE us.manual_reset_credits > 0
ON CONFLICT (user_subscription_id, source_type, source_reference, source_sequence) DO NOTHING;
