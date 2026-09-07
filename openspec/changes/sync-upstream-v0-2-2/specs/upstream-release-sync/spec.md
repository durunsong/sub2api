## MODIFIED Requirements

### Requirement: Import the exact upstream release

The fork SHALL incorporate the v0.2.2 release at commit `5485f368b29d05adb95a00f71801c7c23d8f48af`, excluding later main commits, and display version 0.2.2.

#### Scenario: Compare release input
- **WHEN** the final working tree is compared with the v0.2.1 to v0.2.2 delta
- **THEN** every changed path is accounted for, including rename adaptation and the version override

### Requirement: Preserve fork behavior

The fork SHALL retain Kiro routing, credits and cooldown, XorPay, Access Ban including trusted-IP login protection, prompt audit, subscription reset cards and availability filtering, Ops deletion, Claude user branding, GLM plan classification and payment/UI customizations.

#### Scenario: Shared gateway and subscription code changes
- **WHEN** upstream modifies gateway admission, simple-mode grouping or trusted fulfillment
- **THEN** Kiro remains supported, Access Ban still intercepts requests, and subscription reset-card source idempotency and validity snapshots remain unchanged

### Requirement: Preserve migration history

The fork SHALL add the official 235 migration without rewriting existing migrations and document the change from models-list display configuration to request admission.

#### Scenario: Existing model configuration
- **WHEN** a separately authorized deployment applies migration 235
- **THEN** existing values survive the column rename and nonempty allowlists constrain both model listing and request admission

### Requirement: Verify before delivery

The fork SHALL provide fresh test/build/static-check results, customization preservation evidence and UTF-8 checks; unavailable infrastructure tests SHALL be reported explicitly.

#### Scenario: Local verification
- **WHEN** source integration is complete
- **THEN** local automated checks are run against final source, limitations are documented, and no automatic commit, push, migration or deployment occurs
