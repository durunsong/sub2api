## ADDED Requirements

### Requirement: Recent activity audiences
The campaign SHALL support `active_7d` and `active_30d` using the latest API usage timestamp and rolling UTC durations, including the lower boundary.

#### Scenario: Inclusive and nested windows
- **WHEN** the latest call is exactly 7 days ago
- **THEN** the user qualifies for both active audiences
- **WHEN** the latest call is exactly 30 days ago
- **THEN** the user qualifies for `active_30d` only

#### Scenario: Audience union
- **WHEN** overlapping audiences are selected
- **THEN** each eligible user appears once, with one coupon per batch

#### Scenario: Existing protections
- **WHEN** recent activity is selected
- **THEN** disabled, deleted, administrative and invalid-email users remain excluded, and zero balance cannot be combined with other audiences

#### Scenario: Create-time revalidation
- **WHEN** a selected user no longer qualifies at batch creation
- **THEN** the transaction rejects the batch without issuing coupons

#### Scenario: UI and persistence
- **WHEN** an administrator selects recent activity
- **THEN** search, confirmation, creation and history preserve the audience, with Chinese and English labels, and the database accepts the batch
