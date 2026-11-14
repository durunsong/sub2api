# 验证记录（2026-11-14）

## 业务与根因

用户明确确认：90 天卡重置一小时后，应剩余 89 天 23 小时。消费卡事务原本已正确保存 90 天期限，错误来自后续模型、服务、SQL 筛选和前端把周期额度重新解释为固定 30 天。修复覆盖这些读写路径；兼容字段仍名为 monthly_*。

## 失败复现与回归

- 原前端倒计时断言得到 29 天而非 89 天；后端模型得到 12 月 14 日而非次年 2 月 12 日。改为订阅到期时间后通过。
- 临时恢复旧 SQL 自动月刷新谓词，真实 PostgreSQL 集成测试在第 61 天错误返回已耗尽订阅，断言失败；还原新谓词后通过。该临时修改已恢复。
- 延期保留真实锚点测试先捕获旧代码把起点推到未来的失败。移除旧重定锚逻辑后，通过延期、过期状态恢复、行锁/回滚及两种旧 HTTP 参数的兼容测试。
- 90 天卡消费验证期限、日/周/周期 USD 与 token 清零、扣卡一次、幂等重放保留后续用量；管理员单条/批量重置不延期且历史记录唯一；延期不清零用量或伪造锚点。
- 固定时钟覆盖一小时后、第 30/60 天、到期前一秒及精确到期，兼容普通/旧午夜/已滚动锚点；7/30/90/180 天期限、周窗口七天和日午夜保持原规则。
- 标准协议与 Google 协议认证中间件都拒绝第 61 天已耗尽的 90 天订阅，保持各自原有 429 错误格式。
- 实际用户/管理员订阅视图断言 89 天 23 小时；密钥用量页同时验证摘要和明细的周期额度标签。

## 实际命令

以下路径均相对仓库根目录。

| 目录 | 命令 | 结果 |
| --- | --- | --- |
| backend | `go build ./...` | 退出 0 |
| backend | `go vet ./...` | 退出 0 |
| backend | `go test -tags=unit ./...` | 全量退出 0；service 包 242.463s，通过后再次完成最终定向回归 |
| backend | `go test -tags=unit ./internal/service ./internal/handler/admin ./internal/server/middleware -run 'Test(SubscriptionTermQuota|AdjustSubscription|SubscriptionExtendAcceptsLegacy|APIKeyAuthRejectsExhausted)' -count=1` | 三个包通过 |
| backend | `go test -tags=unit ./internal/service ./internal/repository ./internal/handler ./internal/handler/admin ./internal/handler/quotaview ./internal/server/middleware ./internal/server/routes ./internal/pkg/adminaccess -run 'Subscription|ResetCard|ResetHistory|Monthly|UserPlatformQuota|ConsumeReset' -count=1` | 最终 8 个包全部通过 |
| backend | `go test -tags=integration ./internal/repository -run 'Test(UserSubscriptionRepoSuite|ConsumeResetCard|SubscriptionReset|AdminReset)' -count=1` | PostgreSQL/Redis 测试容器回归通过 |
| backend | `go test -tags=integration ./internal/repository -run 'Test(ResetHistory|ResetDailyQuota_Concurrent)' -count=1` | 历史记录与兼容旧重置并发回归通过 |
| frontend | `pnpm exec vitest run` | 最终 410 个文件、3337 项测试全部通过 |
| frontend | `pnpm typecheck` | 最终退出 0 |
| frontend | `pnpm lint:check` | 最终退出 0 |
| frontend | `pnpm build` | 退出 0，包含 i18n 完整性检查及类型编译 |
| 仓库根目录 | `git diff --check`、变更 Go 文件 gofmt 检查、UTF-8/BOM/异常替换字符扫描 | 通过 |

## 审查与边界

独立只读审查覆盖支付/兑换/管理员分配、卡消费行锁和幂等、显式重置、自动维护、可用额度 SQL、认证中间件、计费缓存、进度与历史处理。首轮发现的延期伪造锚点、密钥用量页旧月标签均已修复，复审无阻塞问题。

不新增依赖或迁移，不执行生产部署或修改生产数据；不恢复旧版本已经清零的历史用量，不补扣用户。回退版本会恢复旧 30 天规则，部署需保持实例版本一致。构建存在既有 Browserslist 数据过期和大体积 chunk 提醒，不影响退出状态。OpenSpec CLI 未安装，未安装新依赖；规格、场景、任务和实现进行人工交叉核对。
