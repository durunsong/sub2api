# 升级 v0.2.10 并保留 Fork 定制

## Why
把官方 v0.2.10 功能整合至当前 v0.2.9 Fork，保留现有定制。

## What Changes
官方 v0.2.9 → v0.2.10 共 33 commits、118 files、+3,759/-332。包括 Sonnet 5.5、Claude 原生重置额度查询、风控用户白名单、仪表盘费用趋势，以及流式用量、工具重写、聚合 WS 路由、白名单模型映射和 Antigravity 保活修复。官方 VERSION 仍为 0.2.9，Fork 设为 0.2.10。

## Boundaries
保留 Kiro、XorPay、Access Ban、提示词审计、订阅重置卡、active_available、Ops、品牌别名、GLM 分类、支付 UX 与 VersionBadge。无依赖或迁移变更，不接触秘密/部署配置，不部署、不提交或推送。

## Acceptance
核对全部 118 个路径、42 个 Fork 重叠路径；其余受管理定制源码及 301 个历史 SQL 保持不变。上游新增行为红绿验证，后端默认/unit 回归、build/vet/lint，前端全量测试/typecheck/lint/build；检查最终差异、UTF-8 和冲突标记。
