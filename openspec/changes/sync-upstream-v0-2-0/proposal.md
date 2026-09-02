## Why

当前 Fork 已同步到官方 `v0.1.185`，并在该基线上维护 Kiro、XorPay、Access Ban、提示词安全审计、订阅重置卡、订阅可用额度筛选、Ops 和 UI/支付品牌等定制。官方 `v0.2.0` 已发布，新增 OpenAI Fast 分组策略、按模型配置 reasoning effort、Kimi 原生 Responses 转发、Claude Fable 5.1 和无 call ID 的自动化启动请求支持，同时包含 WebSocket、调度、缓存身份和 Anthropic beta 字段处理修复。

本 Fork 不能直接合并或用官方整树覆盖。当前 `HEAD` 与官方 `v0.2.0` 的共同祖先仍是 `5a7d469622911a6b1291a692376df5fa03f9ac2e`，普通 merge 会重复引入已手工同步的版本并扩大冲突。官方 `v0.1.185..v0.2.0` 增量修改 148 个文件，其中 54 个也被 Fork 修改，因此必须按官方 tag-to-tag delta 逐行为整合。

## What Changes

- 以官方 annotated tag `v0.2.0` peeled commit `aa236488351eb71e120fc2b6fb32e36b0374c918` 为目标，只处理 `upstream-v0.1.185..upstream-v0.2.0` 的 60 个提交。
- 完整合入该范围内 148 个文件的功能、优化和修复，不带入 tag 之后的官方 `main`。
- 合入四个官方 migration：`232_channel_cache_write_1h_pricing.sql`、`232_group_force_openai_fast.sql`、`232_group_reasoning_effort_over_limit.sql`、`233_group_free_openai_fast.sql`，以及对应 Ent、DTO、仓储、服务和前端契约。
- 保留 `AGENTS.md`、`docs/FORK_VS_UPSTREAM.md` 和 `FORK_CUSTOMIZATIONS.md` 记录的全部 Fork 定制；54 个重叠文件必须按官方 185、当前 Fork、官方 2.0 三份事实逐项核对。
- 采用纵向 TDD：优先移植 43 个官方 Go 测试文件和 10 个前端 spec 中与各切片相关的测试，观察缺失行为导致的预期失败后再移植最小实现。
- 将 `backend/cmd/server/VERSION` 设为 `0.2.0`。官方 tag 内该文件仍为 `0.1.185`，Fork 延续按 release tag 标记运行版本的既有约定。
- 更新 Fork 差异文档和协作摘要；不新增依赖，不执行生产迁移、部署、commit、push 或 PR。

## Capabilities

### New Capabilities

- `upstream-release-sync`: 定义官方 release tag 精确同步、Fork 定制保护、迁移兼容、纵向 TDD 和版本一致性要求。

### Modified Capabilities

- 官方 `v0.2.0` 已有能力按 release tag 的源码和测试契约更新；本变更不重新设计官方产品语义。

## Impact

- **数据库与实体**：Group 新增 `force_openai_fast`、`free_openai_fast`、`max_reasoning_effort_over_limit`，渠道定价新增 `cache_write_1h_price`。
- **OpenAI/Kimi 网关**：Fast 强制与免费计费策略、按模型 reasoning effort 映射及超限拒绝/降级、Kimi 原生 Responses、自动化 bootstrap、WebSocket terminal event 判定和 API Key 缓存身份。
- **Anthropic/模型**：server-side-fallback beta 字段清理、模型级冷却错误码修复、Claude Fable 5.1。
- **前端**：分组 Fast/reasoning effort 配置、渠道一小时缓存写价格、模型广场价格和相关 i18n。
- **Fork 高风险区**：Group/DTO/缓存/计费/gateway 的 54 个重叠文件，以及 Kiro、XorPay、Access Ban、订阅重置卡、提示词审计和品牌覆盖测试。

## Non-Goals

- 不同步官方 `v0.2.0` tag 之后的 `main`。
- 不删除、弱化、重写或重新命名 Fork 定制。
- 不借同步进行无关重构、依赖升级、批量格式化或配置清理。
- 不读取或修改 `.env*`、密钥、本地配置、依赖目录和生成产物。
- 不执行数据库迁移、容器发布、生产操作或任何 Git 提交/推送。
