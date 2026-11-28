# Upstream Sync

## ADDED Requirements

### Requirement: Integrate the pinned release while preserving the fork
The fork SHALL expose the v0.2.26 release functionality while retaining Kiro, XorPay, Access Ban, subscription term quotas/reset cards/history, capped wallet settlement, coupon campaigns, branding, and image/video studios.

#### Scenario: Shared gateway or statistics code changes
- **WHEN** upstream code changes a shared path
- **THEN** existing Kiro billing, sticky TTL and privacy guarantees remain active alongside the upstream behavior.

### Requirement: Image safety events respect local-only studio content
Image safety reporting SHALL preserve the NoContentPersistence boundary, including asynchronous work, while ordinary gateway requests retain upstream risk-center reporting.

#### Scenario: Studio image request is refused for safety
- **WHEN** the gateway reports an upstream safety refusal for a private studio request
- **THEN** no prompt or upstream-echoed creative content is persisted, and refusal handling remains effective.

### Requirement: Verified and traceable delivery
The upgrade SHALL retain historical migration bytes, set VERSION to 0.2.26, pass applicable checks and independent review, and use a normal authorized push without rewriting remote history.
