## MODIFIED Requirements

### Requirement: Synchronize the pinned release while preserving customizations
系统 MUST 完整纳入官方 v0.2.3 增量，并保留所有 Fork 定制。

#### Scenario: Ollama protocol compatibility
- **WHEN** 实际请求目标为官方 Ollama Cloud 且符合上游模型/账号条件
- **THEN** 输出 token 限制和 Anthropic Bearer 鉴权符合 v0.2.3，非目标供应商行为不变。

#### Scenario: Model picker
- **WHEN** 实时目录缺少显示名或类型
- **THEN** 测试选择器补齐 ID/模型类型，不改共享目录、缓存或空列表语义。

#### Scenario: Database repair
- **WHEN** 236 在隔离升级环境执行
- **THEN** groups.model_allowlist 收敛到官方 schema，保留有效配置及全部 Fork 历史迁移。

#### Scenario: Fork preservation
- **WHEN** 同步完成
- **THEN** Kiro/XorPay/Access Ban/订阅重置卡/提示词审计/Ops/UI 与支付定制仍在，非重叠代码逐字节不变，重叠定制差异保持等价。
