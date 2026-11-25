# Verification — 2026-11-25

## Scope and accounting

Approved option 1: admitted requests complete; new consumption never takes a nonnegative wallet below zero. Requested = wallet charged + platform absorbed. Existing historical debt is neither forgiven nor enlarged. Real usage and quota counters retain full cost. No frontend masking, production migration, deployment, or historical adjustment.

## Verified results

- PostgreSQL/Redis integration: `go test -v -tags=integration ./internal/repository -run 'TestUsageBalanceFloor|TestUsageBillingRepository|TestUserRepoSuite' -count=1` — PASS. Covers zero/exact/insufficient funds, eight-decimal amounts, existing debt, legacy deduction, single/batch equality, duplicate/archive requests, concurrent single/batch charges, concurrent recharge, rollback of wallet/ledger/dedup on later quota failure, missing users, existing user mutation regressions.
- Repository unit tests: `go test -tags=unit ./internal/repository -count=1` — PASS.
- `go build ./... && go vet ./...` — PASS.
- Reproduced before fixes: notification reconstructed old balance 3 instead of 1; batch-image settlement accepted 0.500000009 against 0.5 held balance. Added regression checks and corrected both paths.
- Independent read-only review of SQL, dedup, precision, caches, notifications, migration and legacy path found no blocking P0/P1/P2 findings.
- Historical SQL files unchanged. New migration creates only `usage_balance_settlements`. UTF-8/BOM/replacement-character and whitespace checks passed.

## Final suites

- `go test ./... -count=1` — PASS, 57 tested packages (other packages have no tests).
- `go test -p 1 -tags=unit ./internal/service ./internal/repository -count=1` — PASS (service 301.815s, repository 12.489s).
- `go test -tags=unit ./internal/service -run 'TestResolveOldBalance|TestBalanceNotificationAmounts|TestSyncBalanceCache|TestCheckBillingEligibility|TestUsageNotifications|TestBalanceNotify' -count=1` — PASS. Covers notification amounts, settled cache precision in both synchronous and queued paths, eligibility and notification worker lifecycle.
- `go test -race -v -tags=integration ./internal/repository -run 'TestUsageBalanceFloor|TestUsageBillingRepositoryApplyBatch' -count=1` — PASS, 11 top-level cases; real PostgreSQL and Redis containers, no data races.
- The earlier broad integration command passed 21 top-level cases/suites. An initial compile failure from a test package's integer `max` helper shadowing the builtin was fixed by using `math.Min`/`math.Max`. An older redundant unit compilation was stopped to reduce resource contention, then the final complete unit run above passed.
- Final `git diff --check`, `gofmt -l` on changed Go files, UTF-8 decoding, BOM/replacement-character checks and local documentation-reference checks — PASS.
- Frontend tests were not rerun: there are no frontend changes. No real paid API calls or production data verification were performed.

## Pre-push review

- Reviewed the working tree against `6813ba960`, including every new file. Independent standards review found no violations; spec review found one rounding edge case, fixed and re-reviewed before commit.
- Reproduced `ActualCost=0.000000004` returning a nonzero cache deduction while the database rounds the charge to zero. `TestResolveBalanceCharge_RoundedZeroCost` failed with `4e-09` before the fix. The no-snapshot fallback now uses the same eight-decimal quantization as wallet settlement; actual usage cost remains unchanged.
- After the fix, `go test -tags=unit ./internal/service ./internal/repository -run 'TestResolve|TestBalanceNotificationAmounts|TestSyncBalanceCache|TestBillingCache|TestApplyUsageBillingEffects|TestSettleUsageBillingBalances|TestCaptureUsageBillingBatchImageBalance|TestCheckBillingEligibility' -count=1` passed (service 3.220s, repository 1.640s).
- Re-ran the PostgreSQL/Redis integration command above: all 21 top-level cases/suites passed (14.575s), with no skips. No repository code changed afterward.
- Re-ran `go build ./... && go vet ./...` after the fix: passed. The earlier full default/unit/race logs were also checked and contain no failures; those broad suites were not repeated for the one-line cache fallback fix.
- Final changed-file UTF-8, BOM, replacement-character, personal-path, gofmt and `git diff --check` checks passed.

## Deployment and rollback

Apply additive migration 254 before serving wallet traffic on the new code. Do not delete settlement history during a code rollback. Old code restores overdraft behavior; pause wallet traffic and reconcile in-flight settlement when switching versions. Historical negative users need a separate audited adjustment if debt is to be forgiven. No live data has been read or changed.
