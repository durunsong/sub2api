# Implementation plan

- [x] 1. Backend TDD: video config/catalog, strict parameters, private JWT relay, status/content and privacy routes; targeted Go tests.
- [x] 2. Frontend TDD: parameters/API, conditional IndexedDB storage, sessions and resumable controller; Vitest.
- [x] 3. UI: upstream video composer/settings/results/history/prompts/assets; user/admin Tab integration and translations.
- [x] 4. Verify: real-browser storage/recovery/account isolation/mobile/themes, image regressions, types/lint/build, fresh review and UTF-8/diff checks.

Evidence and final review: [verification.md](verification.md).

No automatic Git commits or deployment. Keep changes in the requested clean sub2api checkout.
