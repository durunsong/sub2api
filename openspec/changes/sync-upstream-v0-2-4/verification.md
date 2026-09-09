# v0.2.4 同步验证记录

## 固定输入与范围

- 初始 Fork：main / HEAD `0e5fd73f6`，工作树干净。
- 官方旧基线：v0.2.3 / `8fa67d477d6651a744754392a8982ea589c26ae6`。
- 官方目标：v0.2.4 tag object `d681d0798064ee0ffff376d19687d12f09fe600f`，commit `5de5e2bed035d43591a2e10e51f420ef6a84eb98`。
- 官方增量：70 commits / 266 files / +5,696 / -726。只合入此 tag，无后续 main 内容。
- 官方 VERSION 仍为 0.2.3，Fork 设为 0.2.4。
- 修改范围：后端/前端 release 增量与兼容测试、VERSION、三个 Fork 说明文档，以及本 OpenSpec 目录。

## 合并和定制保留

先在临时目录执行逐文件三方预演，再写入当前工作树；不移动 HEAD、不暂存。80 个文件需要合并 Fork 内容，25 个有文本冲突。除受保护文件外，261 个 release 路径已纳入，其中 179 个与官方目标逐字节相同，82 个有版本/Fork/测试兼容差异。

491 个不与 release 重叠且允许读取的 Fork 路径中，最终 486 个逐字节不变；其余仅三个版本文档和两项加强后的回归测试（Kiro Badge、监控竞态/排行）。另一个旧上游监控测试调整为按目录验证数量并明确包含 MiniMax。295 个历史 SQL 全部逐字节不变。

- Kiro 与 MiniMax 同时保留在账号/分组类型、共享目录、配额、Composite、渠道定价匹配与调度快照；Wire 的 Kiro/Grok 注入和 Access Ban 路由保留。
- 237 的 quota/Composite CHECK 补回 Kiro，监控 provider 按官方范围增加 MiniMax；不改历史迁移及任何数据。
- Kiro credits/cache_read、独立冷却、真实上游请求 ID、XorPay、Access Ban/登录自动封禁、提示词审计、订阅重置卡/来源幂等/退款事务、active_available、Ops 删除、支付/GLM/VersionBadge/UI 定制保留。
- 管理端 Kiro Badge 显式保留真实名称，用户端 Claude 别名保持原规则。
- 账号批量 Token 刷新保留确认弹窗，并纳入官方“失败账号继续选中”；监控页保留独立取消/序号保护，同时支持隐藏用户排行。
- 插件安装关闭逻辑统一为一次且检查错误；关闭在两次 rename 之前执行，失败时使用原有临时/最终产物清理规则，安装和数据库原件恢复共用入口。

## 已执行检查

Go 命令在 backend，pnpm 命令在 frontend，其余在项目根执行。环境：Go 1.27.0、pnpm 9.15.9、Vitest 2.1.9。

| 命令/检查 | 结果 |
| --- | --- |
| `go test -tags=unit ./internal/service -run TestGetUserGroupVisibility -count=1`，先导入官方测试 | 预期失败：有效订阅分组缺失、订阅仓库错误未传播；合入后定向通过 |
| `go test ./migrations -run TestMiniMaxPlatformMigration -count=1` | Kiro 断言对原上游 237 预期失败；补回 Kiro 后通过 |
| `pnpm exec vitest run src/components/common/__tests__/PlatformTypeBadge.spec.ts` | 新增回归先失败（ClaudeOAuth），修复后与 Grok/平台别名测试共 13 项通过 |
| `go test ./internal/service -run TestPluginPackageInstaller -count=1` | 修复重复关闭后通过；全量原先四项安装测试因 file already closed 失败 |
| `go test ./internal/handler/admin -run TestGrokMediaEligibility -count=1` | 通过 |
| `pnpm exec vitest run src/views/admin/__tests__/AccountsView.selectAllResults.spec.ts src/views/admin/__tests__/ChannelMonitorView.grok.spec.ts src/views/user/__tests__/ChannelStatusV2View.race.spec.ts` | 3 文件 / 12 项通过 |
| `pnpm run test:run --maxWorkers=4 --minWorkers=2` | 最终 282 文件 / 2084 项通过，退出 0 |
| `pnpm run typecheck` | 最终通过，退出 0 |
| `pnpm run lint:check` | 最终通过，退出 0 |
| `pnpm run build` | i18n 门禁、vue-tsc、Vite 构建通过，退出 0；存在大 chunk 提示 |
| `go vet ./... && go build ./...` | 插件最终修复后通过，退出 0 |
| `bash deploy/tests/apple-container-test.sh` | 通过，使用 fixture container 和临时目录，无真实容器操作 |
| `sh backend/scripts/resolve-version.sh` | 0.2.4 |

`go test ./... -count=1` 最终通过，退出 0，52 个测试包，service 包 123.652 秒。随后 `go test -tags=unit ./... -count=1` 全量通过，退出 0，60 个测试包，service 包 168.601 秒；两轮均无失败。

## 失败诊断与复核

初次全量 Go 的新 Grok 测试未包含 Fork 多出的 Kiro 构造参数，已适配。随后发现插件安装的双方关闭逻辑叠加；4 项现有安装测试证明回归并验证修复。修复前已启动的 unit 轮也出现相同四项失败，最终重新串行执行两套全量。

初次前端全量 8 项失败：上游刷新测试调用原生 confirm、旧监控测试硬编码平台数量、Fork 竞态测试缺少新 flag mock。均按实际组件契约适配，保留原行为断言；增加确认前不刷新、MiniMax 显示、隐藏排行不请求的检查。最终 2084 项通过。

独立只读审查覆盖 80 个重叠文件的旧 Fork 增删差异及关键链路；发现 Kiro 管理端标签回归并复核修复，另对插件关闭、清理与共享调用路径单独复核通过。

## 文件与编码检查

最终差异共 270 个修改/新增路径（含两项官方图片资源），268 个文本文件为 UTF-8，无 BOM、替换字符、冲突标记或新增个人绝对路径。`git diff --check`、新增文件逐个 `git diff --no-index --check /dev/null <file>` 和改动 Go 文件 `gofmt -l` 均无诊断；no-index 对存在差异的文件可返回 1，因此同时检查输出为空。HEAD 保持初始值，暂存区为空。历史 SQL/非重叠定制按 Git blob 与当前内容逐字节核对。

## 执行边界与剩余限制

- 受保护文件 `deploy/.env.example` 和 `deploy/docker-compose.dev.yml`、`deploy/docker-compose.local.yml`、`deploy/docker-compose.standalone.yml`、`deploy/docker-compose.yml` 本次不读不写。它们属于官方 266 路径中的 5 个配置路径，未宣称整树逐字节等于官方。
- 新日志保留采用应用默认 30 天，后台运行时日志设置可调整，访问日志持久化默认关闭；Apple 可选子网支持已合入脚本/文档，未迁移现有网络。
- 本机无 Docker/psql/postgres 命令，未运行 PostgreSQL integration 或真实 237 迁移；迁移文件测试不能替代实际数据库验证。上线须先备份并验证已有 Kiro 数据及 CHECK；旧版本回滚需要与已有 MiniMax 数据兼容。
- 未使用真实 OAuth、支付或远端模型服务联调；未做 Windows 实机验证。插件 Windows 兼容性由句柄生命周期审查和本机安装测试支持。
- 构建的大 chunk 提示未阻断产物生成，本次不扩大到打包重构。
- OpenSpec CLI 和 CodeGraph 不可用，使用现有 spec-driven 目录与 Git/rg；未执行 CLI validate。
- 未执行 git add/commit/push、真实数据库迁移或部署。
