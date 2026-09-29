# v0.2.10 同步验证记录

## 输入与范围
- Fork 起点 b92ce2353，工作区初始干净。
- 官方 v0.2.9 `4c00df2e0183e2c70b7fa8ba45914205e36aad0c` → v0.2.10 `2f3fed2fdb0787141294cec81487a5df30426f7f`；远程 tag object `e8cbebf37505922d30ffccd0a5bc19c4beec60ac`。
- 官方增量 33 commits / 118 files / +3,759 / -332；全量增量均已合入，没有部署配置或依赖变更。
- 42 个 Fork 重叠路径按三方合并，手工处理 Wire、billing_service、Responses 转换、Dashboard 测试四处冲突。保留 Kiro/IPBan 注入、历史模型价格、Kiro credits 和仪表盘默认今天。
- 其余 76 个路径中，74 个与官方目标逐字节一致；两个 settings.ts 仅收窄风控白名单提示，明确独立提示词审计、Access Ban、登录防护和上游限制继续生效。
- 560 个受版本管理、允许检查的非重叠定制文件及 301 个历史 SQL 保持原样。秘密、部署配置、IDE 状态、忽略文件及历史已删除截图不读取。
- 新 Claude 原生重置额度查询与 Fork 用户订阅永久重置卡相互独立；XorPay、Kiro credits/cache_read、Access Ban、登录失败封禁、提示词审计、重置卡与退款幂等、active_available、Ops、品牌别名、GLM 分类、支付 UX、VersionBadge 禁用在线更新均保留。

## 行为验证
- 先仅导入官方 Sonnet 5.5 测试，运行 `go test ./internal/pkg/apicompat -run '^TestSonnet55SignedThinkingResponsesRoundTrip$' -count=1`，旧实现缺失签名思考块（预期 3 个 output、实际 2 个），退出 1。
- 合入实现后 `go test ./internal/pkg/apicompat -run '^TestSonnet55' -count=1` 通过。
- 后端 `go test ./... -run '^$'` 通过，确认默认标签所有包可编译。
- 后端 `go test ./... -count=1` 通过：53 个有测试的包，service 用时 151.708s。
- 后端 `go build ./...`、`go vet ./...`、`golangci-lint run ./...` 均退出 0，lint 输出 `0 issues`。
- 前端 `pnpm run test:run` 通过：349 个文件、2,624 项测试。
- 双语提示最终编辑后，`pnpm exec vitest run src/i18n/__tests__/localeKeyCompleteness.spec.ts src/views/admin/__tests__/SettingsView.spec.ts` 通过：44 项测试。

## 独立审查
只读审查覆盖 118 个上游路径及 42 个重叠路径，确认历史 Fork 增删保留。提出的白名单提示范围和文档末尾基线问题均已修正，复核无剩余发现。

## 最终验证
- 后端 `go test -tags=unit ./... -count=1` 退出 0：61 个有测试的包，service 用时 192.620s。
- 双语提示最终编辑后，前端 `pnpm run typecheck`、`pnpm run lint:check`、`pnpm run build` 均退出 0；构建内置 i18n 检查 3 项通过。
- 根目录 `sh backend/scripts/resolve-version.sh` 输出 `0.2.10`。
- 后端 `go build -tags embed -o <临时输出文件> ./cmd/server` 退出 0，包含最终前端资源；未启动服务。
- `git diff --check` 通过，暂存区无变更。最终 126 个修改/新增文本通过 UTF-8 解码、无 BOM、无异常替换字符、无冲突标记检查。
- 最终范围仅为 118 个上游路径、3 份 Fork 说明和 5 份 OpenSpec 文档；74 个官方路径逐字节一致，两个双语提示仅一行差异，560 个非重叠定制文件及 301 个历史 SQL 不变。

## 限制
未执行真实 PostgreSQL/Redis 集成、供应商 OAuth/模型调用、支付网关联调、数据库迁移或部署；测试使用本地测试设施，不能替代上线环境验证。OpenSpec CLI/CodeGraph 当前不可用，沿用项目已有 spec-driven 文档结构及 Git/rg，不宣称 CLI validate 通过。前端构建有既有的大 chunk 和 Browserslist 数据过期提示，不为此升级依赖。没有暂存、提交、推送或创建 PR。
