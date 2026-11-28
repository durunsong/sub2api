# v0.2.26 同步验证（2026-11-29）

## 来源和范围

GitHub releases/latest 的 HTTP 302 指向 v0.2.26，发布 API 返回 403，因此不推测发布时间。实际 fetch 官方 tag，目标 commit `9494105c7e350937a61bd78b3f913a5a10867f58`；Fork 初始 main/origin/main 均为 `ac387298c` 且工作区干净。

官方增量 152 commits / 201 paths / +10030 -1110。两份部署示例不变；199 个源码路径中 1 个 VERSION 已与官方相同、105 个无 Fork 差异、34 个新增、48 个自动三方合并、11 个需手工解决冲突。Fork VERSION 按标签升至 0.2.26。无新增依赖或 SQL，不执行生产迁移/部署或真实模型、支付、邮件请求。

## 回归与兼容

- RED：先引入官方 ProxySelector.accountCount.spec.ts，旧实现 3 项失败；整合后与 Kiro/上游错误统计组件共 11 项通过。
- 审核队列新测试初次编译失败，因为 Fork enqueueRecord 有额外 context 参数；三个官方测试调用补齐 context，保持原有测试断言。
- 三类 Fork 兼容问题先由新测试复现：图片安全 service/handler 保存提示词；Antigravity QUOTA_EXHAUSTED 日志回显私有内容；粘性 CAS 和降级写回均把 Kiro 2h TTL 改回 1h。
- 修复：图片安全 handler 在异步闭包前清除内容并传递隐私标志，service 再清空 excerpt/error，保留安全事件元数据；新配额日志调用 ephemeralLogText；CAS 与 fallback 使用原分组 TTL。审核 lazy queue 保留隐私请求的同步安全处理，Kiro credits 在新字节事件循环内继续采集。
- TTL 初次 GREEN 暴露测试夹具遗漏 Hydrated/Status，导致分组上下文无效。补齐真实分组必要字段，并以 Go overlay 临时恢复固定 TTL 实现重新验证 RED；正式源码不作回退。独立审查确认测试覆盖 CAS 和 fallback 两条路径。
- 前端 `pnpm --dir frontend exec vitest run`：447 文件、3655 项通过，包含图片/视频、支付、订阅、品牌等既有回归。

## 上游行为变化

按官方 v0.2.26 源码同步：Chat 转 Responses 因 length 截断以 `response.incomplete` 收尾；Antigravity 配额耗尽立即冷却并切换账号；原生 Anthropic 流中断仍保留 message_start 已上报 output_tokens；新增 `k3.1-pro` 与 `k3.1-pro-highspeed` 参考计价，远端价格和渠道自定义价格仍优先。渠道实际响应模型展示开关默认关闭。上线时应核对依赖这些行为的客户端和已有定价覆盖。

## 已完成验证

- `pnpm --dir frontend typecheck && pnpm --dir frontend lint:check`：退出 0。
- `pnpm --dir frontend run build`：退出 0，i18n 3 项通过，1199 模块完成生产构建。存在 Browserslist 数据较旧和大于 500 kB chunk 提示，未安装依赖或扩大范围调整拆包。
- 后端目录运行 `go build -p 2 ./... && go vet -p 2 -tags=unit ./... && go mod verify`：退出 0，all modules verified。
- 后端目录运行 `go test -p 1 -tags=unit ./... -count=1`：退出 0，全部包通过；service 267.514 秒，包含既有 Kiro/支付/订阅/媒体隐私和本次新增上游、Fork 回归。
- `go build -p 2 -tags=embed ./...`：退出 0，验证本次前端产物与后端嵌入发布组合。
- `go test -p 1 -tags=unit ./internal/service ./internal/handler -run 'TestEphemeral|TestStickyWritebackPreservesKiroTTL|TestRecordImageSafetyEvent|TestOpenAIGatewayHandlerImages|TestContentModerationQueue|TestGatewayHandler.*Sticky' -count=1`：退出 0，service/handler 均通过。内存竞争期间取消过一次重复定向编译，最终已完整重跑。
- 修正 TTL 分组夹具后，以 Go overlay 只把本次 TTL 修复临时替换为固定 TTL：CAS/fallback 两个子测试均按预期因 1h != 2h 失败（退出 1），证明正式通过不是无效断言。
- 后端目录运行 `go test -p 1 -tags=integration ./internal/repository -run 'TestGatewayCacheSuite|TestUsageLogRepoSuite|TestUsageBalanceFloor|TestResetHistory_ConsumptionReplayAndBulkReset|TestConsumeResetCard_NinetyDayQuotaEndsWithTerm' -count=1 -v`：退出 0，真实临时 PostgreSQL/Redis 容器，79 条 PASS、无跳过，覆盖余额地板并发/幂等/回滚、90天全期卡与重置历史、Redis CAS 和用量/上游错误统计。
- 独立只读审查：确认三类兼容问题修复，未发现未解决 P0-P2，无阻塞补测建议。审查覆盖 Kiro credits/TTL、NoContentPersistence、审核队列、用户响应模型和 Wire/账号统计。
- 133 个无 Fork 重叠的上游路径与目标逐字节一致；其余 6 个上游新增/无重叠路径仅做隐私适配和测试 context 参数兼容。所有 59 个 Fork 重叠路径按三方合并保留定制。
- 351 份历史迁移文件（含 323 份 SQL）与初始 Fork 字节一致；两份部署示例不改；图片/视频独立模块、支付核心、余额结算和订阅核心未被同步覆盖。
- 207 个交付文本文件严格 UTF-8 解码通过，无 BOM/U+FFFD/冲突标记；改动 Go 文件 `gofmt -l` 无输出，16 个文档相对引用存在，`git diff --check` 通过。
- 推送前 fetch origin/main，仍为初始 ac387298c，无远端新增提交。

## 交付

适用自动化检查与独立审查均已完成，无未解决阻塞项。按用户要求提交并快进 main 后正常推送，实际提交及远端核对结果由最终交付回复记录。没有进行生产部署、生产迁移或真实模型/支付/邮件联调；构建警告见上文。
