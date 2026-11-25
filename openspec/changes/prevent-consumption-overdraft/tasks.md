# Implementation Plan

**Goal:** implement approved option B without hidden debt or lost top-ups.
**Architecture:** shared row-locked settlement SQL with durable ledger; existing billing transaction/dedup, Redis invalidation and decimal precision.
**Tech stack:** Go, PostgreSQL, Redis; existing dependencies only.
**Spec:** specs/balance-settlement/spec.md and design.md.

- [x] Add failing integration tests for overdraw, batch equivalence, concurrency, top-up, duplicate billing, rollback and legacy deductions; add cache/notification cases.
- [x] Add additive settlement migration and shared atomic SQL; integrate single, batch and legacy paths.
- [x] Use actual charged amounts in cache and balance notifications; close batch-image epsilon overdraft path.
- [x] Run focused integration/unit/race tests, broader impacted suites, build/vet, migration and UTF-8/diff checks; independent review.
- [x] Update Fork documentation and record verification/remaining limits.

## Execution record
- Native worktree creation did not return after several minutes and was cancelled. The initial checkout is clean; continue on dedicated codex/balance-floor branch in the existing checkout.
- User already selected the accounting policy and requested implementation; no additional plan approval gate.
- Historical balances and production are untouched. The user subsequently authorized review, commit and push of this accounting change.

- Continued existing uncommitted implementation after the user confirmed this chat is authoritative. PostgreSQL floor/concurrency tests passed; remaining notification and batch-image rounding regressions were reproduced before fixing.

- Final verification: default backend suite (57 tested packages), full service/repository unit suite, focused service checks, 21 top-level integration cases/suites, 11 top-level race integration cases, build/vet, formatting and UTF-8 checks all passed. Independent review found no blocking issues. See verification.md.
