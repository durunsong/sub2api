## ADDED Requirements

### Requirement: Exact release synchronization
The fork SHALL include upstream v0.2.18 changes from the verified v0.2.17 baseline without importing later main commits.

#### Scenario: Release features are available
- **WHEN** the updated source is built
- **THEN** Vertex quota, Gemini image gating, stream cancellation accounting, Codex recovery and release security updates are present and VERSION is 0.2.18.

### Requirement: Fork compatibility
The fork SHALL retain Kiro, XorPay, Access Ban, subscription reset cards, coupon email campaigns, proxy protection and UI customizations.

#### Scenario: Overlapping upstream changes
- **WHEN** shared gateway, account or UI code changes
- **THEN** existing fork behavior and its regression checks remain valid, including zero-balance coupon targeting and Kiro cache/credits accounting.

### Requirement: Migration safety
Historical migrations SHALL remain byte-identical and the new Gemini migration SHALL coexist with the fork coupon migration.

#### Scenario: Source delivery
- **WHEN** the upgrade is delivered
- **THEN** migration files are supplied but not executed and no production data is changed.
