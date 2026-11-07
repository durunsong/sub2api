## ADDED Requirements

### Requirement: Exact release synchronization
The fork SHALL include upstream v0.2.20 source changes from v0.2.19 without importing later main commits; deployment configuration SHALL remain unchanged.

#### Scenario: Release build
- **WHEN** source is built
- **THEN** MCP/OAuth, admin-key permissions, email blacklist, image routing and release fixes are available and VERSION is 0.2.20.

### Requirement: Fork preservation
The fork SHALL retain Kiro, XorPay, Access Ban, reset cards, coupon campaigns, proxy protection, pricing semantics and customized UI.

#### Scenario: Shared code changes
- **WHEN** upstream changes overlap fork code
- **THEN** fork regression tests pass and upstream changes coexist with the customized behavior, including support QR and subscription countdown.

### Requirement: Migration and delivery safety
Historical migrations SHALL remain byte-identical; the OAuth migration SHALL be supplied without execution.

#### Scenario: Delivery
- **WHEN** this upgrade is delivered
- **THEN** no production data, deployment configuration, Git index or commits are modified.
