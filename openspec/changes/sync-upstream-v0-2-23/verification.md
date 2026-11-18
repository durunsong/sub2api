# v0.2.23 同步记录（2026-11-18）

- 官方 `v0.2.22..v0.2.23`：32 个提交、90 个路径、+5,251/-453；标签对象 `ce19a640eab5c4b5ef1827c7c1fcfce49569411a`。Fork 起点 `39d34433c`，工作区干净。
- 90 个路径全部整合，其中 27 个与历史 Fork 改动重叠；其余 63 个路径在初次合并后与发布标签逐字节相同。无依赖或部署配置变更，320 份既有 SQL 未改。
- `go test -tags=integration ./internal/repository -run '^TestChannelUltrafastPricing' -count=1 -v`：PostgreSQL/Redis 测试容器内两项通过，覆盖旧价格卡及新倍率的增改替换。
- `pnpm exec vitest run`：433 个文件、3,509 项通过。首次运行暴露中英文说明丢失 `service_tier=ultrafast`；已补回文案，未放宽测试。`pnpm typecheck`、`pnpm lint:check`、`pnpm build` 通过。
- 本机 Chrome 对本地 fixtures 执行 `node e2e/image-studio-storage.cjs`、`image-studio-ui.cjs`、`image-studio-tabs.cjs` 均通过，覆盖本地存储、遮罩、独立任务、跨标签同步与不自动重放。
- 缓存明细回归：短期 3、长期 5、无聚合值时，修复前日志为 0，修复后为 8，含输入输出的总 token 为 27，Kiro credits 不变。`TestBuildRecordUsageLogPersistsPositiveKiroCredits`、`TestOpenAICacheWriteTTLRecordUsage` 通过。
- `go build -p 2 ./...`、`go vet -p 2 -tags=unit ./...`、`go build -p 2 -tags=embed ./...` 通过。未向生产供应商、支付或数据库发送请求。

## 提交前复审

- 缓存 TTL：复现跨帧仅报短期 30、长期 50 时，强制短期后实际缓存费用为 0.10 而非 0.16。按上游原始分桶合并、再投影到客户端指定档位后，两个方向均以 80 个 token 结算；后帧清零和未提供明细分别处理。交易、用量日志与账号统计均有断言。
- Kiro：Chat 流式和缓冲路径采集 credits，不向客户端回传，非 Kiro 平台不落库；生图隐私请求触发月配额 402 时，仍保留限流和切换，但不将上游正文写入 Warn 日志。
- OpenAI：持续分块传输较大的 SSE 行不再被误判为空闲，半行停流仍会在最后一个字节后触发超时。部分字节不算首个语义输出，不放宽首输出切换预算。
- 渠道：目录与强制档位尊重账号快照和官方套餐授权，模型默认倍率文案明确渠道价可覆盖。前后端拒绝不能精确写入 `NUMERIC(12,6)` 的倍率，NULL 继续代表默认值。未改历史价格或迁移。
- Go 全量 `go test -p 2 -tags=unit ./... -count=1`：64 个测试包通过；复审后的构建、vet 和 embed 构建通过。`go test -race` 定向覆盖空闲超时、跨帧缓存和 Kiro 用量，无竞态报告。数据库容器额外通过 `TestUsageBillingBatchForkTermQuotaAndTokens` 的免费和收费路径。
- 前端最终 `pnpm exec vitest run`：433 个文件、3,512 项通过；`pnpm typecheck`、`pnpm lint:check`、`pnpm build` 通过，构建有非阻塞的 chunk 大小提示。未安装 OpenSpec CLI；按仓库既有结构核对规划、引用及编码。
