## ADDED Requirements

### Requirement: 精确发布整合
系统 SHALL 包含官方 v0.2.22 发布标签的功能增量，保留本地定制，版本显示为 0.2.22。
#### Scenario: 官方新功能
- **WHEN** 使用 Ultra Fast、Vertex 多区域、透传模型过滤或并发扣费
- **THEN** 行为符合 v0.2.22 且不引入标签后的代码。

### Requirement: Fork 定制兼容
系统 SHALL 保留 docs/FORK_VS_UPSTREAM.md 与 AGENTS.md 的全部定制，尤其 Kiro、支付、封禁、订阅全期额度与永久重置卡。
#### Scenario: 长期订阅扣费
- **WHEN** 90 天订阅跨越第 30/60 天并产生合并计费
- **THEN** 周期额度仍累计到 expires_at，不自动补额；重置记录与重置卡语义不变。
#### Scenario: Kiro 与新协议修复并存
- **WHEN** Kiro 直连或桥接请求经过共享网关
- **THEN** credits、缓存计量、真实请求头、代理与粘性策略保持。
#### Scenario: 用户用量界面
- **WHEN** 用户查看费用分布与缓存命中率
- **THEN** 原有品牌、倍率/标准成本隐藏与命中率展示保持，管理员仍可查看账号成本。

### Requirement: 可验证交付
系统 SHALL 提供验证记录，保持历史迁移与生产配置，不执行部署或真实数据库变更。
#### Scenario: 迁移共存
- **WHEN** 检查新增 251 SQL
- **THEN** 既有订阅历史 251 SQL 完整保留，两份上游新迁移按完整文件名共存。
