## ADDED Requirements

### Requirement: 精确同步官方 v0.2.8
系统 SHALL 合入官方 v0.2.8 release tag 的功能和修复，并显示版本 0.2.8。

#### Scenario: 目标范围
- **WHEN** 比较官方 v0.2.7 与 v0.2.8 的允许路径
- **THEN** 473 个 release 路径进入 Fork，且不带入 tag 外的 main 内容。

### Requirement: 保留 Fork 定制
系统 SHALL 保留 Kiro、XorPay、Access Ban、提示词审计、订阅重置卡、批量删除、Ops、VersionBadge、GLM 和 UI/支付定制。

#### Scenario: 配额和路由迁移
- **WHEN** 检查 `157`、`237`、`238` 迁移及共享平台目录
- **THEN** Kiro、MiniMax、OpenCode Go 约束同时存在，Access Ban 路径继续覆盖网关入口。

#### Scenario: 插件账号状态
- **WHEN** 插件使用已保存的账号 ID 请求凭据
- **THEN** disabled/error 账号返回空结果；active 但暂时暂停调度的账号仍可返回凭据。

### Requirement: 可追溯验证
升级 SHALL 记录实际验证命令、结果和环境限制。

#### Scenario: 交付
- **WHEN** 声明同步完成
- **THEN** 版本、测试、构建、编码、差异和未执行迁移/部署项均有记录。
