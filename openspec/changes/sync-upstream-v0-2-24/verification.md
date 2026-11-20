# v0.2.24 同步验证（2026-11-20）

## 来源与范围

- GitHub release API 确认最新版本为 v0.2.24，发布于 2026-11-20 17:19:50（Asia/Shanghai）。tag object `7515db6dc56c2a6120dfb7abed70ef089ebc1457`；commit `5c2982a81b46e2142d0143d2f83c1ce9c4da1d24`。
- 初始 Fork `67bcc888951506789ba74d1385b951afb0176768`，main 与 origin/main 一致，工作区干净。上游增量 110 commits / 190 paths / +11,853 / -610，无依赖或迁移变更。
- 136 个无 Fork 重叠路径初次合并与目标标签逐字节一致。随后其中一份新增路由测试补上 Fork 的 IPBanService 参数；其余 135 个仍保持相同。54 个原有 Fork 重叠路径逐文件合并，9 个有文本冲突。
- 迁移目录 349 份历史文件（含 321 份 SQL）全部与初始 Fork 字节一致。支付、订阅、Kiro 专属模块与部署配置不变；共享网关保留封禁、会话/credits、全期额度及生图隐私标记。

## 已运行验证

- RED：仅引入上游 `requestmodel_test.go`，在旧实现执行 `go test ./internal/pkg/requestmodel -run 'TestFromBodyCandidatesFollowsWhatJSONHandlersParse|TestFromBodyCandidatesMultipartWithJSONPreamble' -count=1`。BOM、伪 multipart 相关四个子场景全部按预期失败（未提取到模型）。首次误在仓库根运行提示缺少 Go module，改为 backend 后得到上述行为失败；未修改模块配置。
- `pnpm exec vitest run`：433 文件、3,551 项通过。
- `pnpm typecheck`：通过。
- 本地 Chrome：`node e2e/image-studio-storage.cjs`、`node e2e/image-studio-ui.cjs`、`node e2e/image-studio-tabs.cjs` 全部通过；使用本机预装 Playwright，通过 NODE_PATH 解析，无安装新依赖。覆盖 IndexedDB 原子更新、账号隔离、批次/遮罩/预览、多标签同步、刷新不重放与无内容上传。
- `go test -p 2 -tags=integration ./internal/repository -run '^(TestOpsRepositoryGetOpenAITokenStats_WeightsRateByDuration|TestSchedulerCacheSnapshotUsesSlimMetadataButKeepsFullAccount|TestUsageBillingBatchForkTermQuotaAndTokens)$' -count=1 -v`：临时 PostgreSQL/Redis 容器内通过，包含免费与收费订阅 USD/token 全期累计。
- 首轮全量 Go 测试发现新增路由测试漏传 Fork 的 IPBanService 参数；补充 nil 后 `go test -p 2 -tags=unit ./internal/server/routes -count=1` 通过。首轮 vet 同样读取了修复前测试，后续完整复跑已通过。
- 首轮 repository 的 `TestServerTimingConnectorRecordsDriverCallsWithoutRowLifetime` 在并发编译时 app=34.6ms/db=45.1ms 触发耗时阈值；实现及测试均无本次差异，随后低并发全 repository 复跑通过（13.736s），未放宽断言。

## 边界

不执行生产迁移或部署，不调用真实供应商、支付或邮件。Grok 公共编辑接口支持官方 5 张图；工作台既有 3 张参考图/42 MiB 总请求预算保留。重复字段严格拒绝可能影响非标准客户端。

OpenSpec CLI 未安装，按项目既有结构保留规划、规范、任务与验证记录；最终检查引用、UTF-8、BOM 和差异。

## 审查与复核

- 独立审查：无 P0–P2 发现，代码层面可合并，最终推送以本任务完整验证通过为门槛。覆盖九处冲突、请求字段校验、公共/内部生图路由、Kiro 会话/credits、凭据 CAS、隐私与主要协议转换。审查员重复启动的 service 定向测试因主机负载主动终止（143）；未用此结果宣称通过，以主任务全量测试为准。
- `go test ./internal/pkg/requestmodel -count=1`：全包通过，BOM/multipart 回归 GREEN。
- `go test -p 1 ./internal/repository -run '^TestServerTimingConnectorRecordsDriverCallsWithoutRowLifetime$' -count=10`：未修改实现或断言，十次通过。失败时 2ms 驱动 sleep 在高并发编译下延迟累计超过30ms应用等待；属于已有依赖墙钟的测试波动，最终低并发完整测试也已通过。
- 改动/新增 197 份文本文件均为 UTF-8，无 BOM、替换字符及遗留冲突标记；`git diff --check` 通过。手动冲突 Go 文件及最终 Kiro 测试 gofmt 通过。全增量 gofmt 检查发现上游 `openai_oauth_model_support_test.go` 四行注释间距差异，保持官方原文，不新增纯样式差异。

- 修复后 `go build -p 2 ./... && go vet -p 2 -tags=unit ./...`：退出 0。前端 `pnpm lint:check`：退出 0。

- `pnpm build`：退出 0，存在非阻塞 Browserslist 数据过期和 >500kB chunk 提示。`go build -p 2 -tags=embed ./...`：退出 0。未更新依赖锁文件来消除提示。

## 首轮 service 失败的定位与收尾

- `TestAccountUsageService_AddWindowStatsKeepsKiroCredits` 使用旧的 int64 缓存键，生产实现已由上游升级为账号 ID 与窗口时长组合键。仅把测试读取键改为 `windowStatsCacheKey{accountID: 123, duration: 5 * time.Hour}`，保留 credits、缓存类型与存在性断言；生产 Kiro credits 转换未改动。
- `TestFilterThinkingBlocksSignedHistorySkipsDecode` 首轮全包测试记录 40 次分配，阈值为 16；过滤函数、调用 helper 和该测试没有本次差异。两个测试独立运行 `go test -p 1 -tags=unit ./internal/service -run '^(TestFilterThinkingBlocksSignedHistorySkipsDecode|TestAccountUsageService_AddWindowStatsKeepsKiroCredits)$' -count=20` 全部通过。
- `go test -p 1 -tags=unit ./internal/service -run '^$' -bench '^BenchmarkFilterThinkingBlocksSignedHistory/Precheck$' -benchmem -benchtime=10x -count=3` 三次均为 0 B/op、0 allocs/op。现有分配检查使用进程级计数，首轮现象符合其他后台任务干扰；未更改阈值或过滤实现，最终仍以完整测试复跑为准。
- 独立审查员追加复核最终一行缓存键修改：与生产 FiveHour 窗口读写键吻合，无新发现；未重复启动测试。
- 曾提前启动的完整复跑在发现首轮 service 失败后主动取消，不作为通过证据；最后完整验证使用 `go test -p 1 -tags=unit ./... -count=1`。

## 最终验收

- 最终 `go test -p 1 -tags=unit ./... -count=1` 退出 0：64 个含测试包全部通过，52 个包无测试；其中 service 277.856s、repository 11.987s、routes 11.795s。首轮两项已有性能计时/分配检查均保持原断言并通过。
- `git fetch origin main` 后远端仍为初始 Fork 提交；未改写历史。197 份交付文件限定于 190 个上游路径、Fork 兼容测试、版本与同步文档；无秘密配置、依赖或构建生成物进入提交范围。
- 交付边界：未部署、未执行生产迁移、未调用真实供应商。前端构建的 Browserslist 与大 chunk 提示、上游原有四行 gofmt 注释间距差异不阻塞构建；OpenSpec CLI 不可用，文档采用人工和脚本核验。
