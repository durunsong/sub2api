## Context

本 Fork 当前 `HEAD` 为 `9f34e7850`，版本为 `0.1.183`，工作区在规划开始时干净。官方 release tags 已按只读调查获取为本地 `upstream-v0.1.183` 和 `upstream-v0.1.184`：

- `v0.1.183` tag object `c21fd3382a1c39fe491a96ac6780bac927327ae4`，peeled commit `e8cb019fabf8b55199436229044cbf9aa7a82564`。
- `v0.1.184` tag object `cbfc2922ce353adc39142a1ebf9d6d54ad06ca4d`，peeled commit `e98ef32eb29aecd30d1def615912ec4dc93173f3`。
- 当前 Fork 与官方 184 的 merge base 为 `5a7d469622911a6b1291a692376df5fa03f9ac2e`，即 Fork 已合入的官方 182 水位。
- 官方 183 到 184 增量为 170 commits、342 files、20,205 insertions、1,325 deletions。
- 官方增量与 Fork 相对官方 183 的差异有 89 个重叠文件。

Fork 已经通过手工整合拥有官方 183 行为，但官方 183 commit 不是当前 HEAD 的祖先。因此 Git 的普通 merge 无法把 183 当作已经同步的基线。

## Goals / Non-Goals

**Goals:**

- 精确、完整地移植官方 184 release。
- 保留全部 Fork 定制和 183 后本地修复。
- 让每个高风险纵向切片都有先失败后通过的回归证据。
- 保留迁移、外部 API、计费、状态机和 UI 契约兼容性。
- 形成可重复检查的上游范围、冲突处理和验证证据。

**Non-Goals:**

- 不同步 release tag 之后的官方 main。
- 不清理历史 Fork 结构或重新设计官方 184 功能。
- 不新增依赖、不升级工具链版本声明。
- 不执行生产数据操作、部署、发布或 Git 提交。

## Decisions

### 1. 使用官方 tag-to-tag delta，不直接 merge v0.1.184

权威输入为：

```sh
git diff --binary upstream-v0.1.183..upstream-v0.1.184
```

实施按目录和功能切片应用该差异。直接 `git merge upstream-v0.1.184` 被排除，因为 merge base 是官方 182，会扩大为 182→184 并把已手工同步的 183 再次纳入三方合并。

逐 commit cherry-pick 也被排除：170 个提交包含大量 merge commits，容易因依赖顺序、重复修复和自动 commit 扩大风险。

### 2. 用三份事实解决重叠文件

每个重叠文件同时读取：

1. 官方 183 基线：`git show upstream-v0.1.183:<path>`。
2. 当前 Fork：工作区 `<path>`。
3. 官方 184 目标：`git show upstream-v0.1.184:<path>`。

先识别官方 delta 的行为意图，再将该意图合入当前 Fork。禁止用官方 184 整文件覆盖当前 Fork，也禁止只保留 Fork 而跳过官方增量。无法同时满足时必须暂停并更新 OpenSpec，不得猜测业务结果。

### 3. 按纵向行为切片实施并使用官方测试作为首选验收

官方增量包含 109 个 Go 测试文件和 25 个前端测试文件。每个切片按以下顺序：

1. 移植该切片新增或修改的官方测试及必要测试工具。
2. 在当前 Fork 实现未移植时运行聚焦测试，确认因缺少 184 行为而失败；仅缺少新类型导致无法编译时，先加入最小接口/fixture，使测试到达行为失败。
3. 移植官方实现，并在重叠文件中保留 Fork 行为。
4. 运行聚焦测试直到通过。
5. 运行受影响 Fork 定制测试。

切片顺序：数据库/契约 → Codex/账号/分组 → OpenAI 传输与计费 → 其他 provider/适配器 → 支付/订阅/用量 → 前端 → Fork 红线总回归 → 版本与文档。

### 4. migration 使用现有同号多文件规则

官方 184 新增：

```text
backend/migrations/231_add_usage_log_native_compaction_v2.sql
backend/migrations/231_add_usage_log_requested_reasoning_effort.sql
backend/migrations/231_user_restrict_public_groups.sql
```

仓库已经允许同号、不同完整文件名的 migration 共存，因此保留官方文件名，不重编号。同步 Ent 生成文件时只接受官方 schema 所需差异，并复核 Fork schema 仍包含 Kiro、Access Ban 和订阅重置卡字段。

### 5. Fork 红线采用显式检查矩阵

| 范围 | 必须保留的事实 | 主要检查 |
| --- | --- | --- |
| Kiro | OAuth、调度、配额、模型、缓存、Web Search、用户侧 Claude 品牌 | Kiro tests、Wire tests、平台目录/Composite tests |
| XorPay | provider、webhook、退款、前端支付宝展示 | payment provider/service/frontend tests |
| Access Ban | ip/ua/ip_ua/email_suffix、网关/auth 中间件、管理 UI | accessban、IPBanService、route tests |
| 订阅重置卡 | 224/225、来源幂等、期限快照、过期重开、退款作废、旧接口映射 | reset card/service/repository tests |
| 提示词审计 | 独立模块、迁移、设置、管理控制台 | securityaudit tests 与路由测试 |
| UI/支付 | VersionBadge、GLM 命名、倍率隐藏、金额分离、Select/ConfirmDialog | 对应 Vitest |
| 本地后续修复 | transient refresh 保持会话、Claude 套餐标签一致 | client/payment view tests |

### 6. 工具链问题是环境门禁，不通过绕过验证处理

`backend/go.mod` 要求 Go 1.27.0。规划时默认 shell 中 `go` 不在 PATH；Node 为 16.20.2，当前 Corepack 使用 `URL.canParse` 而失败。实施开始前先定位项目现有 Go/Node 运行时或使用用户已配置的版本管理器。不得修改 `go.mod`、lockfile 或依赖版本来迎合错误的本地运行时。

## Error Handling

- 上游 tag SHA 不匹配：停止，不继续同步。
- 工作区出现用户新改动：保留并重新计算重叠范围；只有无法兼容时询问用户。
- 测试在红阶段因环境失败：先修复运行时入口，不把环境错误当作预期红。
- 官方行为与 Fork 金额、状态机、迁移或权限契约冲突：停止并更新 design/spec，由用户确认互斥结果。
- 应用补丁产生 `.rej` 或冲突标记：不得继续到下一切片；逐文件解决并删除临时产物。
- 集成测试依赖 PostgreSQL/Redis 而本地不可用：报告限制，运行可用的 unit/SQL contract tests，并不得宣称完整集成验证通过。

## Verification Strategy

聚焦测试按实际改动路径选择。最终门禁至少包括：

```sh
cd backend && go test ./... -count=1
cd backend && go vet ./...
cd backend && go build ./...
pnpm --dir frontend run test:run
pnpm --dir frontend run lint:check
pnpm --dir frontend run typecheck
pnpm --dir frontend run build
```

补充静态检查：

```sh
git diff --check
rg -n '^(<<<<<<<|=======|>>>>>>>)' -g '!node_modules/**' -g '!dist/**'
find . -name '*.rej' -o -name '*.orig'
```

迁移和 Fork 红线使用聚焦测试与文件/符号断言；中文文本检查 UTF-8、BOM、`U+FFFD` 替换字符。所有完成结论使用最终编辑后的新鲜输出。

## Rollback

本任务不执行生产迁移或部署。工作区修改通过 Git diff 可审查，且不自动 commit。若某切片无法完成，停止在该切片并报告已修改文件；不得使用 hard reset、clean 或递归删除回滚用户工作。
