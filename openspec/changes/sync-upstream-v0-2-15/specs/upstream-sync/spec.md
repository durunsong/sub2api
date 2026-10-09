## ADDED Requirements

### Requirement: 同步固定发布并保留 Fork
系统 MUST 包含官方 v0.2.15 的功能和修复，且保留所有既有 Fork 行为。

#### Scenario: 新旧平台并存
- **WHEN** 管理员创建平台账号、分组或配置配额
- **THEN** Command Code、Cline 和 Kiro 均可用，Composite 和调度使用统一平台目录

#### Scenario: 定制回归
- **WHEN** 运行 Fork 测试
- **THEN** Kiro、XorPay、封禁、审计、永久重置卡、Ops、品牌与支付 UI 的既有契约继续满足

### Requirement: 迁移与交付边界
同步 MUST 保留历史迁移字节，不自动部署或执行数据库变更。

#### Scenario: 检查迁移
- **WHEN** 对比 Fork HEAD 的历史 SQL
- **THEN** 全部一致，仅新增上游 242 迁移，Kiro 由应用目录校验
