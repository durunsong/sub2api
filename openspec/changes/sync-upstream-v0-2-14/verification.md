# v0.2.12–v0.2.14 同步验证记录

## 范围与版本

- Fork 起点 `2fd1d9ae7`，工作区初始干净。
- 官方基线 v0.2.11 `96f4c115c9749078f90cbf210a01d39baf3f53b6`，目标 v0.2.14 `0363b8cdba8cec3e2ba4b2dbd49c4481143fa55d`，tag object `1400a7b482974d98db5b284a8b2afbe3eaf9aaef`。
- GitHub latest release 查询为 v0.2.14，发布于 2026-10-07；覆盖 v0.2.12、v0.2.13、v0.2.14，未引入 tag 之后的 main。
- 上游增量 47 commits / 176 files / +6,801 / -475，合入 168 个非部署路径。8 个 deploy 路径保持原样，不读取本地秘密配置。官方目标 VERSION 为 0.2.13，Fork 按 release tag 设为 0.2.14。
- 80 个重叠路径、27 个文本冲突文件逐项处理。新 TypeSafe 枚举/配额/调度与 Kiro 并存，充值优惠嵌入 Fork 原有页面布局。

## 保留核验

- 301 个历史 SQL 与 HEAD 字节一致。新增 `241_add_payment_order_bonus_amount.sql` 和 `241_add_typesafe_platform.sql`；后者 quota/Composite CHECK 同时保留 kiro/typesafe，监控平台不扩展。
- 536 个可检查的、仍受版本管理的非重叠 Fork 文件保持字节一致。受保护配置、忽略文件及已从 Git 删除的旧资源不作为此数字的检查对象。
- 88 个非重叠上游路径中，86 个与官方字节一致；另两个为新 TypeSafe SQL 与对应测试，唯一兼容调整是保留 Kiro。
- 独立只读审查未发现可确认的定制丢失或合并回归。重点核对 Kiro/TypeSafe、Access Ban、充值金额、XorPay、订阅发卡/退款、ent 及 UI 品牌别名；10 个核心 Fork 文件（订阅/退款/兑换/AuthService/XorPay/Wire/VersionBadge 等）与 HEAD 完全一致。
- 原 Fork 增删行对比的差异仅涉及平台列表扩充、充值促销 UI/条件更新、已有依赖安全升级及测试适配；Kiro credits/cache_read、冷却、上游 ID、冻结计价时刻、永久重置卡、来源幂等、active_available、提示词审计、Ops、GLM 分类、用户端 Claude 别名和隐藏在线更新均保留。

## 先失败、后通过的证据

- 先只引入官方 `easypay_notify_security_test.go`，运行 `go test ./internal/payment/provider -run TestEasyPayNotify -count=1`，旧实现退出 1：签名复用伪造回调、订单 URL 重放、非标准参数三项未拒绝；合法回调通过。合入实现后同命令退出 0。
- 先引入合并后的 AmountInput 回归，在旧实现运行 `pnpm exec vitest run src/components/payment/__tests__/AmountInput.spec.ts`，3 项促销展示用例失败，9 项既有/兼容用例通过；合入后 12 项全部通过。
- 保留人民币快捷金额上限/非法输入回退用例，新增充值页面赠金/折扣两个集成回归：CNY 实付、USD 到账、手续费、促销 Markdown 清理与原布局共存。测试环境 i18n 使用已有上游测试模式，避免 runtime-only 翻译器缺少编译器影响金额插值断言。
- `TestGatewayRoutesAccessBanPrecedesAdmission` 增加 `/v1/systemone`，验证被封禁请求在鉴权之前返回 403。
- 初次全量前端测试 6 项失败，原因是旧测试仍假定 6 个默认配额平台及旧 Codex features 文本。更新两份测试，保留 Kiro 断言并增加 TypeSafe/API Key 模型发现期望；相关 44 项定向测试及全量复跑通过。

## 最终命令与结果

所有命令在各自真实项目目录执行；后端源码最终编辑后完成以下检查：

- `go test ./... -count=1`：退出 0，53 个有测试的包，service 包 145.143 秒。
- `go test -tags=unit ./... -count=1`：退出 0，61 个有测试的包，service 包 194.818 秒。
- `go build ./...`、`go vet ./...`：均成功，无错误输出。
- `golangci-lint run ./...`：退出 0，`0 issues.`。
- `go build -tags embed -o <临时输出文件> ./cmd/server`：退出 0，包含本次前端构建资源；未运行二进制。
- `CI=true pnpm install --frozen-lockfile --ignore-scripts`：退出 0，锁文件匹配；使用 Axios 1.20.0、Vue 3.5.43 及上游 source-map-js 安全覆盖，保留已有 hasown 2.0.4 升级。
- `pnpm run test:run` 最终复跑：退出 0，351 个文件、2,695 项测试全部通过。
- `pnpm run typecheck`、`pnpm run lint:check`：最终测试适配后复跑均退出 0。
- `pnpm run build`：退出 0，内置 i18n 3 项检查通过；后续只调整测试和文档，没有修改产品代码。
- `sh backend/scripts/resolve-version.sh`：输出 `0.2.14`。
- `git diff --check` 和最终文本 UTF-8/BOM/U+FFFD/冲突标记检查通过；没有暂存、提交、推送或创建 PR。

## 交付文件

168 个上游路径，3 份 Fork 说明（AGENTS.md、FORK_CUSTOMIZATIONS.md、docs/FORK_VS_UPSTREAM.md），4 份额外兼容回归（Gateway Access Ban、默认配额、Codex 配置、充值页面），5 份本次 OpenSpec 文档；总计 180 个文本路径。历史迁移及全部部署路径保持原样。

## 运行限制与剩余风险

- 未启动服务、执行迁移、真实 PostgreSQL/Redis 集成测试、供应商 OAuth/TypeSafe 调用、真实支付或部署；本地回归使用现有测试设施，不代表线上联调已完成。
- 两个新 241 SQL 会在未来启动新版服务时被现有 runner 自动执行，需先备份并在测试库核验。新平台行与优惠订单产生后，不能假设直接降级旧程序即可回退。
- 密码重置 token 改哈希存储并原子消费，升级前未使用的旧链接需重新申请。新装管理员邮箱随机化、密码 8–72 字节校验不影响已有管理员。
- v0.2.11 已记录的 API Key 数量上限并发竞态，v0.2.14 未修改相关创建逻辑，本次未额外修复；详见前次 verification.md 的二次审查记录。
- 前端构建仍有大 chunk 与 Browserslist 数据过期提示，不影响退出状态，未扩大范围升级无关依赖。
- OpenSpec CLI/CodeGraph 当前不可用，沿用项目既有 spec-driven 文档结构和 Git/rg；未宣称 OpenSpec CLI validate 通过。
