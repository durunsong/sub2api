## Why

将当前定制 Fork 从官方 v0.2.1 同步到 v0.2.2，同时保留已有 Kiro、XorPay、Access Ban、提示词审计、订阅重置卡、Ops 和 UI/支付定制。

## What Changes

- 固定官方 v0.2.2 tag，移植 102 commits / 261 files / +13,430 / -2,005 的 release 增量。
- 合入分组模型白名单、简易模式分组、Codex 模型发现与 Astra 兼容、推理拒绝映射、支付发货隔离，以及网关、计费、备份和前端修复。
- 新增官方 235 迁移源码；保留既有迁移原文。当前显示版本设为 0.2.2。
- 更新定制清单和可复现验证记录。

## Capabilities

### Modified Capabilities

- `upstream-release-sync`: v0.2.2 release 增量完整性与 Fork 定制兼容。

## Impact

涉及 Go 后端、Ent schema/生成源码、Vue 前端和升级文档。无依赖版本变化。模型白名单由旧模型展示配置迁移而来，将同时影响实际请求准入，部署前必须核对现有非空配置。

## Non-Goals

不引入 release tag 之后的 main，不修改秘密配置，不执行数据库迁移或生产操作，不自动暂存、提交、推送或发布。
