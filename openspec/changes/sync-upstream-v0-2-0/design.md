## Context

规划开始时仓库位于 `main`，`HEAD` 为 `313beda6f78a191dbbd826096a4758aebaeb7582`，工作树干净。官方 tag 已作为本地审计引用获取：

- `upstream-v0.1.185` peeled commit：`2ac784c51a5d0925b324efef2ba6b3446c364781`。
- `upstream-v0.2.0` tag object：`dd07c4d8d484878e617c945cc8bacc304a5a6560`。
- `upstream-v0.2.0` peeled commit：`aa236488351eb71e120fc2b6fb32e36b0374c918`。
- 官方增量：60 commits、148 files、4,926 insertions、543 deletions。
- 官方增量包含 43 个 Go 测试文件和 10 个前端 `.spec.ts` 文件。
- 官方增量与 Fork 相对官方 185 的差异有 54 个重叠文件。

官方 `v0.2.0` tag 内 `backend/cmd/server/VERSION` 仍为 `0.1.185`，但 release tag 和发布说明均为 `v0.2.0`。本 Fork 过去始终按 release tag 修正 VERSION，因此目标版本为 `0.2.0`。

## Goals / Non-Goals

**Goals:**

- 精确、完整地移植官方 `v0.1.185..v0.2.0` release 增量。
- 保留当前 `main` 的全部 Fork 定制和后续本地修复。
- 对数据库、计费、权限、调度、网关和外部 API 契约保留可审计的回归证据。
- 让每个新增官方行为先由官方测试或等价行为测试产生 RED，再移植最小实现变为 GREEN。
- 形成可以重复执行的范围、冲突、红线和最终验证命令。

**Non-Goals:**

- 不同步 `v0.2.0` 之后的官方 main。
- 不重构 Fork 架构，不新增依赖，不修改工具链声明。
- 不执行生产迁移、部署、发布、commit 或 push。

## Decisions

### 1. 使用 tag-to-tag delta，不直接 merge

权威输入为：

```sh
git diff --binary upstream-v0.1.185..upstream-v0.2.0
```

普通 `git merge upstream-v0.2.0` 被排除，因为共同祖先早于本 Fork 已手工同步的官方版本。整树覆盖也被排除，因为 Fork 相对官方 185 有 550 个文件差异。逐 commit cherry-pick 被排除，因为 60 个提交包含 merge 和相互依赖的生成代码，增加重复应用和自动提交风险。

### 2. 先处理非重叠文件，再逐项处理 54 个重叠文件

官方新增文件及只被官方修改的文件可按 tag delta 移植。每个重叠文件同时使用：

1. `git show upstream-v0.1.185:<path>`：官方旧行为。
2. 工作区 `<path>`：当前 Fork 行为。
3. `git show upstream-v0.2.0:<path>`：官方目标行为。

先识别官方 hunk 的行为意图，再将该意图合入 Fork。禁止整文件采用 ours/theirs。金额、状态机、权限、迁移或外部契约出现互斥语义时停止并更新 OpenSpec，不凭猜测决策。

### 3. 按纵向行为切片应用 TDD

切片顺序为：

1. migration、Ent 与 Group/Channel 契约。
2. reasoning effort 按模型匹配和超限拒绝/降级。
3. OpenAI Fast 强制策略与免费 Fast 计费。
4. Kimi Responses、Codex 自动化 bootstrap、API Key cache identity 和调度快照。
5. WebSocket terminal event、模型冷却 404、Anthropic fallback beta 和 Claude Fable 5.1。
6. 前端 Group/Channel/模型广场 UI 与 i18n。
7. Fork 红线回归、版本和文档。

每个切片先移植官方测试或添加等价测试，运行并确认失败原因是缺少 2.0 行为；随后移植最小实现并运行同一测试转绿。官方生成的 Ent 文件与 schema/migration 必须同一切片更新，避免半生成状态。

### 4. migration 追加且同号共存

四个官方 migration 使用原文件名追加，不改写已发布 migration，不重编号。仓库已有同号不同文件名并存规则，因此 `232_*` 可确定性共存。重点复核：

- `cache_write_1h_price` 使用 `NUMERIC(20,12)` 且 NULL 保留旧价格行为。
- 三个 Group 字段有官方默认值：两个 Fast 开关为 `FALSE`，超限策略为 `downgrade`。
- Fork 的 `157` 仍含 `kiro` 与 `grok`。
- `159`/`160` Access Ban、`224`/`225` 订阅重置卡 migration 仍存在且内容不回退。

### 5. Fork 定制使用显式保护矩阵

| 范围 | 必须保留的行为 | 验证入口 |
| --- | --- | --- |
| Kiro | OAuth、调度、配额、模型、缓存模拟、Web Search、独立 429、`kiro_credits`；用户侧显示 Claude | Kiro service/handler tests、Wire tests、平台目录与前端品牌 tests |
| XorPay | 创建、查询、webhook、退款、支付宝展示 | payment provider/service/frontend tests |
| Access Ban | ip/ua/ip_ua/email_suffix、注册邮箱格式、auth/gateway 中间件、管理页 | accessban、IPBanService、route、auth tests |
| 提示词审计 | 扫描、策略、事件、设置、管理控制台及迁移 | securityaudit tests 与 API route tests |
| 订阅重置卡 | 期限快照、来源幂等、过期重开、消费清零窗口、退款作废、旧接口映射和兼容镜像 | reset-card、subscription、payment、redeem tests |
| 订阅筛选 | `active_available` 默认值和所有当前额度窗口过滤 | repository/service/frontend filter tests |
| UI/支付/Ops | VersionBadge 无在线更新、GLM 标签、倍率隐藏、金额分离、Select/ConfirmDialog、错误日志删除 | 对应 Vitest 和 Ops tests |

OpenAI Fast 和 reasoning effort 的 Group/DTO/cache 变更必须与 Kiro 可参与的 Composite、平台目录和调度快照共存。官方定价字段不得改变 Kiro cache_read 不重复计入 input tokens 的计费口径。

### 6. 版本与文档按 Fork 约定更新

将运行版本设置为 `0.2.0`，并更新 `FORK_CUSTOMIZATIONS.md`、`docs/FORK_VS_UPSTREAM.md`、`AGENTS.md`。文档记录官方 tag SHA、增量统计、四个 migration、官方能力和定制保留结论。若最终整合导致统计变化，仅更新 Fork 差异统计，不修改已核验的官方 tag-to-tag 统计。

## Error Handling

- tag SHA 或差异统计不一致：停止，不继续同步。
- 工作区出现新的用户改动：保留并重新计算重叠范围；只有无法兼容时询问用户。
- RED 阶段因运行时、依赖或测试夹具失败：先恢复现有项目验证入口，不把环境错误当作行为 RED。
- 金额、状态机、权限、迁移或外部接口发生互斥结果：停止并请求用户决定。
- 出现冲突标记、`.rej`、`.orig` 或意外删除：在当前切片解决后再继续。
- PostgreSQL/Redis/Docker 不可用：运行可用 unit/contract tests，准确报告未覆盖的 integration 风险。

## Verification Strategy

聚焦测试按每个切片的官方新增/修改测试文件和相邻 Fork 测试选择。最终编辑后的门禁为：

```sh
cd backend && go test ./... -count=1
cd backend && go vet ./...
cd backend && go build ./...
pnpm --dir frontend run test:run
pnpm --dir frontend run lint:check
pnpm --dir frontend run typecheck
pnpm --dir frontend run build
```

补充检查：

```sh
git diff --check
rg -n '^(<<<<<<<|=======|>>>>>>>)' -g '!frontend/node_modules/**' -g '!frontend/dist/**'
find . -name '*.rej' -o -name '*.orig'
```

另外检查 VERSION、四个 migration、Fork 红线符号、54 个重叠文件处置、UTF-8、BOM 和 `U+FFFD`。不把未运行或受环境限制的检查报告为通过。

## Rollback

本变更不执行生产操作。所有工作区变更保持未提交并可通过 Git diff 审阅。若切片无法完成，保留现场、报告已修改文件和失败证据；不使用 reset、clean、checkout 或递归删除回滚用户工作。
