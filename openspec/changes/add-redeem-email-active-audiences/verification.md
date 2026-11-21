# 验证记录（2026-11-21）

## HTTP 请求校验遗漏修复（同日用户反馈）

初次交付漏改 `RedeemEmailRecipientSearchRequest` 和 `CreateRedeemEmailBatchRequest` 的 binding 标签：旧 `oneof` 不接受新范围，旧 `max=4` 拒绝五/六项组合。此前 service/repository 和前端测试未覆盖新范围穿过真实 HTTP JSON binding，因此初次“完整链路”验证结论不充分。

- 新增 `TestRedeemEmailHandlerAcceptsAdvertisedAudiences`，通过 Gin HTTP 请求检查服务公布的每个选项、两种活跃范围组合、五项及六项组合，分别覆盖搜索和建批。
- RED：两个接口均复现 `Segments[0] ... oneof`，五/六项组合复现 `Segments ... max`。
- 修复：两处白名单增加 `active_7d` / `active_30d`，上限改为 6，保留必填、空值和非法枚举校验；零余额互斥继续由服务层校验。
- GREEN：上述测试通过；`TestRedeemEmailHandlerRejectsInvalidAudienceRequests` 验证空范围、非法值与超限请求在服务调用前被拒绝。
- `go test ./internal/handler/admin -count=1`：整个管理 handler 包通过。
- `go test ./internal/service ./internal/repository ./migrations -run 'RedeemEmail|RedeemInactiveCoupon|NormalizeRedeemEmail|NewRedeemEmail|HashRedeemEmail|FormatRedeemEmail' -count=1`：三个包通过。
- `go build ./...`、`go vet ./...`：通过。
- 本补丁未改前端、SQL 或数据库结构；未重跑前端全量及数据库集成测试，未部署或执行实际群发，无新增迁移。

## RED → GREEN

- 新分段校验测试先因 `REDEEM_EMAIL_FILTER_INVALID` 失败，修改后通过。
- 前端 7/30 天选项测试先因选项缺失失败，修改后通过搜索、确认和建批参数断言。
- 真实 PostgreSQL 测试先因活跃用户查不到、新迁移缺失失败，修改后通过。
- 集成回归发现调整 SQL 占位符时误改明细写入参数，已修正并重新通过完整邮件批次生命周期测试。

## 最终检查

- 后端：`go test ./internal/service ./internal/repository ./internal/handler/admin ./migrations -run 'RedeemEmail|RedeemInactiveCoupon|NormalizeRedeemEmail|NewRedeemEmail|HashRedeemEmail|FormatRedeemEmail' -count=1`，四个包通过。
- 后端集成：`go test -tags=integration ./internal/repository -run 'TestRedeemEmail.*Integration' -count=1`，通过。使用 Docker 临时 PostgreSQL/Redis，执行全部启动迁移；新增测试还使用事务临时表验证精确 7/30 天边界、边界外微秒、并集去重、旧默认、账户排除、资格复核及 CHECK 约束兼容。
- 后端：`go build ./...`、`go vet ./...`，均通过。
- 前端：`pnpm vitest run`，433 个测试文件、3553 个测试通过。
- 前端：`pnpm typecheck`，通过。
- 前端：`pnpm exec eslint src/components/admin/redeem/RedeemEmailCampaignDialog.vue src/components/admin/redeem/__tests__/RedeemEmailCampaignDialog.spec.ts src/types/index.ts src/i18n/locales/zh/fork.ts src/i18n/locales/en/fork.ts`，通过。
- 独立审查无阻塞问题；旧筛选错误提示的文案遗漏已修正。
- `git diff --check` 与 UTF-8、无 BOM、无替换字符检查通过。

## 边界

未执行生产迁移、部署或真实邮件群发，未运行无关后端全量测试。迁移 253 随代码交付，部署新版时须应用该迁移。当前机器无 OpenSpec CLI，规格使用已有目录格式保存。前端测试提示 Browserslist 数据过期，不影响退出状态；未升级依赖。
