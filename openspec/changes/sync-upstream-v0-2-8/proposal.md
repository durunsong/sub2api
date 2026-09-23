# 升级至 v0.2.8 并保留 Fork 定制

## Why
当前 Fork 基于官方 v0.2.7，需要同步作者发布的 v0.2.8，同时继续支持 Kiro、XorPay、Access Ban、订阅重置卡和现有 UI/支付定制。

## What Changes
合入官方 `v0.2.8` tag（commit `fd80b08c90b55edcad5b00171b53f08721d30da1`）的 OpenCode Go 用量窗口、Claude Code 版本同步、reasoning effort 倍率计费、月度备份归档、联盟线下提现幂等、Codex 推荐/积分、TypeSafe 内容审核、GPT-6 Sol/Luna/Claude Opus 5.5、图片余额和网关/前端修复；新增迁移 `238b`、`239`、`240`。

## Boundaries
保留 Fork 的 Kiro/XorPay/Access Ban、提示词审计、批量删用户、订阅重置卡、`active_available`、Ops 删除、VersionBadge 禁用在线更新、GLM 分类、支付金额分离和品牌定制。只同步 release tag，不带入 tag 之外的 main；不执行数据库迁移、部署、真实供应商调用或秘密配置操作。

## Acceptance
版本脚本输出 `0.2.8`；上游 tag 的允许文件进入；关键 Fork 路径和配额 CHECK 保留 Kiro/MiniMax/OpenCode Go；插件 disabled/error 账号不返回凭据但 active 暂停账号仍可用；后端编译、聚焦测试、前端类型检查/测试/构建及 UTF-8、whitespace 检查通过。
