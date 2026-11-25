# Prevent consumption overdraft

User approved option 1 (formerly B) on 2026-11-25: finish admitted requests, deduct at most the available wallet balance, and record the excess as platform-absorbed cost. No presentation-only clamp.

Scope: single and batched usage settlement, legacy wallet deduction, cache, balance notifications, regression tests. Preserve actual usage/pricing and subscription/API-key/account quotas. No production changes or historical debt forgiveness.
