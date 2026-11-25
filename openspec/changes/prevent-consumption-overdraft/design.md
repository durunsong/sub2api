# Design

## Approved behavior and alternatives
The user selected option 1: wallet consumption floors at zero and the platform absorbs excess. Display-only masking was rejected; strict prepayment/admission is a separate alternative and is not implemented.

## Accounting
All USD amounts use existing NUMERIC(20,8) and half-away-from-zero quantization. One row-locked PostgreSQL statement settles all positive wallet costs for one user in command order. The wallet update and durable usage_balance_settlements rows are atomic with the existing billing dedup and quota transaction. Each row stores request/key/user, requested, charged, absorbed, before and after. requested = charged + absorbed; before - after = charged. Dedup/archive behavior and original fingerprints remain unchanged. The legacy path shares this statement with an independent legacy request ID.

For preexisting negative balances, charge zero and leave the old debt untouched. Do not silently forgive history. Normal nonnegative balances never become negative. Keep batch image reservation semantics; reject settlement above the rounded held amount so epsilon cannot create a negative wallet.

## Concurrency and caches
Row locks serialize consumption and recharge; never write a stale application snapshot. Batch and sequential results match. Redis decrement floors at zero (existing negative entries remain unchanged); use actual charged amount for cache deltas and notifications, retain database generation invalidation at exhaustion. Real usage and quota costs remain the full cost.

## Migration, rollback and boundaries
Add migration 254_usage_balance_settlements.sql without touching historical SQL or live data. No new dependencies. New code requires migration; rollback code only retains ledger records but restores old overdraft policy, so pause wallet traffic during rollback. Existing negative users need a separately approved adjustment. Do not execute production migrations or deploy. OpenSpec CLI is unavailable; write standard artifacts directly.

## Acceptance
$1 with $3 cost => balance 0, charged 1, absorbed 2. Exact balance reaches 0. Duplicate requests have no further effects. Concurrent charges never charge more than available funds; concurrent top-ups are preserved. Transaction failure rolls back balance and ledger. Single/batch/legacy paths agree. Eight-decimal rounding, missing users, existing negative balances, unchanged quota/subscription behavior and cache/notification results are covered.
