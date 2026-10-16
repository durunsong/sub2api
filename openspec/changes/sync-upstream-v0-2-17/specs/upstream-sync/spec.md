## ADDED Requirements

### Requirement: 精确发布同步
系统 SHALL 纳入 v0.2.16 到 v0.2.17 的非部署增量，版本为 0.2.17，不带入标签之后的提交。
#### Scenario: 新版模型和协议
- **WHEN** 使用 Gemini 4 Argon、思考档位或 OpenAI API Key 搜索模拟
- **THEN** 新版上游行为及其测试成立。

### Requirement: 保留 Fork 行为
系统 SHALL 保留现有定制和新增群发邮件功能。
#### Scenario: Kiro 与支付
- **WHEN** Kiro 搜索或订阅重置卡退款
- **THEN** 保留 Kiro 缓存计量、渠道开关以及来源卡事务作废和幂等行为。
#### Scenario: Pending 退款
- **WHEN** 网关尚未确认退款终态
- **THEN** 已扣金额保持扣除，成功时不重复扣除，失败时原子归还；来源重置卡仅成功时作废。

### Requirement: 安全交付
系统 SHALL 保持既有 SQL 和秘密/部署配置不变。
#### Scenario: 本地交付
- **WHEN** 同步完成
- **THEN** 交付未暂存差异和新鲜验证记录，不自动部署、迁移、提交或发送邮件。
