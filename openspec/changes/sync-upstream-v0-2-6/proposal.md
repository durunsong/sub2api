# 同步官方 v0.2.6 并保留 Fork 定制

## Why
用户要求将当前 v0.2.5 Fork 更新至官方 v0.2.6，已有定制必须继续可用。

## What Changes
仅同步官方 v0.2.5 `86f93c28ee34cc74b629dafb748bd5ac5ca8c5ea` 到 v0.2.6 `49a39b6dc1abed30fd227611e8af1108bc427610`，共 60 commits / 132 files / +4836 / -304。纳入 Codex 票据采集/注入与后台热设置、兑换历史分页、Gemini 混合模型发现、DeepSeek 工具媒体兼容、Chat developer 角色兼容、Antigravity attribution 处理、分组统计优化和前端修复；VERSION 设为 0.2.6。

## Capabilities
### Modified Capabilities
- `upstream-release-sync`: 更新目标发布并保证全部 Fork 功能兼容。

## Impact
后端、前端、既有 Go 依赖安全升级及版本文档。无新增 SQL，前端依赖不变。
不增加独立业务功能，不执行部署、真实迁移、Git 暂存/提交/推送，不读写受保护配置。
