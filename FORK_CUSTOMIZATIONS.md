# sub2api Fork 自定义功能清单

> **用途**：本仓库是 `Wei-Shaw/sub2api`（经 `durunsong/sub2api` 中转）的 Fork，在官方版本之上叠加了一批自定义功能。
> **AI 编程前必读**：修改本项目任何代码前，先读本文件，确认改动是否触碰以下自定义模块；
> 与上游同步（merge/rebase 新版本）时，本文件列出的文件全部属于「需保留、易冲突」区域。
>
> - 统计基准：当前代码相对官方 **v0.1.141** tag 的差异（246 个文件：约 72 新增 + 174 修改）。
> - 维护方式：新增自定义功能后，请在对应章节补登文件与说明。
> - 最近更新：2026-06-30（同步至官方 v0.1.141 后整理）。

---

## 零、改错风险最高的「共享文件」（上游也会动）

以下文件**既被你改过、上游也常改**，是合并冲突与 AI 误改的高发区，改动前务必先看现有实现：

| 文件 | 你的改动性质 | 注意 |
|------|------------|------|
| `backend/cmd/server/wire.go` / `wire_gen.go` | 注入 Kiro / Grok 的 TokenProvider、UsageService | 构造函数签名含 `kiroTokenProvider`+`grokTokenProvider`，顺序：kiro 在前、grok 在后 |
| `backend/internal/domain/constants.go` | 新增 `PlatformKiro`、`PlatformGrok` | 平台常量列表，多处 switch 依赖 |
| `backend/internal/handler/admin/account_handler.go` | 新增 Kiro / Grok 的可用模型分支 | 两个 `if account.Platform == ...` 分支并存 |
| `backend/internal/service/account_test_service.go` | struct 字段 + 构造参数含 kiro/grok | |
| `backend/internal/service/account_usage_service.go` | `ProvideAccountUsageService` 带 Kiro | 上游用 `NewAccountUsageService`，本 fork 用 Provide 版注入 Kiro |
| `backend/internal/payment/types.go` / `provider/factory.go` | 注册 `TypeXorPay` | 支付类型枚举与工厂 |
| `backend/internal/i18n` 与 `frontend/src/i18n/locales/{zh,en}.ts` | Kiro / XorPay 文案 | |
| `frontend/src/types/index.ts` / `types/payment.ts` | Kiro / XorPay 类型 | |
| `backend/migrations/157_user_platform_quotas_add_grok.sql` | **已改**：CHECK 约束含 kiro+grok | 见第四章「迁移红线」 |

---

## 一、Kiro 平台完整支持（最大模块）

把 AWS Kiro（CodeWhisperer/Q）作为一个上游账号平台接入，含 OAuth、令牌刷新、请求翻译、缓存模拟、粘性会话、计费、用量查询、冷却等。

### 后端 — 独立包（全部新增）

- `backend/internal/pkg/kiro/` — 核心 SDK：`oauth.go`、`signature.go`、`translator.go`（请求/响应翻译）、`models.go`、`fingerprint.go`、`websearch.go` / `websearch_stream.go`
- `backend/internal/pkg/kirocooldown/` — 429 冷却存储 `store.go`
- `backend/internal/pkg/anthropictokenizer/` — 给 Kiro 估算 token 用的 Anthropic 分词器（`tokenizer.go` + `claude.json` 词表 + `NOTICE.md`）

### 后端 — service 层（全部新增 kiro_*.go）

- OAuth/令牌：`kiro_oauth_service.go`、`kiro_token_provider.go`、`kiro_token_refresher.go`、`kiro_usage_fetcher.go`
- 运行时：`kiro_runtime.go`、`kiro_runtime_state.go`、`kiro_http_helpers.go`、`kiro_error_classifier.go`
- 能力：`kiro_cache_emulation.go`（缓存模拟）、`kiro_websearch.go`
- handler：`backend/internal/handler/admin/kiro_oauth_handler.go`

### 后端 — 数据库迁移（新增）

- `135_add_group_kiro_cache_emulation.sql`
- `145_allow_kiro_user_platform_quotas.sql` — **把 kiro 加入配额表 CHECK 约束**（与 157 相关，见第四章）
- `151_add_group_kiro_auto_sticky.sql`、`152_..._sticky_session_ttl.sql`、`153_..._endpoint_mode.sql`、`153_add_usage_log_kiro_credits.sql`

### 前端

- `frontend/src/api/admin/kiro.ts`、`frontend/src/composables/useKiroOAuth.ts`
- 紫色主题与图标：`frontend/src/utils/platformColors.ts`、`components/common/PlatformIcon.vue` / `PlatformTypeBadge.vue`
- 账号管理弹窗、平台用量统计接入 Kiro（见 `AccountsView.vue`、`admin/UsageView.vue` 等）

### 关键口径（勿改错）

- Kiro 计费用 **credits / 自定义单价**，cache_read token **不重复计入** input_tokens。
- 429 走独立冷却，**跳过** `HandleUpstreamError`，并双向同步 DB `rate_limit_reset_at`。
- 支持 group 级粘性会话（auto sticky / ttl）与推理端点模式（q / krs）。

## 二、XorPay 支付服务商（支付宝扫码）

新增 XorPay 作为支付渠道，二维码扫码支付。

- 后端：`backend/internal/payment/provider/xorpay.go`（新增），在 `payment/types.go`（`TypeXorPay`）、`provider/factory.go`、`service/payment_config_providers.go`、`service/payment_order_lifecycle.go`、`handler/payment_webhook_handler.go` 接线
- 前端：`components/payment/` 下 `ProviderCard.vue`、`PaymentMethodSelector.vue`、`PaymentQRDialog.vue`、`PaymentStatusPanel.vue`、`paymentFlow.ts`、`providerConfig.ts`；`views/user/PaymentQRCodeView.vue`、`views/admin/orders/AdminOrdersView.vue`
- 注意：与上游退款逻辑（`payment_refund.go` 的 `QueryAndFinalizeRefund`）并存，互不影响。

## 三、Grok 平台口径补齐

`ca72bba5` 补齐 Grok 与 Kiro 的平台口径（常量、配额、约束）。Grok 模型分支在 `account_handler.go`，配额 CHECK 见 157 迁移。

## 四、迁移红线（数据库 CHECK 约束顺序）

`user_platform_quotas.platform` 的 CHECK 约束被多次 DROP+重建，**执行顺序决定最终成员**：
- `142`（上游建表）：4 基础平台
- `145`（你的）：+ `kiro`
- `157`（上游 v0.1.141 新增）：原版只写 `grok` → **会丢掉 kiro**

**已修复**：157 改为 `CHECK (... 'antigravity', 'kiro', 'grok')`，kiro+grok 都保留。
**今后同步上游若 157 被覆盖，务必重新确认约束含 kiro。**

## 五、配色与 UI 改造

- `30d6789a`：首页 + 登录/注册页改用**靛蓝紫罗兰**配色并加交互动效（`views/HomeView.vue`、`components/layout/AuthLayout.vue`）
- 全局把原生 `<select>`、原生 `confirm/alert` 替换为自定义 `Select.vue` / `ConfirmDialog`
- 移除代理广告横幅组件（`ProxyAdBanner` 相关）

## 六、其他自定义点

- `feat: apply affiliate rebate to subscription payments` — 订阅支付应用推广返佣
- `feat(scheduling): prefer soonest reset account selection` — 可选「优先最快重置」账号调度
- `feat(admin): 管理员删除错误日志` — ops 错误日志删除
- `fix(auth): 邮箱绑定后缀白名单`

---

## 同步上游新版本的标准流程（实测可行）

1. `git remote add upstream https://github.com/Wei-Shaw/sub2api`（已添加可跳过）
2. `git fetch upstream tag vX.Y.Z --no-tags`
3. `git merge vX.Y.Z`（**不要自动提交**）
4. 冲突统一原则：**Kiro/XorPay/Grok 你的代码与上游代码都保留**，kiro 在前、grok 在后
5. 重点复查：`wire_gen.go` 构造签名、`account_handler.go` 平台分支、`157`/配额 CHECK 约束、支付相关文件
6. 验证：后端 `go build ./... && go vet ./... && go test -tags=unit ./internal/...`；前端 `pnpm install --frozen-lockfile && pnpm vitest run && pnpm typecheck`
7. UTF-8 乱码复核后再交付

