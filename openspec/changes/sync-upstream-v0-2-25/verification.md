# v0.2.25 同步验证（2026-11-24）

## 来源与范围

GitHub release API 核实 latest=v0.2.25，发布于 2026-11-24 11:21:08（Asia/Shanghai）；tag object `8fab111f92447a3edbcdf67a517e45b6a3cc4d31`，commit `348997267d756eab5729bb6b78a275a15d18796a`。初始 Fork `7431026d1` 与 origin/main 一致且干净。严格整合 v0.2.24..v0.2.25，15 commits / 91 paths / +3486 -319，36 路径含 Fork 差异，三方合并无文本冲突。

## 回归与兼容适配

- RED：仅引入上游 HeaderOverrideEditor.spec.ts，旧实现运行 5 项全部按预期失败（没有变量按钮、预览及错误提示）。
- 首轮前端全量 434 文件中 433 通过、1 失败（3571 通过/1 失败）：AccountsView.sparkShadow.spec.ts 依赖旧列布局 self-start。新布局将名称和同步按钮放同一 flex 行，测试改为断言父级 items-center；链接协议、noopener、样式与提示内容断言均保留。
- 首轮 Go 失败：Fork buildKiroRelayModelsRequest 漏适配上游新增 context 参数。补为 withoutHeaderOverrideRequestScope(req.Context())，与其它共享模型抓取一致；不向后台模型列表上报用户身份。原 Kiro 请求 URL/认证测试保留。
- 三项新增 Kiro UI 测试先失败，再为创建/编辑/列表同步入口加平台守卫。创建时同时省略 Kiro 不支持的 probe/groupSync 开关（后端既有排除保持不变），测试检查最终提交及不调用探测；首轮新增提交场景缺少必填 base URL，补齐有效测试输入后继续验证。
- 临时 PostgreSQL/Redis 集成测试：`go test -p 1 -tags=integration ./internal/repository -run '^(TestAccountUpdatePreservesConcurrentProbeSnapshot|TestAdminAccountEditPreservesRateSynchronizedAfterLoad|TestProbeSnapshotSyncsRateOnlyForSuccessfulEnabledAccount|TestAccountUpdatePreservesConcurrentProbeEnableFlag|TestAccountUpdateClearsProbeSnapshotWhenIdentityChanges|TestBulkUpdateAndCredentialUpdateDeleteProbeKey|TestProbeSnapshotCASIncludesLoadedEnabledState|TestProbeSnapshotCASProtectsManualRateAfterSyncDisabled|TestProxyIdentityUpdateInvalidatesProbeAndRejectsInFlightSnapshot|TestSweepExpiredProxyWithoutFallbackInvalidatesOnlyExistingProbeSnapshot|TestSweepExpiredProxyFallbackRerouteDeletesProbeSnapshot|TestProbeGroupWriteHonorsDisableDuringRequest)$' -count=1 -v` 通过（12 个顶层测试，非跳过）。

## 交付边界

未部署、未执行生产迁移、未发送真实模型/支付/邮件请求。OpenSpec CLI 本机不可用，按项目既有结构记录并手动/脚本校验。最终验证结果与审查见下文；提交与推送结果由交付回复记录。

## 独立审查

只读独立 reviewer 未发现 P0–P2 或需单列的 minor 问题。检查身份变量可信来源、模板秘密脱敏、WS 连接隔离、生图内部认证及 NoContentPersistence、新 preview 凭据/URL/重定向/响应体与受限管理员 Key 权限、组名备注并发写入、Kiro 兼容适配及 OpenAI 错误脱敏。未重复全量测试，未调用真实外部/生产服务；这些边界不构成生产验证。

## 已完成最终验证

- `pnpm --dir frontend exec vitest run`：434 文件、3575 项通过。包含身份变量编辑器 GREEN、三项 Kiro 兼容回归、邮件券受众/批量兑换、订阅与生图相关测试。
- `go build -p 2 ./... && go vet -p 2 -tags=unit ./... && go mod verify`：退出 0，all modules verified。
- `pnpm --dir frontend run build`：退出 0；有非阻塞的大于 500 kB chunk 提示。
- 55 个无 Fork 重叠的上游路径与 v0.2.25 逐字节一致。其余 36 个路径保留原三方合并结果，仅在版本、Kiro 请求头调用、三个同步 UI 入口/创建 payload 及对应测试中做必要适配。额外修改仅为旧布局测试和 8 份版本/同步文档。
- 350 份历史 migration 文件（含 322 份 SQL）全部与初始 Fork 字节一致；专属 Kiro、订阅、支付、邮件、生图核心模块及 VersionBadge 未被替换。
- 全部 100 个交付文本文件通过 UTF-8 解码、无 BOM/U+FFFD/冲突标记检查；变更文档的本地 Markdown 引用均存在；`git diff --check` 通过。

- 最终 `pnpm --dir frontend typecheck && pnpm --dir frontend lint:check`：退出 0。
- `go build -p 2 -tags=embed ./...`：退出 0，确认前端产物可嵌入后端。
- 推送前重新 fetch origin/main，仍为初始 Fork 提交，没有需要整合的远端新增提交。

## 最终验收

- `go test -p 1 -tags=unit ./... -count=1`：退出 0，64 个含测试包全部通过，52 个包无测试；service 273.324s。
- 改动 Go 文件 `gofmt -l` 无输出；最终 `git diff --check`、100 个文本文件 UTF-8/BOM/U+FFFD 检查通过。
- 无未处理的 P0–P2 或 minor 审查发现。唯一构建提示为前端大 chunk，不影响构建；未开展生产/真实供应商验收。
- 所有代码已冻结并完成上述检查，交付使用单次普通提交、快进 main、正常 push origin/main；不强推、不改写历史。最终提交哈希与远端核对结果以本次交付回复为准。
