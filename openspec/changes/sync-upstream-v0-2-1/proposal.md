## Why

将已定制的 v0.2.0 升级到官方 v0.2.1，完整保留 Kiro、XorPay、Access Ban、提示词审计、订阅重置卡、订阅筛选、Ops 和 UI/支付定制，并纳入 Fork 已有的登录失败自动封禁与重置卡提示提交。

## What Changes

- 精确移植官方 release tag v0.2.0 到 v0.2.1 的 82 commits / 297 files / +12,622 / -1,033。
- 合入 Codex 固定账号模型目录、Astra/ultrafast、上游请求 ID、图片 URL 回填、定价热重载、Claude 推理计费及网关、Ops、支付修复。
- 追加四个官方迁移，保留全部历史 Fork 迁移及数据语义。
- 设置运行 VERSION 为 0.2.1，记录定制保护和验证证据。

## Capabilities

### Modified Capabilities

- `upstream-release-sync`: 同步官方发布增量并保留 Fork 行为。

## Impact

影响 Go 后端、Ent schema/生成源码、Vue 前端和版本文档；依赖清单与锁文件无上游变化。只修改工作区，不执行数据库迁移、部署、提交或推送。
