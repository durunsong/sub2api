# v0.2.11 同步验证记录

## 输入与范围
- Fork 起点 `352146f35`，工作区初始干净。
- 官方 v0.2.10 `2f3fed2fdb0787141294cec81487a5df30426f7f` → v0.2.11 `96f4c115c9749078f90cbf210a01d39baf3f53b6`；远程 tag object `881a722349105e21aaa38e11de90dd05046ebe4a`。
- 官方增量 23 commits / 90 files / +5,146 / -296。同步 89 个非部署路径；`deploy/config.example.yaml` 保持原样，新增参数默认值由源码提供；无依赖或数据库迁移变化。
- 24 个 Fork 重叠路径三方合并；Wire 冲突保留 Kiro/IPBan 注入并前移幂等协调器；版本按 release 设为 0.2.11（官方文件仍是 0.2.10）。
- 65 个非重叠上游路径中，63 个与官方逐字节一致；两个新增 handler 测试各补齐两个 Fork Kiro 构造参数。
- 588 个受版本管理、允许检查的非重叠定制文件和 301 个历史 SQL 保持原样；保护范围、秘密、本地配置和忽略文件不读取。
- Kiro credits/cache_read、请求 context/冻结计价时刻、历史模型兜底价格、XorPay、Access Ban/登录防护、提示词审计、订阅重置卡/退款幂等/active_available、Ops、品牌别名、GLM 分类、支付 UX、VersionBadge 均保留。

## 行为验证与适配
- 仅导入官方模型白名单新断言后，`pnpm exec vitest run src/composables/__tests__/useModelWhitelist.spec.ts` 在旧实现上退出 1：GPT-6.1 Sol 模型及预设映射两项缺失，其余 22 项通过。合入实现后同命令退出 0，24 项全部通过。
- 官方 GPT-6.1 Sol 协议桥接测试在旧版本即可通过，故仅作为兼容回归，不作为红灯证据。
- 默认编译预检发现 `gateway_inflight_reservation_test.go` 少两个 Fork 构造参数；unit 标签发现 `gateway_handler_claude_code_fallback_test.go` 同类问题。仅适配测试构造调用，未改变服务接口。修复后 `go test -tags=unit ./internal/handler -count=1` 通过。

## 已完成自动化检查
- 后端 `go test ./... -count=1` 退出 0：52 个有测试的包，service 包用时 151.189s。
- 后端 `go build ./...`、`go vet ./...`、`golangci-lint run ./...` 均退出 0，lint 输出 `0 issues.`。
- 前端 `pnpm run test:run` 退出 0：349 个文件、2,667 项测试。
- 前端 `pnpm run typecheck`、`pnpm run lint:check`、`pnpm run build` 均退出 0；build 内置 i18n 3 项测试通过。
- `sh backend/scripts/resolve-version.sh` 输出 `0.2.11`。
- 后端 `go build -tags embed -o <临时输出文件> ./cmd/server` 退出 0，包含最终前端资源；未启动服务。

## 最终验证
- 修正测试参数后，`go test -tags=unit ./... -count=1` 全量重跑退出 0：61 个有测试的包，service 包用时 193.520s。
- 两处新增测试仅增加 Fork 构造参数；源码、前端和依赖未再改变。最终上游路径一致性复核：63 个非重叠文件逐字节一致，两个测试文件为兼容适配，23 个非 VERSION 重叠文件完整保留历史定制增删行。
- 最终 97 个修改/新增文本 UTF-8 解码通过，无 BOM、异常替换字符或冲突标记。`git diff --check` 通过；OpenSpec 引用路径存在。
- 最终范围仅为 89 个上游路径、3 份 Fork 说明和 5 份 OpenSpec 文档；暂存区为空。301 个历史 SQL 及 588 个可检查非重叠定制文件不变。

## 独立审查
只读审查核对 23 个非 VERSION 重叠文件：基线到 Fork 的定制增删行在目标上游到合并结果中全部保留；另检查 Kiro 桥接、Access Ban、订阅/退款和 UI 关键文件。余额预留通过更新后的请求 context 交接至异步计费；订阅模式跳过。没有发现需要修复的合并回归。

## 限制与运行行为
未启动服务、执行数据库迁移、真实 PostgreSQL/Redis 环境联调、供应商 OAuth/重置调用、支付或部署。本地测试使用测试设施；不能替代上线环境验证。余额预留是估算准入，沿用上游首个请求放行、Redis 故障/未知价格默认 fail-open，不是资金冻结或绝对防透支保证；订阅永久卡与管理员 Claude 原生重置独立。
OpenSpec CLI/CodeGraph 不可用，沿用现有 spec-driven 文档结构及 Git/rg，不宣称 CLI validate 通过。前端构建有大 chunk 和 Browserslist 数据过期提示，未因此升级依赖。未暂存、提交、推送或创建 PR。

## 用户要求的二次审查（2026-09-30）

升级范围与 Fork 保留核对通过，但复审新增确认一项未修复问题：**[P2] API Key 未删除数量上限存在并发竞态**。`backend/internal/service/api_key_service.go:449` 的独立 Count 与第 590 行的 Create 之间没有同用户原子保护；不同幂等键不会被串行化。已有 199 个 Key、上限 200 时，两次并发创建可得到 201 个。该缺陷来自本次引入的上游功能，不是 Fork 合并丢失。

项目规则与需求两个审查方向均发现上述同一问题，没有确认其他漏合、越界或定制回归。建议在同用户计数和插入间加入数据库事务锁或等价原子准入，并保留并发边界回归。本轮为审查，尚未修改业务实现。

新鲜检查：
- `go test -tags=unit ./internal/handler ./internal/service ./internal/repository ./internal/pkg/accessban ./internal/pkg/kiro ./internal/payment/provider -run 'Inflight|ClaudeReset|APIKeyCreate|Kiro|XorPay|ResetCard|IPBan|ActiveAvailable|ClaudeCodeOnly' -count=1` 退出 0。accessban/provider 在该名称筛选下无匹配测试，另运行下述全包检查。
- `go test -tags=unit ./internal/pkg/accessban -count=1` 和 `go test ./internal/payment/provider -count=1` 均退出 0；provider 现有测试不包含 XorPay 专门用例，未宣称真实支付验证通过。
- `pnpm exec vitest run src/components/account/__tests__/ClaudeResetCreditsCell.spec.ts src/components/common/__tests__/PlatformTypeBadge.spec.ts src/components/keys/__tests__/UseKeyModal.spec.ts src/composables/__tests__/useModelWhitelist.spec.ts src/views/user/__tests__/PaymentView.spec.ts src/views/user/__tests__/SubscriptionsView.spec.ts src/utils/__tests__/platformColors.spec.ts` 退出 0：7 文件、128 项测试。
- 临时 Go overlay 添加同步两次 Count 的 service 层并发测试，不修改项目测试文件。`go test -tags=unit -overlay <临时 overlay.json> ./internal/service -run '^TestReviewConcurrentAPIKeyCountLimit$' -count=1 -v` 按预期退出 1，输出 `successful creates=2; active keys=201; configured limit=200`。这是服务层确定性交错复现，未连接真实数据库。
- 再次运行 `git diff --check` 及全部 97 个变更文本 UTF-8/BOM/替换字符/冲突标记检查通过；本轮仅补充此审查记录。
