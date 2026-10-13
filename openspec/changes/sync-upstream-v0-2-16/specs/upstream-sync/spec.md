## ADDED Requirements

### Requirement: 同步固定发布并保留 Fork
系统 MUST 包含官方 v0.2.16 的功能和修复，且保留所有既有 Fork 行为。

#### Scenario: Gemini 缓存与 Kiro 并存
- **WHEN** 客户端调用 Gemini cachedContents，或 Kiro 账号发生无法定价的用量
- **THEN** 缓存路由经过 Access Ban，存储费可报价；Kiro 仍使用 credits 与保守兜底，不改到期重置卡

#### Scenario: 定制回归
- **WHEN** 运行 Fork 测试
- **THEN** Kiro、XorPay、封禁、审计、永久重置卡、Ops、品牌与支付 UI 的既有契约继续满足

### Requirement: 迁移与交付边界
同步 MUST 保留历史迁移字节，不自动部署或执行数据库变更。

#### Scenario: 检查迁移
- **WHEN** 对比 Fork HEAD 的历史 SQL
- **THEN** 全部一致；仅新增 `242_group_i18n.sql` 与 `243`/`244`/`245`，既有 `242_drop_platform_check_constraints.sql` 保持不变
