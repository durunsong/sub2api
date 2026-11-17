# v0.2.22 同步验证记录

日期：2026-11-17。首次交付为工作区源码；用户随后授权复审通过后提交并推送，复审记录见文末。生产部署和真实数据库迁移不在本次范围内。

## 来源与范围

- 初始 Fork HEAD：`e9c30cb40`，开始时工作区干净。
- 官方基线：`v0.2.21` / `10ba8f9b4dc2e6023d4a9879ab77b21841babd10`。
- 官方目标：`v0.2.22` / `b0fa9f3e6eeab1968a53e1289b69daa4bb40ca2b`；tag object `b11573c9d55d39b6780e1e9f6e193ba2e6ca2321`。
- 发布增量 187 commits、340 路径、+20,708/-1,652；仅同步指定标签，不引入后续 main。
- 三方比对：243 个无 Fork 差异路径、79 个自动合并路径、16 个冲突、1 个 VERSION 已一致路径、1 个排除的部署示例配置。合入 339 个发布路径；官方 VERSION 仍是 0.2.21，Fork 明确设为 0.2.22。
- 保留 `deploy/config.example.yaml`；三个 Dockerfile 及 CI 的 Go 声明同步 1.27.4，保留 Fork 镜像源和安装策略。jsonschema-go 0.4.2 仅由间接改为直接引用，前端依赖/锁文件无变化；Wire 生成补齐已有工具依赖 google/subcommands v1.2.0 的两行 go.sum 校验和，不新增应用依赖。

## 冲突与适配

- Ent 三个冲突文件基于合并后的 schema 重新生成，保留 Kiro 字段并插入 force_openai_ultrafast；Wire 重新生成，同时接入新 UsageBillingLaneCombiner、Kiro、封禁与邮件 Worker 清理。
- 网关继续使用 Fork 的默认等待 ping/切号预算和 BindStickySessionForGroup，新断连写回拥有 5 秒超时。两个 Anthropic 兼容桥接保留 Kiro 直连、真实响应头、代理失败拒绝直连和 credits，同时支持上游 OAuth token 轮换恢复。
- 账号成本保留 ctx 与冻结 pricingAt，新增 Vertex 默认模型价倍数；网关保留 Kiro opts。鉴权缓存继续使用完整 groupForSnapshot，新增 Ultra Fast 字段不遗漏 Kiro。
- 账号用量加载保留 Kiro 元数据同步并接入卸载/换行保护；Kiro API Key 空 base_url 规则与透传白名单开关并存。
- 端点图表支持管理员账号成本，同时保留用户端标准成本/账号成本隐藏及用户费用标签。账号用量测试保留 Kiro credits 与上游快照新鲜度测试。
- 适配上游新增测试与 Fork 签名：Claude scoped usage 移除多传的 nil；两个 handler 用例补齐 Kiro provider/cooldown 参数；DeepSeek 成本用例保留 ctx，Gemini 计费用例补齐 Fork opts。不改生产依赖注入。
- 新批量扣费复用原有订阅更新 SQL，合计 USD 和原始 token，包含零金额请求，仅累计已认领请求，按订阅 ID 排序写入。保留全期额度、8 位金额量化和幂等；不会恢复第 30/60 天补额。

## 红绿证据

- 先接入上游 TestValidateConcurrencyPingInterval，旧实现对 interval=0 返回错误；合入后定向测试通过。
- 批量订阅回归先在仅适配函数签名、token=0 的上游逻辑上运行：付费请求预期 210 token 实际 0，免费请求无更新。修复后同一测试通过，覆盖混合 token 桶、不同订阅、免费请求与已认领过滤。
- 实际数据库回归额外验证 90 天订阅第 40 天累计用量、冻结周期起点与到期时间以及同请求重放；结果见下表。

## 已执行验证

Go 命令在 `backend/` 执行，pnpm 在 `frontend/` 执行。Go 1.27.4 / pnpm 10.33.2。

| 命令 | 结果 |
| --- | --- |
| `go test -p 2 -tags=unit ./... -count=1` | 最终重跑退出 0；64 个测试包通过、52 个包无测试；service 236.887s |
| `go generate ./ent` / `GOFLAGS=-p=2 go generate ./cmd/server` | 退出 0，ORM 与依赖注入重新生成 |
| `go test -p 2 -tags=unit ./internal/repository ./internal/config -run 'TestUsageBillingBatchSubscriptionsPreserveTokensAndTermQuota\|TestValidateConcurrencyPingInterval\|TestClaudeUsageFetcher_ScopedFableRoundTrip' -count=1` | 退出 0；订阅兼容、并发 ping 和 Claude scoped usage 通过 |
| `go test -p 2 -tags=integration ./internal/repository -run 'TestUsageBillingRepositoryApplyBatch\|TestUsageBillingBatchForkTermQuotaAndTokens\|TestGroupUsageRollup' -count=1 -v` | 退出 0；23 个顶层测试通过，隔离 PostgreSQL/Redis 容器，含新迁移、事务/幂等、全期额度与汇总并发行为 |
| `go build -p 2 ./... && go vet -p 2 -tags=unit ./... && go build -p 2 -tags embed ./...` | 最后源码/测试编辑后运行，退出 0 |
| `go test -race -p 2 -tags=unit ./internal/service ./internal/handler -run 'TestUsageBillingLane_\|TestSchedulerFallback\|TestGatewayMessagesStickyWritebackAfterForward' -count=1` | 退出 0；service 7.883s、handler 6.046s，无竞争报告 |
| `pnpm exec vitest run` | 退出 0；412 文件 / 3,386 测试通过 |
| `pnpm exec vitest run src/components/charts/__tests__/EndpointDistributionChart.spec.ts` | 退出 0；新增 2 项费用可见性回归通过 |
| `pnpm lint:check && pnpm typecheck && pnpm build` | 退出 0；应用构建通过 |
| `pnpm exec eslint src/components/charts/__tests__/EndpointDistributionChart.spec.ts && pnpm typecheck` | 新测试最后编辑后运行，退出 0 |
| `go test -p 2 -tags=unit ./internal/service -run '^$' -bench 'BenchmarkEnforceCacheControlLimit\|BenchmarkRequestToolHelpers' -benchmem -benchtime=100ms -count=1` | 退出 0；6 个基准变体通过 |

最终全量后端测试已通过。首轮全量测试发现新增上游测试未适配 Fork 函数签名，已修正参数并以最终重跑为准。一次测试命令误在仓库根目录执行，未运行测试，后在 backend/ 正确执行。

基准测量为约 28 KB 固定请求的本机单轮微基准：常见缓存控制 legacy/optimized 为 133,647/25,304 ns/op、689/0 B/op；工具 helper 为 88,864/59,863 ns/op、103,525/69,023 B/op。B/op 是每次操作累计分配，不是 RSS/保留堆；不据此承诺生产吞吐。前端构建仅有 Browserslist 数据与大 chunk 非阻塞提示。

## 定制与数据保护

- 318 份历史 SQL 与初始 HEAD 逐字节一致；新增 `251_group_force_openai_ultrafast.sql`、`251_group_usage_rollup_nonblocking_build.sql`，与 Fork 的订阅重置历史 251 按完整文件名共存。
- Fork 独有业务文件未替换；Kiro/XorPay/Access Ban、全期额度、永久卡、重置记录、邮件券、批量兑换、微信客服、CC Switch 品牌、密钥模型、订单筛选与缓存命中率均保留。
- 独立静态审查覆盖 Kiro 直连/桥接、credits/TTL/代理/粘性、鉴权与封禁、用户费用隐藏、调度和定价交叉，无可行动发现；批量扣费补丁二次复核无可行动发现，Ent 58 个 Group runtime descriptor 下标均与 schema 对齐。
- 生产服务、供应商、支付、邮件和真实业务数据库未连接；数据库验证仅使用隔离测试容器。

## 最终静态验证

- 最终变更 351 路径，包含 339 个发布路径、3 份 Fork 文档、3 个兼容测试、5 份 OpenSpec 文档及 Wire 工具校验和补全。
- 243 个原先无定制路径中 238 个与目标标签内容完全一致（含发布删除），其余 5 个仅为测试签名和订阅 token 批量适配；自动合并路径与 16 个冲突均保留 Fork 差异。
- 所有变更文本严格 UTF-8 解码正常，无 BOM、替换字符或合并标记；变更 Go 文件 gofmt 检查无输出；`git diff --check` 与 81 个未跟踪文件的 `git diff --no-index --check` 均无空白错误。
- 新文档引用路径有效；HEAD 仍为初始提交，暂存区为空。没有 git add/commit/push、业务服务启动或生产迁移。所有验收任务已完成，无待用户决策项。

## 交付限制

- 新 SQL 仅交付；生产部署前需备份并验证；隔离容器结果不代表生产大表性能。新汇总触发器/状态列的回退须验证兼容性，不能自动删列或索引。
- 保持现有 pending 退款及订阅全期语义的降级限制。生产实例应使用相同业务版本。
- OpenSpec CLI 未安装，未安装新工具；沿用现有 spec-driven 结构并检查文档、引用和编码。

## 提交前复审（2026-11-17）

用户追加要求：审查本次修改，无问题后推送代码。两路独立只读审查分别覆盖计费/订阅/迁移，以及 Kiro/封禁/前端/Ent/Wire/工具链；均未发现确定、可行动的问题。主控额外核对上游覆盖、退款金额格式化、用户端成本数据清理和提交范围。

本轮重新执行并确认退出状态：

| 命令 | 结果 |
| --- | --- |
| `git fetch origin main`、`git rev-list --left-right --count HEAD...origin/main` | 拉取成功，提交前本地与远端基线一致（0/0） |
| `go test -p 2 -tags=unit ./... -count=1` | 退出 0，64 个测试包通过、52 个包无测试 |
| `go build -p 2 ./... && go vet -p 2 -tags=unit ./... && go build -p 2 -tags embed ./...` | 退出 0 |
| `go test -p 2 -tags=integration ./internal/repository -run 'TestUsageBillingRepositoryApplyBatch\|TestUsageBillingBatchForkTermQuotaAndTokens\|TestGroupUsageRollup' -count=1 -v` | 退出 0，23 项顶层测试通过，使用隔离测试容器 |
| `go test -race -p 2 -tags=unit ./internal/service ./internal/handler -run 'TestUsageBillingLane_\|TestSchedulerFallback\|TestGatewayMessagesStickyWritebackAfterForward' -count=1` | 退出 0，两个包通过，无竞争报告 |
| `pnpm exec vitest run` | 退出 0，413 个文件 / 3,388 项测试通过，已包含新增图表回归 |
| `pnpm lint:check`、`pnpm typecheck`、`pnpm build` | 均退出 0；构建仍有非阻塞的大 chunk 提示 |

全部 351 个变更路径均属于本次升级；无本地配置、依赖目录或构建产物混入。318 份历史 SQL 再次逐字节核对未变；349 个变更文本 UTF-8/BOM/替换字符/冲突标记检查通过，295 个 Go 文件 gofmt 检查通过，新增行未发现常见凭据格式。复审与测试期间对全部变更内容进行 SHA-256 比对，确认没有并发改写；本轮只补充本验证记录，未改业务实现。

推送范围为当前 `main` 分支，不推送 release tag、不启动发布或生产迁移。线上供应商、真实支付/邮件和生产大表性能仍不属于本轮验证结论。
