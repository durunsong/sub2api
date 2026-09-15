# 同步官方 v0.2.5 并保留 Fork 定制

## Why
当前 Fork 基于官方 v0.2.4。用户要求纳入 v0.2.5 功能，并保证已有定制继续可用。

## What Changes
仅同步官方 v0.2.4 (`5de5e2bed035d43591a2e10e51f420ef6a84eb98`) 到 v0.2.5 (`86f93c28ee34cc74b629dafb748bd5ac5ca8c5ea`)：197 commits，447 files，+23038/-1538。包含 OpenCode Zen/GO、多协议图像路由、API Key 及订阅批量操作、站点计费模式、配额/认证/网关修复。
版本设为 0.2.5；所有现有 Fork 功能为兼容性约束。

## Capabilities
### Modified Capabilities
- `upstream-release-sync`: 目标发布版本更新，增加新平台及批量操作与定制的兼容验证。

## Impact
后端、前端、两个新增 238 SQL、版本文档和本变更目录。原 Go/前端依赖清单无上游增量，无需安装依赖。
不执行部署、真实迁移、Git 暂存/提交/推送，不读取或修改受保护配置。配置同步差异单独记录。
