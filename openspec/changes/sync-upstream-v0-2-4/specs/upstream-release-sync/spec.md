## ADDED Requirements
### Requirement: 精确同步 v0.2.4
系统 SHALL 纳入目标 release 的应用增量，不带入 main 后续提交，并设置版本 0.2.4。
#### Scenario: MiniMax 与现有平台并存
- **WHEN** 管理员配置 MiniMax 或已有 Kiro 账号/分组
- **THEN** 共享目录、路由、配额、调度均保留两者；237 数据库约束兼容现有 Kiro 行。
### Requirement: 保留全部 Fork 定制
系统 SHALL 保留 docs/FORK_VS_UPSTREAM.md 定义的定制行为和历史迁移。
#### Scenario: 支付与订阅
- **WHEN** 有效固定期限订阅重复获得或消费重置卡
- **THEN** 原有按期限发卡、来源幂等、清零 USD/token、重开周期和退款事务规则保持不变。
#### Scenario: 访问与界面
- **WHEN** 请求经过网关、认证或用户界面
- **THEN** Access Ban、登录自动封禁、Kiro→Claude、XorPay、GLM 筛选及隐藏在线更新继续生效。
### Requirement: 可审计验证
同步 SHALL 提供测试、构建、迁移保留和文件差异证据。
#### Scenario: 本地交付
- **WHEN** 完成同步
- **THEN** 给出实际检查结果及环境限制；不执行提交、生产部署和真实迁移。
