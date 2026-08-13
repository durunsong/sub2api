# sub2api Fork — AI 协作规则

本仓库是 **[Wei-Shaw/sub2api](https://github.com/Wei-Shaw/sub2api)** 的 Fork（`origin`: durunsong/sub2api）。
修改任何代码前，**必须先阅读** [`docs/FORK_VS_UPSTREAM.md`](docs/FORK_VS_UPSTREAM.md) 了解相对官方的全部差异。

同时遵守仓库根目录 `D:\Dowsure\AGENTS.md` 的安全边界；本文件为 **sub2api 子项目就近规则**，冲突时以本文件 Fork 定制说明为准（不得突破金融/安全红线）。

Cursor 场景下还会加载 [`.cursor/rules/sub2api-fork.mdc`](.cursor/rules/sub2api-fork.mdc) 作为强制提醒。

---

## 1. 项目身份

| 项 | 值 |
|----|-----|
| 上游官方 | https://github.com/Wei-Shaw/sub2api |
| 已同步基线 | tag **v0.1.176**（官方无 v0.1.174 tag） |
| 当前 VERSION | `backend/cmd/server/VERSION` = **0.1.176** |
| 完整差异文档 | **`docs/FORK_VS_UPSTREAM.md`**（相对历史基线；含 Fork 扩展见文档 §8.2 / §12；快捷清单见 `FORK_CUSTOMIZATIONS.md`） |
| 快捷索引 | `FORK_CUSTOMIZATIONS.md` |

**本 Fork 不是官方纯净版**。不得为了「与上游一致」删除 Kiro、XorPay、Access Ban 或 UI/支付定制。

---

## 2. Fork 独有模块（改代码前必确认）

### Kiro 平台（最大模块）

- 后端：`backend/internal/pkg/kiro/`、`kirocooldown/`、`anthropictokenizer/`、`service/kiro_*.go`
- 前端：`api/admin/kiro.ts`、`useKiroOAuth.ts`、紫色平台图标与账号 UI
- 用户端品牌别名：仪表盘、API Key、可用渠道、订阅、购买与错误请求统一把内部 `kiro` 显示为 **Claude**，并支持用 `claude` 搜索；管理端仍显示真实 Kiro
- 迁移：`135`、`145`、`151`–`153`（kiro 相关）
- **计费口径**：`kiro_credits`；cache_read **不重复计入** input_tokens
- **429**：独立冷却，跳过通用 `HandleUpstreamError`

### XorPay 支付

- `backend/internal/payment/provider/xorpay.go`、`TypeXorPay`
- Webhook：`/api/v1/payment/webhook/xorpay`
- 前端展示名：**XorPay（支付宝）**

### 全局访问封禁 Access Ban（官方 v0.1.164 无）

- 已提交：IP/CIDR 封禁 — 迁移 `159`、`IPBanService`、`ip_ban_guard` 中间件、`IpBansView`
- **Fork 扩展**：`accessban` 包 — 规则类型 `ip` / `ua` / `ip_ua` / `email_suffix`；迁移 `160`；注册邮箱格式校验 `registration_email_format.go`
- 管理端批量删用户：`POST /api/v1/admin/users/batch-delete`

### 其他 Fork 定制

- **Grok + Kiro 配额**：迁移 `157` CHECK 必须含 `kiro` 与 `grok`
- **Ops**：`DELETE /ops/errors` 批量删错误日志
- **VersionBadge**：隐藏「立即更新」，保留版本提示与 changelog 链接
- **支付 UX**：用户端隐藏倍率；二维码 payAmount / creditedAmount 分离
- **购买页智普筛选**：OpenAI 协议下名称、分组名或描述含 `GLM` 的套餐显示为 `智普-GLM Plan MAX`，普通 OpenAI 分类必须排除这些套餐
- **UI**：首页/登录靛蓝紫罗兰主题、`logo.svg`、Select/ConfirmDialog 替换原生控件
- **ProxyAdBanner**：已移除

---

## 3. 合并冲突高危区（上游常改 + Fork 已改）

改动或 merge 上游时**先读现有实现**，禁止 blindly 采用上游版本覆盖：

```
backend/cmd/server/wire.go
backend/cmd/server/wire_gen.go
backend/internal/domain/constants.go
backend/internal/handler/admin/account_handler.go
backend/internal/service/account_test_service.go
backend/internal/service/account_usage_service.go
backend/internal/payment/types.go
backend/internal/payment/provider/factory.go
backend/migrations/157_user_platform_quotas_add_grok.sql
backend/migrations/159_create_ip_bans.sql
backend/migrations/160_extend_access_ban_rules.sql
backend/internal/service/ip_ban_service.go
backend/internal/server/middleware/ip_ban_guard.go
backend/internal/service/auth_service.go
frontend/src/types/index.ts
frontend/src/types/payment.ts
frontend/src/i18n/locales/zh.ts
frontend/src/i18n/locales/en.ts
frontend/src/views/user/PaymentView.vue
frontend/src/views/admin/IpBansView.vue
frontend/src/components/common/VersionBadge.vue
backend/internal/server/routes/gateway.go
backend/internal/handler/admin/group_handler.go
```

**Wire 注入顺序**：`kiroTokenProvider` **在前**，`grokTokenProvider` **在后**；`IPBanService` 须注入 `AuthService`。

---

## 4. 迁移红线

`user_platform_quotas.platform` CHECK 约束最终必须为：

```sql
CHECK (platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'kiro', 'grok'))
```

上游 `157` 原版仅加 `grok` 会**丢掉 kiro** — merge 后必须复核。

Access Ban 迁移顺序：`159` 建表 → `160` 扩展 rule_type / ua_pattern，**不可删除或回退**。

---

## 5. 与官方已 merge、不算 Fork 独开发的特性

以下来自官方 v0.1.142+，勿误当 Fork 独有而重复实现或删除：

- 用户端用量分析与管理员对齐（`UsageView`、`request_type`）
- 订阅支付金额显示修复
- `prefer_soonest_reset` 调度、`affiliate` 订阅返佣
- 注册邮箱后缀白名单（`registration_email_suffix_whitelist`）
- Grok 官方 OAuth / gateway（Fork 仅补齐与 Kiro 并存口径）
- OpenAI Spark 链接型影子账号、Grok media 路由、Anthropic dateline 归一化、Sonnet 5 适配
- 代理/兑换码批量删除（`BatchDeleteProxies` / `BatchDeleteRedeemCodes`）— **用户批量删除为 Fork 工作区新增**
- v0.1.158：用户限额批量修改、分组一键复制、Grok 上游端点快捷切换及 OAuth / WebSocket / 图片链路修复
- v0.1.159：OpenAI 独立搜索 API Key 调度修复、客户端 IP 识别统一、Grok Free 函数工具缓存、Stripe 按需加载与 API Key 上游跳转
- v0.1.160：OpenAI 兼容提示词安全审计、Grok 媒体账号隔离、被动图片 namespace 权限修复、S3 配置 step-up TOTP 与审计事件筛选删除
- v0.1.161–v0.1.162：安全开关默认关闭、客户端 IP 解析可配置、异步生图对象存储后台配置、Grok 工具缓存与受保护视频代理、OpenAI/Codex 兼容修复
- v0.1.163：分组级 OpenAI 推理策略、Grok `/responses/compact`、Redis ACL 用户名、优雅关停用量清理、调度/计费/SSE/移动端修复
- v0.1.164：聚合分组、Ollama Cloud 用量同步、支付宝移动端深链，以及 OpenAI/Grok/计费/审计修复；Fork 额外支持聚合分组显式路由到 Kiro
- v0.1.165–v0.1.168：Live WS、面板限流、Passkey、模型广场等官方能力（详见 `FORK_CUSTOMIZATIONS.md`）
- v0.1.169：上游 URL 路径片段校验（安全修复）、代理断流熔断 fail-open、glm-5.2 定价与订阅到期标签修复；Fork 保留 Access Ban / Kiro→Claude / XorPay
- v0.1.170：分组利润控制、API Key 全平台上游倍率探测/自动同步、利润预演工具，以及 Anthropic 部分用量计费、OpenAI/Grok/订阅窗口等修复；Fork 保留 Kiro 粘性 TTL、Access Ban / Kiro→Claude / XorPay 与自定义批量操作交互
- v0.1.171：腾讯天御/阿里云验证码 2.0 三选一门禁、Codex 身份与版本同步、composite 推理强度、OpenAI reset credit 缓存、退款强制确认与 Stripe 幂等、订阅并发锁、token refresh 单飞、计费失败留痕、Messages/WS/调度/模型广场等修复；Fork 叠加保留全部既有定制
- v0.1.172：安全加固、上游响应模型审计（迁移 `194`/`195`）、订阅日额度午夜重置、金额 `NUMERIC(20,8)` 量化，以及网关、模型、验证码、Ops 等修复；官方 tag 内 VERSION 仍为 0.1.171，本 Fork 按 release tag 设为 0.1.172，并保留全部 Fork 定制
- v0.1.173：Grok SSO/refresh/跨实例会话，媒体 Voice、搜索计费与调度，渠道监控 V2，邮箱域名限量注册，以及 Gemini/OpenAI 集中修复；包含迁移 `194`–`206`、`217`–`220` 和安全默认变化，并继续保留 Kiro/XorPay/Access Ban、Kiro→Claude、VersionBadge、GLM 与支付定制
- v0.1.175：Codex OAuth 设备指纹收敛、按上游响应模型计费、大文件备份分卷上传/恢复，以及 OpenAI/Grok/Gemini/WS/审计等集中修复；官方无 v0.1.174 tag。官方 tag 内 VERSION 仍为 0.1.173，本 Fork 按 release tag 设为 0.1.175，并继续保留全部 Fork 定制
- v0.1.176：Grok 4.6 / JWT 订阅档位、分组逐模型定价（`model_pricing` + `long_context_pricing_enabled`）、原生 `POST /x_search`，以及备份 leader 锁、渠道缓存、定价冲突、Responses 探测、Realtime 音频计费等修复；新增迁移 `221`。官方 tag 内 VERSION 仍为 0.1.175，本 Fork 按 release tag 设为 0.1.176。Gateway `/x_search` 保留 Access Ban；分组 `oneof` 仍含 `kiro`

---

## 6. 开发工作流

1. 读任务 → 读 `docs/FORK_VS_UPSTREAM.md` 相关章节 → 读目标文件
2. 结构分析优先 CodeGraph（若可用）
3. 最小 diff；不扩大改动面；不做无关格式化
4. 新增 Fork 定制 → 更新 `docs/FORK_VS_UPSTREAM.md` 与本文件摘要
5. 验证：
   - 后端：`go build ./...`、`go vet ./...`、相关 `go test`
   - 前端：`pnpm vitest run`、`pnpm typecheck`
6. 文本文件 UTF-8，交付前检查中文乱码

---

## 7. 同步上游

见 `docs/FORK_VS_UPSTREAM.md` §14。原则：**Kiro + XorPay + Access Ban + Grok 定制全部保留**。

`upstream/main` 可能领先于 v0.1.176；merge 时以 tag 为基线，逐文件保留 Fork 模块。

---

## 8. 禁止事项

- 不修改 `.env*`、生产密钥与 token（除非用户明确要求）
- 不恢复 VersionBadge「立即更新」
- 不删除 XorPay / Kiro / Access Ban 代码以简化 merge
- 不自动 `git commit` / `git push`（除非用户要求）
- 新增 SQL 须遵循项目迁移约定并注意 CHECK / Access Ban 约束顺序

---

*维护：Fork 差异变更时同步更新 `docs/FORK_VS_UPSTREAM.md` 与本文件。*
