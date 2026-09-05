## ADDED Requirements

### Requirement: 精确同步 v0.2.1

系统 MUST 包含官方 v0.2.0 到 v0.2.1 release tag 的全部功能与修复，不引入后续 main 提交。

#### Scenario: 上游功能可用

- **WHEN** 使用 Codex 模型目录、Astra/ultrafast、请求 ID、图片回填、定价热重载及相关网关接口
- **THEN** 行为与官方 v0.2.1 契约一致，并通过官方相关测试

### Requirement: 完整保留 Fork

系统 MUST 保留保护基线 e316f5189 的 Kiro、XorPay、Access Ban、提示词审计、订阅重置卡和筛选、Ops、品牌及支付交互。

#### Scenario: 升级后继续使用定制

- **WHEN** 使用原有 Kiro、支付、访问封禁、登录防护、订阅重置卡及 UI 功能
- **THEN** 原有接口、计费、期限、幂等、可信 IP 和交互语义保持，并通过对应回归

### Requirement: 迁移追加与交付可审阅

系统 MUST 追加四个官方迁移，保留历史迁移原文，显示版本 0.2.1，并留下真实验证证据。

#### Scenario: 本地交付

- **WHEN** 完成升级
- **THEN** 工作区保持未提交，生产迁移未执行，差异无冲突/乱码，验证失败或未运行项明确记录
