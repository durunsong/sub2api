# v0.2.3 同步验证记录

## 固定输入和范围

- 初始 Fork：main，HEAD `ebd2d484f`，工作树干净。
- 官方旧基线：v0.2.2 / `5485f368b29d05adb95a00f71801c7c23d8f48af`。
- 官方目标：v0.2.3 tag object `fe2b5c04b1c9503fba7e01a099b206f14867cfc1`，commit `8fa67d477d6651a744754392a8982ea589c26ae6`。
- 增量：9 commits / 22 files / +1,517 / -24。仅后端源码和测试、一个新增 SQL；无前端或依赖升级。
- 官方 VERSION 为 0.2.2，Fork 根据 release tag 设为 0.2.3。

## 合入与定制保留

使用 v0.2.2 / Fork HEAD / v0.2.3 逐文件三方合并，不移动分支，不暂存。4 个代码重叠文件无冲突：account_test_service.go、gateway_anthropic_passthrough.go、gateway_count_tokens.go、gateway_upstream_request.go。

对这 4 个文件比较“旧 tag → 旧 Fork”与“新 tag → 新 Fork”的编辑块：删除/添加行完全相同，仅行号移动，原有 Kiro provider、直连分流、credits 与按账号心跳完整保留。其余 17 个 release 文件与 v0.2.3 逐字节一致，版本文件按 Fork 惯例单独设置。

567 个不与 release 重叠的 Fork 文件在合入代码后全部逐字节不变；最终仅其中 3 个版本文档按任务更新（AGENTS.md、FORK_CUSTOMIZATIONS.md、docs/FORK_VS_UPSTREAM.md），其余 564 个保持不变。294 个历史 SQL 逐字节不变，236 新迁移与官方相同。Kiro/XorPay/Access Ban、登录失败自动封禁、提示词审计、订阅重置卡及提示、可用额度筛选、支付幂等、Ops、Claude 别名、GLM 套餐及 UI 定制保留。

## 已完成验证

Go 命令在 backend 执行，pnpm 命令在 frontend 执行，其余命令在项目根执行。Go 1.27.0，Vitest 2.1.9。

| 命令 / 检查 | 结果 |
| --- | --- |
| `go test ./internal/service -run '^TestFetchOpenAIAccountModels' -count=1`（仅先导入新测试） | 预期失败：OAuth/API Key 模型显示名为空；空目录测试通过 |
| `go test ./internal/service -run 'TestFetchOpenAIAccountModels\|OllamaCloud' -count=1`（合入后） | 通过，退出 0 |
| `go test ./... -count=1` | 52 个测试包通过，退出 0；service 131.385 秒 |
| `go test -tags=unit ./... -count=1` | 59 包通过；1 个 schema 包因并发临时目录冲突失败，整体退出 1；该包随后单独复跑通过，service 180.681 秒 |
| `go vet ./... && go build ./...` | 通过，退出 0 |
| `pnpm run test:run --maxWorkers=4 --minWorkers=2` | 277 文件 / 2008 tests 通过，退出 0 |
| `pnpm run typecheck` | 通过，退出 0 |
| `go test -tags=unit ./ent/schema -count=1` | 并发冲突后的单包复跑通过，退出 0，2.506 秒 |
| `sh backend/scripts/resolve-version.sh` | 0.2.3，退出 0 |
| `git diff --check`；新增文件逐个 `git diff --no-index --check /dev/null <file>` | 通过 |
| UTF-8 / BOM / 替换字符 / 冲突标记 | 修改和新增文本文件检查通过，中文文档差异已复核 |
| 范围/版本/暂存区 | 代码仅官方增量，HEAD 未移动，暂存区为空 |

## 环境和执行限制

- 第一次显示名测试命令误在仓库根运行，因 Go module 位于 backend 返回找不到模块；改为 backend 后观察到上述预期行为失败。
- default 与 unit 全量同时运行时，已有 `TestAuthIdentityFoundationSchemas` 的 `load.Config{Path: "."}.Load()` 共用 `.entc` 临时文件。default schema 包通过，unit 报临时 Go 文件不存在；相同代码单独复跑 unit schema 包通过。属于本次并发验证冲突，未修改 schema 或业务代码；以后两套 Go 全量应串行运行。
- 本机未发现 Docker 或 psql，未运行 PostgreSQL integration 测试和真实 236 迁移。已纳入官方覆盖旧列重命名、双列回填/非空保留、缺列补建及重放的 integration 测试，默认测试覆盖迁移源码检查；不等同于真实数据库执行证明。
- 未使用真实支付/OAuth 或远端模型服务联调。前端无源码/依赖变化，运行完整 Vitest 和类型检查，未重复前端 build/lint。
- OpenSpec CLI/CodeGraph 不可用，使用现有 spec-driven 文档结构和 Git/rg，未运行 CLI validate。
- 未执行 git add/commit/push、数据库迁移或部署。上线时 236 会按现有迁移机制执行；需先备份并核对实际白名单及 schema，旧版回退必须与数据库结构匹配。
