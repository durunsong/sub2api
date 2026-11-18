## ADDED Requirements

### Requirement: 精确同步且保留定制
系统 SHALL 合入官方 v0.2.23 发布差异，保留 Fork 行为，不替换历史迁移。

#### Scenario: 聚合用量缺失
- **WHEN** 返回缓存写入 TTL 明细但没有总量
- **THEN** 扣费、用量日志和订阅 token 使用一致的写入总量，Kiro credits 独立保留。

#### Scenario: 渠道倍率
- **WHEN** 管理员保存 Ultra Fast 倍率或留空
- **THEN** 渠道使用已配置倍率或模型默认价，原有价格卡无需回填。

#### Scenario: 用户生图工作台
- **WHEN** 用户局部编辑、独立提交任务、刷新或切换标签页
- **THEN** 遮罩、本地历史和跨标签任务状态保留，刷新不重放扣费请求。
