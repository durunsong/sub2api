# 发布同步要求

## ADDED Requirements

### Requirement: 精确同步发布增量
系统 SHALL 包含 v0.2.24 的全部发布功能与修复，不包含后续 main。

#### Scenario: 上游功能与安全修复
- WHEN 合并后执行新版请求字段、搜索、模型和协议回归
- THEN 通过测试，版本为 0.2.24，无新增依赖或迁移

### Requirement: 保留 Fork 定制
系统 SHALL 保留 AGENTS.md 与 docs/FORK_VS_UPSTREAM.md 中的定制行为。

#### Scenario: 共享网关与定制兼容
- WHEN 使用 Kiro、支付、订阅、生图、封禁和品牌功能
- THEN 现有回归保持通过，生图内容不持久化，所有历史 SQL 保持不变

### Requirement: 审查后推送
变更 SHALL 在新鲜验证和独立审查通过后推送 origin/main。

#### Scenario: 交付可核验
- WHEN 推送成功
- THEN 远端提交与本地 HEAD 一致，文档记录命令、结果和残余限制
