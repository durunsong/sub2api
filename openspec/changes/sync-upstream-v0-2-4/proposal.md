## Why
将当前 v0.2.3 Fork 升级到官方 v0.2.4，同时保留用户已有定制，避免官方平台约束覆盖 Kiro。

## What Changes
- 合入官方 70 commits / 266 files / +5,696 / -726 中的应用、测试、文档和构建期源码增量。
- 新增 MiniMax、Grok 媒体资格、周成本估算、监控排行开关、Markdown 支付帮助及各项上游修复。
- 237 平台约束保留 Kiro；共享平台注册表、Composite、配额和调度同时支持 MiniMax 与 Kiro。
- 更新版本和 Fork 文档，记录定制保留与最终验证证据。

## Capabilities
### Modified Capabilities
- upstream-release-sync: 精确同步 release tag 并保留 Fork 扩展。

## Impact
Go 后端、Vue 前端、Ent 构建期源码、新迁移、现有 Redis 依赖及文档。受保护的 `.env*` 和部署 compose 配置保持原样；新增日志参数采用应用默认值，可选配置另行记录。

## Non-goals
不增加上游 tag 以外功能，不改支付/重置卡业务规则，不提交、推送、部署或执行数据库迁移，不读取秘密配置。
