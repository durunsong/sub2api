## ADDED Requirements

### Requirement: 同步固定上游版本
系统 SHALL 提供官方 v0.2.10 的新增能力和修复，并显示 0.2.10。

#### Scenario: 上游协议与界面升级
- **WHEN** 使用 Sonnet 5.5、Claude 额度查询、风控白名单、费用趋势或修复后的协议路径
- **THEN** 行为符合固定目标标签及其回归测试

### Requirement: 保留 Fork 定制
系统 SHALL 保留 Kiro、XorPay、Access Ban、提示词审计、订阅重置卡、可用额度筛选、Ops 与 UI 定制。

#### Scenario: 升级后使用现有定制
- **WHEN** 运行现有定制回归并核对源码与历史迁移
- **THEN** 定制链路保持可用，非重叠定制源码和历史 SQL 不变，Wire 同时包含新服务与原有依赖
