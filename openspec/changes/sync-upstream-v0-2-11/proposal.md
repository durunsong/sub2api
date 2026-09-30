# 升级 v0.2.11 并保留 Fork 定制

## Why
将用户指定的官方 v0.2.11 完整功能增量整合至当前 v0.2.10 Fork，保留既有自定义能力。

## What Changes
官方增量 23 commits / 90 files / +5,146 / -296：余额在途预留、API Key 创建限制、Claude 原生额度重置、GPT-6.1 Sol、Astra Ultrafast、Codex 远程模型目录/订阅识别、Claude Code only 兼容入口降级。同步 89 个非部署路径；部署示例保持原样，新增参数使用源码默认值。无依赖或数据库迁移变化。

## Boundaries
保留 Kiro、XorPay、Access Ban/自动登录封禁、提示词审计、永久订阅重置卡/退款幂等/active_available、Ops、Claude 品牌别名、GLM 分类、支付 UX、VersionBadge。仅本地源码升级，不部署、不执行迁移、不暂存/提交/推送。

## Acceptance
逐文件核对 89 个增量路径和 24 个 Fork 重叠路径；非重叠定制和 301 个历史 SQL 不变。先运行新增行为测试观察预期失败，合入后通过后端默认/unit 测试、build/vet/lint 与前端测试/typecheck/lint/build；检查版本、范围、UTF-8/BOM/乱码及冲突标记。
