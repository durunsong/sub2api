# 升级 v0.2.9 并保留 Fork 定制

## Why
将已同步 v0.2.8 的 Fork 升级到官方 v0.2.9，同时保留现有定制功能。

## What Changes
固定官方 release tag v0.2.9（commit `4c00df2e0183e2c70b7fa8ba45914205e36aad0c`）。相对 v0.2.8 共 70 commits、117 files、+3,133/-378，涵盖模型白名单任意位置通配符、协议转换、客户端取消 499、长上下文账号成本、图片价格继承、OpenAI 配额/自动重置/WS、Antigravity 及前端修复。官方 VERSION 仍为 0.2.8，Fork 按标签设为 0.2.9。

## Boundaries
保留 Kiro、XorPay、Access Ban、提示词审计、订阅重置卡、active_available、Ops、用户品牌别名、GLM 分类、支付金额分离及 VersionBadge 禁用在线更新。无新增依赖或数据库迁移。不部署、不访问真实支付或供应商、不提交或推送。三个 deploy/docker-compose 文件遵守部署配置边界，不读取或同步。

## Acceptance
114 个允许路径完成核对；与 Fork 重叠文件保留定制；全部历史 SQL 和非重叠定制字节不变；版本解析输出 0.2.9；相关后端测试、unit 测试、build/vet、前端 Vitest/typecheck/lint/build 通过；完成最终差异与 UTF-8 检查。
