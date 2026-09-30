## ADDED Requirements

### Requirement: Synchronize the exact v0.2.11 release
The fork SHALL incorporate the upstream v0.2.10 to v0.2.11 functional delta and report version 0.2.11, excluding protected deployment configuration.

#### Scenario: Release features remain available
- WHEN running the upgraded fork
- THEN balance inflight reservations, API key creation limits, native Claude reset redemption, GPT-6.1 Sol/Astra Ultrafast support and remote Codex model catalogs are available with upstream defaults.

### Requirement: Preserve all fork customizations
The upgrade SHALL retain Kiro, XorPay, Access Ban, subscription reset cards, billing customizations and customized UI behavior.

#### Scenario: Existing fork behavior survives synchronization
- WHEN upstream code overlaps custom code
- THEN the merge preserves custom behavior and its regression tests, including unchanged historical migrations and independent native Claude/user subscription resets.

### Requirement: Verify before delivery
The upgrade SHALL pass applicable backend/frontend automated checks and text integrity checks without running production operations.

#### Scenario: Reviewable local delivery
- WHEN implementation is complete
- THEN the changed files, commands, actual results, excluded configuration and remaining verification limits are recorded without committing or deploying.
