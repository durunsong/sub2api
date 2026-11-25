## ADDED Requirements

### Requirement: Nonnegative consumption settlement
Wallet consumption MUST deduct only available funds and record requested, charged and platform-absorbed USD amounts atomically. Existing negative debt MUST NOT be forgiven or increased by consumption.

#### Scenario: Request exceeds funds
- **WHEN** a wallet has 1 USD and a completed request costs 3 USD
- **THEN** its balance is 0, charged amount is 1, and absorbed amount is 2

#### Scenario: Concurrent and repeated billing
- **WHEN** concurrent requests exhaust a wallet or a billing request is replayed
- **THEN** funds are consumed at most once under existing dedup semantics, balance and ledger agree, and recharge is not overwritten

#### Scenario: Batch equivalence
- **WHEN** commands are applied together or sequentially in the same order
- **THEN** per-request balance/charged/absorbed results and quota effects match

#### Scenario: Existing debt
- **WHEN** a previously negative wallet has an already admitted request settle
- **THEN** no additional wallet amount is deducted and the existing debt is unchanged

### Requirement: Consistent cache and notifications
Caches and balance notifications MUST use actual wallet deduction and MUST NOT introduce negative values for nonnegative balances. Real usage cost and quotas MUST remain unmodified.
