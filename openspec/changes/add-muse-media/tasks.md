# Muse Media Implementation Plan

**Goal:** Complete Muse images/videos in existing gateways and workbench.
**Architecture:** Custom OpenAI API-key gateway adapter with existing ownership and billing.
**Tech Stack:** Go, Gin, Vue, TypeScript; no dependencies.
**Spec:** specs/muse-media/spec.md and design.md.

## Global Constraints
No credentials, paid requests, production deployment, migrations or dependencies. Preserve all Fork customizations and concurrent source changes. Use apply_patch and test-first for behavior. The source reference is /tmp/sub2api-muse-contract-20261129 (public source only). On 2026-11-29 the user authorized committing and pushing after review and verification.

## Review Focus
Cross-user task reads; ambiguous paid create replay; repeated settlement; private/foreign media URLs and credential forwarding; unsupported input silently accepted.

### Task 1: Video Gateway
- [x] Test Muse video create/status/content contract and safety failures before implementation.
- [x] Implement dedicated service adapter using the existing media handler/Seedance pattern, custom OpenAI API-key account capability and explicit model mapping.
- [x] Use namespaced public task IDs, strict ownership binding, no ambiguous create failover, bounded content proxy and NoContentPersistence.
- [x] Require configured video pricing, preserve create metadata and use success-only one-shot settlement. Route POST /v1/videos and relevant status/content lookups for OpenAI/composite.
- [x] Run affected Go tests and report exact commands/results.

### Task 2: Images and Studios
- [x] Add failing tests for Muse images and studio discovery/request contracts.
- [x] Implement API-key-only Muse image compatibility and explicit pricing; existing image count accounting stays intact.
- [x] Extend video catalog/settings and frontend protocol/parameter controls for Muse, preserving existing local history/recovery.
- [x] Add administrator video pricing controls for OpenAI groups and Muse families.
- [x] Run frontend tests/typecheck and backend studio/image tests.

### Task 3: Verification and Delivery
- [x] Review complete change for ownership, billing, privacy, capability gates and regressions.
- [x] Run backend build/vet and focused tests; frontend lint/typecheck/tests/build and browser fixture acceptance.
- [x] Update AGENTS.md, Fork documentation and setup instructions; inspect final diff and UTF-8/BOM/replacement characters.
- [x] Integrate into original checkout only if concurrent changes can be preserved; otherwise retain reviewable worktree and explain exact status.
