# 验证记录（2026-11-14）

## 范围与实现

- 管理端与用户端订阅页均可分页查看重置卡消费和管理员单条/批量重置，展示时间、方式、日/周/月范围与卡期限。普通用户仅能访问自己的订阅；管理员可以查看已撤销订阅。
- 重置卡直接读取既有消费事实，包含消费后退款的卡；不回填没有可靠明细的历史管理员操作，不记录自动周期刷新。
- 管理员额度清零和记录写入使用同一个 PostgreSQL CTE，复用批量操作外层事务；保留原有金额、token、期限与幂等语义。
- 新增迁移 `251_subscription_quota_reset_history.sql`；没有修改历史迁移，没有执行生产迁移或部署。

## 自动化验证

| 目录 | 命令 | 结果 |
| --- | --- | --- |
| `backend` | `go test -tags=unit ./internal/service ./internal/repository ./internal/handler ./internal/handler/admin ./internal/server/routes ./internal/pkg/adminaccess -run 'Subscription\|ResetCard\|ResetHistory' -count=1` | 6 个包通过 |
| `backend` | `go test -tags=integration ./internal/repository -run 'TestResetHistory\|TestUserSubscriptionRepoSuite\|TestConsumeResetCard' -count=1` | PostgreSQL 测试容器集成回归通过 |
| `backend` | `go build ./...` | 通过 |
| `backend` | `go vet ./...` | 通过 |
| `frontend` | `pnpm vitest run src/components/subscription/__tests__/SubscriptionResetHistoryDialog.spec.ts src/views/user/__tests__/SubscriptionsView.spec.ts src/views/admin/__tests__/SubscriptionsView.filter.spec.ts` | 3 个文件、18 项测试通过 |
| `frontend` | `pnpm vitest run --maxWorkers=2 --minWorkers=1` | 最终全量回归 410 个文件、3317 项测试全部通过 |
| `frontend` | `pnpm typecheck` | 通过 |
| `frontend` | `pnpm lint:check` | 通过 |
| `frontend` | `pnpm build` | 通过，包含 i18n 检查和类型检查 |

新回归先观察预期失败再实现：管理员记录缺失、插入失败未回滚、用户权限/API 与页面入口缺失均得到验证。最终数据库回归覆盖旧卡、退款保留、稳定分页、重置范围、失败回滚、卡消费幂等和批量重复 ID；接口测试覆盖未认证、越权、无效 ID、已撤销订阅、分页上限及管理 Key 读取权限。前端测试覆盖入口、翻页、错误重试、关闭取消与旧请求响应隔离。

## 界面与审查

- 使用真实共享组件和项目样式、模拟历史数据进行桌面及 390 × 844 手机预览，确认中文记录、翻页、空态和关闭交互。该预览未连接真实账号或生产 API；临时预览文件和进程已移除。
- 独立只读代码审查未发现 Critical、Important 或 Minor 问题；重点检查事务原子性、权限隔离、历史保留、排序与请求竞态。
- `git diff --check` 通过；本次文本通过 UTF-8 解码、BOM 和异常替换字符扫描。原有三个兑换文件按 SHA-256 与任务开始时比较保持不变。

## 上线边界

- 上线需按项目部署流程备份并应用新增迁移，再运行依赖该表的新版应用。应用回退可保留新表和已有记录，旧版本运行期间的管理员重置不会产生新记录。
- 构建保留既有 Browserslist 数据过期及超过 500 kB 分包提示，没有新增依赖或进行无关分包调整。
- 提交只包含本任务文件和 Fork 说明；并行客服功能已独立提交，原有兑换改动不纳入本次提交。
- 初次前端全量运行遇到并行客服功能的两项测试失败；该功能独立修复后，先单独复验 16 项客服测试，再完成上述全量通过结果。
