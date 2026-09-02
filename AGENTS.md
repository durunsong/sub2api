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
| 已同步基线 | tag **v0.2.0**（官方无 v0.1.174 tag） |
| 当前 VERSION | `backend/cmd/server/VERSION` = **0.2.0** |
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
- **订阅重置卡**：迁移 `224`/`225` 保存永久期限明细并按 `group.default_validity_days` 回填旧计数；有效固定期限支付/兑换/管理员分配重复获得改发 `validity_days` 快照卡且不改到期时间，过期则直接重开；消费放弃余期、清零日周月 USD/token 并从点击时重开；旧 `reset-daily` 映射 1 天卡，`manual_reset_credits` 仅保留兼容镜像
- **订阅可用额度筛选**：管理端订阅状态下拉默认选择 `active_available`（“生效中+没用完”）；后端按有效期及当前日/周/月窗口过滤任一额度已耗尽的订阅，无限额订阅保留
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
backend/internal/service/subscription_service.go
backend/internal/service/user_subscription.go
backend/internal/service/user_subscription_port.go
backend/internal/repository/user_subscription_repo.go
backend/internal/service/payment_fulfillment.go
backend/internal/service/redeem_service.go
backend/internal/handler/subscription_handler.go
backend/internal/handler/dto/types.go
backend/internal/handler/dto/mappers.go
backend/internal/server/routes/user.go
backend/internal/server/middleware/audit_log.go
backend/migrations/224_create_user_subscription_reset_cards.sql
backend/migrations/225_backfill_user_subscription_reset_cards.sql
frontend/src/api/subscriptions.ts
frontend/src/types/index.ts
frontend/src/views/user/SubscriptionsView.vue
frontend/src/i18n/locales/zh/misc.ts
frontend/src/i18n/locales/en/misc.ts
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

订阅重置卡迁移顺序：`224` 建永久明细表 → `225` 按分组默认期限回填历史 `manual_reset_credits`；同步时不得恢复有效固定期限购买/分配顺延，也不得删除旧 `reset-daily` → 1 天卡和兼容镜像。

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
- v0.1.179：国产供应商自适应 API 协议、Composite 支持 Codex/Kimi/智谱/DeepSeek、渠道服务层级/上下文区间倍率、可配置代理探测目标、`/v1/responses/input_tokens` 与用量聚合优化；新增迁移 `226_add_usage_log_effective_model_indexes_notx`、`227_composite_routes_add_cn_providers`、`228_channel_pricing_multipliers`；长上下文计费门控改为分组或账号任一开启即生效；Fork 额外将 Kiro 保留在全部共享平台目录、Composite 路由/定价与调度快照中
- v0.1.180：OpenAI 重置卡按用量阈值自动使用、Fast mode `service_tier` 全链路计费、模型列表响应读取上限、模型广场阶梯/分时定价、Ops 错误返回列表和实验性 OAuth 出站传输插件；新增迁移 `229_plugins` / `230_plugin_artifacts`；Fork 继续保留 Kiro 在共享平台目录与 Wire 注入中的位置，并保留 XorPay、Access Ban、订阅重置卡与 UI 定制
- v0.2.0：分组级 OpenAI 强制/免费 Fast、按模型 reasoning effort 映射和超限拒绝/降级、Kimi 原生 Responses、Claude Fable 5.1、无 call ID 自动化启动，以及 WS terminal event、404 model_not_found、Anthropic fallback beta、API Key 缓存身份和 scheduler passthrough 修复；新增 `232_channel_cache_write_1h_pricing.sql`、`232_group_force_openai_fast.sql`、`232_group_reasoning_effort_over_limit.sql`、`233_group_free_openai_fast.sql`。精确增量 `2ac784c51a5d0925b324efef2ba6b3446c364781..aa236488351eb71e120fc2b6fb32e36b0374c918` 为 60 commits / 148 files / +4,926 / -543，tag object 为 `dd07c4d8d484878e617c945cc8bacc304a5a6560`；Fork 继续保留全部定制
- v0.1.185：价格目录支持 `pricing.override_file` JSON 补丁，长上下文阶梯计价改由目录驱动，Codex 快速模型展示 priority service tier；账号统计统一使用模型定价策略与 DeepSeek 峰谷价格，并修复数据库启动重试、OpenAI WS 空闲连接回收、Codex 图像能力/禁用账号筛选、API Key instructions、ctx_pool 容量错误和 delegation bootstrap；无新增迁移，Fork 继续保留全部定制
- v0.1.184：Codex 路由模型目录与能力同步、公共分组访问限制、原生 compaction 和映射前推理强度用量记录、智谱团队 GLM Coding Plan 用量、国产三家平台 Ollama Cloud 用量窗口、OpenAI 图像工具冷却与可配置 TTFT；新增三个 `231_*` 迁移，并修复 OpenAI/WS/Anthropic/Grok/支付/SMTP/计费链路；Fork 继续保留 Kiro/XorPay/Access Ban、提示词审计、订阅重置卡、Ops、VersionBadge、GLM 标签和 UI/支付定制
- v0.1.183：OpenAI OAuth 配额耗尽 429 按重置时间暂停、Codex `session-id` 粘性与容量溢出绑定保护、Responses custom tool ID 修复；Kimi 并发 403 临时冷却；邮箱换绑别名/并发保护；Antigravity 64K 上限与频道监控 V2 Composite SQL 修复。无新增迁移。官方 tag 内 VERSION 为 0.1.182，本 Fork 按 release tag 设为 0.1.183，并继续保留全部定制
- v0.1.182：OpenAI Responses Lite 统一 OAuth/API Key/HTTP/WS 处理并固定并行工具调用、保留数值精度；OAuth 图片生成原样保留用户提示词；OpenCode Go 正确解析用量重置时长；Anthropic 缓存创建明细去重计费；Antigravity 修正 Sonnet 4.5 兼容路由并保留显式 4.5；Composite 支持 Kimi Code K3；渠道监控 V2 将 Composite 错误归属到真实账号平台；余额充值完成后刷新用户余额。无新增迁移。官方 tag 内 VERSION 仍为 0.1.181，本 Fork 按 release tag 设为 0.1.182，并继续保留全部定制
- v0.1.181：Gemini 工具 schema 清理 `deprecated` 并规范化 enum，Responses Lite 保留 `additional_tools` 所需的 `parallel_tool_calls=false`，Responses 拒绝字段重试按同类 input item 批量移除无效 status，Grok 统一使用官方 CLI User-Agent / 0.2.120 身份；无新增迁移。官方 tag 内 VERSION 仍为 0.1.180，本 Fork 按 release tag 设为 0.1.181，并继续保留全部定制
- v0.1.178：新增 Kimi/智谱/DeepSeek 多协议供应商、渠道监控配额模式、渠道模型分时倍率、OpenAI Team 联动熔断和账号批量设置，并集中修复 Codex/Gemini/Claude/Ops；官方迁移 `224_user_platform_quotas_add_cn_providers`、两个 `225_*` 与 `226_channel_monitor_quota_mode` 与 Fork 订阅重置卡 `224`/`225` 按完整文件名并存；全部 Fork 定制继续保留
- v0.1.177：迁移 `222`/`223` 引入分组用量日汇总并跟随服务端配置时区，分组统计提供 `today` / `yesterday` / `total`；Codex 适配 remote compaction v2 与会话级 beta header，回传 `x-codex-turn-state` 并防止跨账号回显，指纹收敛默认改为 opt-in；修复 Grok 长上下文计费、带版本媒体模型识别和账号页自动刷新偏好。官方 tag 内 VERSION 是 0.1.176，本 Fork 按 release tag 设为 0.1.177，并完整保留 Kiro/XorPay/Access Ban、提示词审计、套餐续期、Ops、UI 品牌等全部定制

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

见 `docs/FORK_VS_UPSTREAM.md` §14。原则：**Kiro + XorPay + Access Ban + 提示词审计 + 套餐续期 + Ops + UI 品牌等全部定制保留**。

`upstream/main` 可能领先于 v0.2.0；同步时以 release tag 为基线，不带入 tag 后的 `main` 内容，并逐文件保留 Fork 模块。

---

## 8. 禁止事项

- 不修改 `.env*`、生产密钥与 token（除非用户明确要求）
- 不恢复 VersionBadge「立即更新」
- 不删除 XorPay / Kiro / Access Ban 代码以简化 merge
- 不自动 `git commit` / `git push`（除非用户要求）
- 新增 SQL 须遵循项目迁移约定并注意 CHECK / Access Ban 约束顺序
- 不删除或弱化重置卡期限明细、购买来源幂等、`224`/`225` 回填、旧接口映射和 `manual_reset_credits` 兼容镜像

---

*维护：Fork 差异变更时同步更新 `docs/FORK_VS_UPSTREAM.md` 与本文件。*
