# v0.2.21 同步验证记录

验证日期：2026-11-11。交付为工作区源码改动，不包含部署或数据库执行。

## 来源与范围

- 初始 Fork HEAD：`1e8e2321696c9d41d625931a7455f12f8eb70425`。
- 官方 base v0.2.20：`dab829a4e9b7d1cc307189aea8b0319eb29e29d0`。
- 官方目标 v0.2.21：`10ba8f9b4dc2e6023d4a9879ab77b21841babd10`。
- 标签间增量：82 commits、200 路径、+6,699/-541；没有合入目标标签之后的 main。
- 134 个纯上游路径、57 个干净三方合并路径、7 个冲突路径、1 个已有版本路径、1 个排除配置路径。应用 199 个发布路径（含版本）；部署配置 `deploy/config.example.yaml` 保持原样，`deploy/EDGE_SECURITY.md` 为同步的说明文档。
- 官方目标的 VERSION 仍是 0.2.20，本 Fork 的 `backend/cmd/server/VERSION` 明确设为 0.2.21。
- 未安装依赖或升级工具链；`go.mod` 和前端依赖清单/锁文件保持，`go.sum` 只同步上游校验和。

## 合并与补丁

七个冲突均保留双方功能：

1. `usage_log_repo_insert.go`：保留 Kiro credits，接入 30m/12h，INSERT 两条路径和参数数组均为 65 项。
2. `usage_log_repo_query.go`：SELECT 与扫描顺序同时保留 Kiro 和两档缓存。
3. `usage_log_session_id_unit_test.go`：校验 65 参数及 Kiro/请求 ID/session/native/TTL/created 尾部位置。
4. `openai_gateway_response_handling.go`：用量解析保留 TTL 与 KiroCredits。
5. `openai_gateway_service.go`：OpenAIUsage 同时保留所有计量字段。
6. `payment_refund_test.go`：同时保留 Fork 缓存失效失败回归与上游 EasyPay 回归。
7. `KeysView.vue`：合入加载状态/性能更新，保持用户端倍率隐藏。

适配新增的刷新令牌测试构造函数时保留 Fork `IPBanService` 参数，不修改生产封禁注入。

审查另发现目标版本缺口：WS v2、Messages/Chat typed usage 会遗漏缓存 TTL 桶，造成 12h 写入按通用短档价格计费。补丁将目标版已有严格解析移至 `internal/pkg/openai/cache_usage.go` 共用，并贯通 typed usage、WS 逐轮及累计映射。不改变上游价格规则：缺失、部分、非整数、负数或与聚合量不符的明细仍回退通用价。聚合量和两档计数作为同一快照替换，保留 Kiro credits，避免旧桶残留和跨轮重复累计。

## 红绿回归证据

- 合并前先运行上游 EasyPay 成功码测试，确认旧实现不能满足 code=0、非法配置及小数 code 的要求；合并后同命令退出 0。
- TTL 补丁前，实际 Messages 流式、缓冲转换、Chat 流式与 WS adapter 回归均观察到分档丢失。混合样例预期扣费 0.00195463125 USD，缺陷版本实际扣费 0.00098369375 USD。
- 修复后同一链路验证 token、实际用户扣费、总成本和账号统计成本。覆盖 JSON、SSE、透传 JSON、透传 SSE、Messages SSE、缓冲 SSE、Chat SSE、WS 共 8 条路径，每条包含 30m/12h/混合三种用量。
- 另覆盖 canonical zero、缺失/非法/不一致明细、复用解码对象、逐步快照替换、多轮累计/清空和 Kiro credits 共存。

## 实际执行结果

以下 Go 命令在 `backend/` 运行，pnpm 命令在 `frontend/` 运行。验证使用已有 Go 1.27.3、pnpm 10.33.2。

| 命令 | 结果 |
| --- | --- |
| `go test -p 2 ./internal/payment/provider -run 'TestEasyPayRefundSuccessCode\|TestNewEasyPayRejectsInvalidRefundSuccessCode' -count=1` | 先红后绿；合并后退出 0 |
| `go test -p 2 -tags=unit ./internal/pkg/apicompat ./internal/pkg/openai ./internal/service/openai_ws_v2 ./internal/service -run 'TestResponsesUsageCacheWriteTTL\|TestRelayUsageCacheWriteTTL\|TestOpenAICacheWriteTTL\|TestHandleAnthropicStreamingResponseMergesKiroCredits\|TestReadOpenAICompatBufferedTerminalMergesKiroCredits\|TestCopyOpenAIUsageFromResponsesUsage' -count=1` | 退出 0；所有匹配回归通过，openai 包在该筛选下无匹配用例 |
| `go test -p 2 -tags=unit ./... -count=1` | 最终重跑退出 0；64 个测试包通过，52 个包无测试；service 包 242.083 秒 |
| `go build -p 2 ./... && go vet -p 2 ./... && go build -p 2 -tags embed ./...` | TTL 最终生产代码修改后运行，退出 0 |
| `go test -race -p 2 -tags=unit ./internal/service ./internal/service/openai_ws_v2 -run 'TestAuthServiceRefreshTokenPairConsumesTokenOnce\|TestRelayUsageCacheWriteTTL\|TestOpenAICacheWriteTTLRecordUsage' -count=1` | 退出 0；两个包通过，无数据竞争报告 |
| `pnpm exec vitest run` | 退出 0；406 文件 / 3,254 测试通过，85.37 秒 |
| `pnpm lint:check && pnpm typecheck && pnpm build` | 退出 0；前端产物构建通过 |
| `git diff --check` 与变更 Go 文件 `gofmt -l` | 退出 0；无空白错误、待格式化 Go 文件 |

前端验证后没有再修改前端源码。构建的 Browserslist 版本提示及大于 500 kB chunk 提示为非阻塞警告，没有因此升级依赖或扩大重构。后端首轮全量测试曾发现新增测试少传 Fork 封禁服务参数，修正测试后服务包完整回归通过；最终全量以表中重跑结果为准。

## 定制与数据保护

- Kiro 平台、credits/cache_read、429 冷却、真实请求 ID、XorPay、Access Ban/登录自动封禁、提示词审计、永久重置卡/退款幂等、余额券群发与 Worker 清理、用户批量兑换均保留。
- 用户端 Claude 别名、GLM 分类、隐藏倍率与在线更新、客服二维码、品牌主题和 Ops 删除保留。独立审查覆盖 64 个上游重叠路径、7 个冲突与关键 Fork 锚点；TTL 补丁复审未发现 Critical/Important 问题。
- 初始 3 个已修改兑换文件的 SHA-256 与开始时完全相同：`frontend/src/views/user/redeemCodes.ts`、`frontend/src/views/user/__tests__/redeemCodes.spec.ts`、`frontend/src/views/user/__tests__/RedeemView.spec.ts`。
- 315 份历史 SQL 与初始 SHA-256 完全一致。仅新增 `249_add_usage_logs_api_key_cost_index_notx.sql` 和 `250_openai_cache_write_ttl.sql`；同号 249 与既有 MCP SQL 按完整文件名并存，共 317 份 SQL。另同步上游迁移测试文件，不代表第三份 SQL。
- 源码整合交付时 HEAD 保持初始值、暂存区为空，没有提交或推送；后续用户授权的提交准备见下节。始终未启动业务服务、执行迁移或部署。
- 最终核验 214 个工作区变更路径（其中 3 个是原有未提交文件，本任务涉及 211 个路径）；无范围外改动。文本均可严格 UTF-8 解码，无 BOM、替换字符或冲突标记，新增内容无个人绝对路径；152 个变更 Go 文件无待格式化项，OpenSpec 结构和新增文档链接有效。

## 边界与上线前事项

- 未连接真实 PostgreSQL/Redis、供应商、支付或邮件服务；单元回归和 loopback WS 测试不能替代部署环境联调。
- 新 SQL 仅交付。上线负责人需备份数据库并验证大表并发索引、失败索引重试、新增列及可空 12h 价卡兼容性；本次未验证真实迁移性能或回滚执行。
- 不改历史退款/永久卡状态；既有 pending 退款在降级旧版本前须按原有规则结清。
- OpenSpec CLI 未安装，未安装新工具；沿用现有 spec-driven 结构，检查 proposal/design/tasks/spec 的标题、引用和编码。
- Grok Composer 专用聚合 helper 不在当前 GPT/OpenAI TTL 调用链内，未为本次补丁扩展其行为。

## 用户授权后的提交前审查

2026-11-11 用户明确要求“审查一下你修改的代码，然后直接 push”。本节记录该后续授权，不改变源码整合阶段的历史验证结果。

- **Standards**：独立规范审查为 0 个可行动发现，Critical/Important 均为 0；重点覆盖计费、SQL 接线、认证、支付、迁移和 Fork 保护。
- **Spec**：独立需求审查为 0 个可行动发现，Critical/Important 均为 0；199 个发布路径完整，额外补丁与测试符合已记录的完整性修复范围。
- 暂存检查发现新 `250_openai_cache_write_ttl.sql` 文件尾部多余空行，已移除；这是新迁移的格式修整，不改变 SQL 语义或历史迁移。
- 提交范围精确为 211 个文件。原有 3 个兑换文件未暂存且 SHA-256 不变；另通过 TypeScript AST 对比确认这些未提交差异仅为格式，测试行为与提交版本一致。
- 本轮重新执行 `go test -p 2 -tags=unit ./migrations ./internal/repository ./internal/pkg/apicompat ./internal/service/openai_ws_v2 ./internal/payment/provider -count=1`：5 个包通过，退出 0。
- 本轮重新执行服务层缓存 TTL、刷新令牌原子消费、Kiro credits 和退款回归：退出 0，命令如下。

```sh
go test -p 2 -tags=unit ./internal/service -run 'TestOpenAICacheWriteTTL|TestAuthServiceRefreshTokenPairConsumesTokenOnce|TestHandleAnthropicStreamingResponseMergesKiroCredits|TestReadOpenAICompatBufferedTerminalMergesKiroCredits|TestCopyOpenAIUsageFromResponsesUsage|Refund' -count=1
```

- 本轮前端定向回归 6 个文件、138 项测试通过，退出 0，命令如下。

```sh
pnpm exec vitest run src/views/user/__tests__/KeysView.spec.ts src/views/user/__tests__/RedeemView.spec.ts src/views/user/__tests__/redeemCodes.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts src/components/admin/channel/__tests__/types.spec.ts src/components/payment/__tests__/PaymentProviderDialog.spec.ts
```

- `git diff --cached --check`、暂存内容与工作区一致性、UTF-8/BOM/替换字符/冲突标记、315 份历史 SQL 及保护文件检查通过。
- `git fetch --no-tags origin main` 后确认原始 HEAD 与 origin/main 无分叉。推送目标为现有 `origin/main`，使用普通 push，不强推、不创建 release 标签，不执行迁移或部署。
