## Why

当前 Fork 已同步到官方 `v0.1.183`，并在该基线上继续维护 Kiro、XorPay、Access Ban、提示词审计、订阅重置卡、支付/UI 品牌等定制。官方 `v0.1.184` 已发布，包含 Codex 路由模型目录、公共分组访问控制、原生 compaction 用量统计、智谱团队套餐用量、OpenAI 图像工具冷却和 TTFT 配置等新能力，以及网关、计费、支付、账号和前端修复。

本 Fork 不能用官方整树覆盖。当前 `HEAD` 与官方 `v0.1.184` 的共同祖先仍是官方 `v0.1.182`，直接 merge 会同时重放官方 `v0.1.183` 和 `v0.1.184`，扩大冲突并可能回退 Fork 定制。因此需要只移植官方 `v0.1.183..v0.1.184` 的精确增量，并对所有重叠文件执行逐项兼容合并。

## What Changes

- 以官方 annotated tag `v0.1.184` 指向的提交 `e98ef32eb29aecd30d1def615912ec4dc93173f3` 为唯一目标基线，只处理官方 `v0.1.183` peeled commit `e8cb019fabf8b55199436229044cbf9aa7a82564` 到该提交的增量。
- 合入官方 `v0.1.184` 的全部功能、优化和修复，不带入 tag 之后的 `main` 内容。
- 合入三个官方 `231_*` migration、对应 Ent schema/runtime 变更及前后端契约。
- 保留 `AGENTS.md`、`docs/FORK_VS_UPSTREAM.md` 和 `FORK_CUSTOMIZATIONS.md` 定义的全部 Fork 定制，并保留 `v0.1.183` 之后的本地认证刷新和 Claude 套餐标签修复。
- 对 89 个“官方 184 增量与 Fork 差异重叠”的文件逐项三方核对，不采用整文件的 ours/theirs 覆盖策略。
- 先移植适用的官方回归测试并观察预期失败，再合入对应实现；每个纵向切片完成后运行聚焦回归。
- 将 `backend/cmd/server/VERSION` 更新为 `0.1.184`，并同步 Fork 差异文档和 AI 协作摘要。
- 不新增依赖，不执行数据库迁移、部署、发布、commit 或 push。

## Capabilities

### New Capabilities

- `upstream-release-sync`: 定义官方 release tag 精确同步、Fork 定制保护、迁移兼容、测试先行与发布版本一致性要求。

### Modified Capabilities

- 官方 `v0.1.184` 已有能力按其 release tag 行为更新；本变更不重新定义官方产品语义，只要求完整、可审计地移植。

## Impact

- **后端**：Codex 模型目录与能力同步、OpenAI HTTP/WS/配额/计费、账号与分组访问、用量记录、Anthropic/Grok/Antigravity/Ollama/智谱适配、支付和设置等模块。
- **数据库**：新增 `231_add_usage_log_native_compaction_v2.sql`、`231_add_usage_log_requested_reasoning_effort.sql`、`231_user_restrict_public_groups.sql`；不得修改或删除 Fork 的 `224/225` 重置卡迁移和 Access Ban/Kiro 迁移。
- **前端**：账号、分组、设置、用量、注册、支付和 Codex 模型目录相关 API、类型、组件、页面和 i18n。
- **Fork 高风险区**：Wire 注入、平台常量、账号/分组 handler、gateway 路由、订阅服务、支付页面、VersionBadge、Access Ban 和 Kiro 用户侧 Claude 品牌映射。
- **规模**：官方增量为 170 个提交、342 个文件、20,205 行新增和 1,325 行删除；其中包含 109 个 Go 测试文件和 25 个前端测试文件。
- **运行环境**：实施前必须恢复项目要求的 Go 1.27.0 与兼容 Corepack/pnpm 的 Node 运行时；当前默认 shell 中 Go 不在 PATH，Node 16 无法运行当前 Corepack。

## Non-Goals

- 不同步官方 `main` 上 `v0.1.184` tag 之后的提交。
- 不删除、简化或重写 Fork 定制。
- 不借版本同步进行无关重构、依赖升级、格式化或产品行为调整。
- 不执行生产数据库迁移、容器重建、部署或发布。
- 不自动创建 Git commit、push 或 PR。
