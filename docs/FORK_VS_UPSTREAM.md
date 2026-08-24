# Fork 相对官方仓库差异说明

> **上游官方仓库**：[Wei-Shaw/sub2api](https://github.com/Wei-Shaw/sub2api)
> **本 Fork 远程**：`origin` → `durunsong/sub2api`（中转/部署用）
> **对比基准**：官方 tag **`v0.1.181`**（2026-08-25；官方无 v0.1.174 tag）
> **本 Fork 当前版本**：`backend/cmd/server/VERSION` = **0.1.181**
> **统计**：历史相对 v0.1.164 的大盘差异见下文；v0.1.165–v0.1.181 增量以根目录 `FORK_CUSTOMIZATIONS.md` 为准
> **当前工作区**：已合入官方 v0.1.181，并保留 Kiro / XorPay / Access Ban / 提示词审计 / 套餐续期 / Ops / UI 品牌等全部定制
> **维护**：新增 Fork 定制后，请同步更新本文与根目录 `AGENTS.md` 摘要。

---

## 1. 仓库关系与阅读顺序

| 角色 | 说明 |
|------|------|
| **Wei-Shaw/sub2api** | 官方上游，功能与版本以该仓库 Releases / tag 为准 |
| **durunsong/sub2api** | 本 Fork 的 `origin`，在官方之上叠加 Kiro / XorPay / UI 等定制 |
| **本文档** | AI 与开发者改代码前的**完整差异清单** |
| **`AGENTS.md`**（项目根） | AI 协作**强制摘要**与禁区，改代码前必读 |
| **`FORK_CUSTOMIZATIONS.md`**（项目根） | 历史清单，已收敛到本文；保留作快捷索引 |

**注意**：本 Fork 已同步官方至 **v0.1.181**，但 **`upstream/main` 仍可能领先**。与官方同步时以目标 release **tag** 为准，不带入 tag 后的 `main` 内容；merge `main` 前务必先读本文 Fork 定制章节，禁止 blindly 采用上游覆盖 Kiro / XorPay / Access Ban 等模块。

---

## 2. 差异总览（相对 v0.1.146）

| # | 模块 | 类型 | 一句话 |
|---|------|------|--------|
| 1 | **Kiro 平台** | 新增 | AWS Kiro 全链路：OAuth、翻译、缓存模拟、粘性会话、计费 credits、429 冷却 |
| 2 | **XorPay 支付** | 新增 | 支付宝扫码聚合支付，含 webhook、管理端配置、用户端 QR 流程 |
| 3 | **Grok + Kiro 配额口径** | 修改 | 平台常量、默认配额、`157` 迁移 CHECK 约束同时含 `kiro` + `grok` |
| 4 | **UI / 品牌** | 修改 | 首页/登录注册靛蓝紫罗兰主题、`logo.svg`、Kiro 紫色平台图标 |
| 5 | **组件规范化** | 修改 | 原生 `<select>` → `Select.vue`；原生 `confirm/alert` → `ConfirmDialog` |
| 6 | **支付 UX** | 修改 | 隐藏用户端倍率；XorPay 前端展示为「XorPay（支付宝）」；二维码金额展示优化 |
| 7 | **版本弹窗** | 修改 | 管理员版本下拉**隐藏「立即更新」**，保留版本提示与更新日志链接 |
| 8 | **Ops 错误日志** | 新增 | 管理员按条件批量删除 ops error logs |
| 9 | **全局访问封禁（Access Ban）** | 新增 | v0.1.146 之后 Fork 自研：IP/CIDR 封禁 + 中间件拦截注册/API；扩展 UA / IP+UA / 邮箱后缀 |
| 10 | **代理广告** | 删除/禁用 | 移除 `ProxyAdBanner` 相关展示 |
| 11 | **构建与部署** | 修改 | Docker 使用 npmmirror、Kiro 环境变量、`.dockerignore` 文档打包调整 |
| 12 | **资源与文档** | 新增 | Kiro 截图、微信群二维码；`docs/FORK_VS_UPSTREAM.md`、`AGENTS.md`；`docs/legal/` 合规文案 |
| 13 | **管理端批量删用户** | 新增 | `POST /api/v1/admin/users/batch-delete`，UsersView 多选批量删除（跳过 admin） |
| 14 | **订阅重置卡** | 新增 | 固定期限支付/正数订阅兑换重复购买改发永久重置卡；按购买时 `validity_days` 快照消费并从点击时重开周期 |

**以下能力已在官方 v0.1.142+ 中，本 Fork 通过同步拥有，不算 Fork 独有开发**（合并时保留了 Kiro/XorPay 定制）：

- v0.1.181 官方能力：Gemini 工具 schema 递归移除 `deprecated`、规范化标量 enum 并丢弃非标量 enum；Responses Lite 识别 `additional_tools` 并保留必要的 `parallel_tool_calls=false`；Responses 拒绝字段重试一次清理同类 input item 的无效 `status`；Grok OAuth、模型探测与计费探针统一使用官方 CLI User-Agent，CLI 身份更新到 `0.2.120`。无新增迁移；官方 tag 内 VERSION 仍为 0.1.180，本 Fork 按 release tag 设为 0.1.181，并保留全部 Fork 定制。
- v0.1.180 官方能力：OpenAI 重置卡按用量阈值自动使用、Responses/Chat/WS Fast mode `service_tier` 传递与按上游实际档位计费、可配置模型列表响应读取上限、模型广场上下文阶梯/渠道分时定价、Ops 错误详情返回列表，以及实验性 OAuth 出站传输插件；新增迁移 `229_plugins.sql` / `230_plugin_artifacts.sql`。本 Fork 在新插件与自动重置服务 Wire 注入之外继续注入 Kiro、Kiro OAuth 与 IP Ban，并保留 Kiro 作为 Composite、渠道匹配和调度快照的具体平台。
- v0.1.179 官方能力：国产供应商自适应 API 协议、Composite Codex/Kimi/智谱/DeepSeek 路由、渠道 fast/flex 与上下文区间倍率、Anthropic Fast 计费、可配置代理探测、`/v1/responses/input_tokens` 和用量聚合优化；新增迁移 `226_add_usage_log_effective_model_indexes_notx.sql`、`227_composite_routes_add_cn_providers.sql`、`228_channel_pricing_multipliers.sql`。长上下文计费门控由“分组且账号开启”改为“任一开启”；本 Fork 将 Kiro 合并进共享平台目录、Composite 与调度快照并继续保留全部定制。
- 用户端用量分析与管理员端对齐（`UsageView` 重构、`request_type` 筛选等）
- 订阅支付金额显示修复
- `prefer_soonest_reset` 账号调度可选策略
- 订阅支付推广返佣（affiliate rebate）
- 注册/绑定邮箱后缀白名单（`registration_email_suffix_whitelist`）
- Grok 平台官方支持（OAuth、Responses 转发等）
- OpenAI Spark 链接型影子账号（`gpt-5.3-codex-spark` 独立配额调度）
- Grok media routing / 图像上传转换
- Anthropic OAuth dateline 指纹归一化与 Claude Sonnet 5 适配
- 订阅撤销软删除缓存失效修复、账号列表 Count clone 修复、Codex OAuth reasoning 保留修复
- v0.1.146 官方能力：账号请求头覆写、账号数据拖拽/批量导入增强、API key 并发统计、EasyPay 自定义支付方式、订阅 CNY 预览、Redis 索引清理加固、OpenAI Responses 入站端点归一化与新模型/定价更新
- v0.1.155 官方能力：Grok 渠道健康监控、Web SSO 批量导入、免费配额滚动 24 小时估算与 Free 计划徽标
- v0.1.155 官方能力：系统日志 Host 筛选、可选 Server Timing、调度器全量重建合并与到期事件修复
- v0.1.155 官方能力：OpenAI HTTP/2 keep-alive、reset credits、账号级长上下文计费，以及 Responses namespace / 图像生成链路修复
- v0.1.156 官方能力：OpenAI Agent Identity、Responses/ChatCompletions 兼容与流式边界修复、图片工具链和首包超时加固
- v0.1.156 官方能力：Grok OAuth 池健康与刷新容错、账号重复创建保护、调度缓存生命周期与批量重建优化
- v0.1.156 官方能力：账号/Key ID 列、OpenAI 认证模式提示、DataTable 缓存修复及 Server Timing 扩展
- v0.1.157 官方能力：操作审计日志、会话 IP/UA 绑定、敏感操作 step-up 2FA 与管理员角色提升保护
- v0.1.157 官方能力：异步图片生成任务、对象存储结果、图片输入 token/费用独立计费
- v0.1.157 官方能力：上游计费倍率探测与调度、API Key 倍率自省、Grok 自定义上游与渠道监控复制
- v0.1.158 官方能力：用户限额批量修改、分组一键复制、Grok 上游端点快捷切换，以及 Grok OAuth / WebSocket / 图片链路修复
- v0.1.159 官方能力：OpenAI 独立搜索 API Key 调度修复、真实客户端 IP 识别统一、Grok Free 函数工具缓存、Stripe 按需加载与 API Key 上游跳转
- v0.1.160 官方能力：OpenAI 兼容提示词安全审计、Grok 媒体账号隔离、被动图片 namespace 权限修复、S3 配置 step-up TOTP 与审计事件筛选删除
- v0.1.161 官方能力：敏感操作/会话绑定安全开关、Grok 受保护视频代理与媒体链路修复、OpenAI WS/Responses 流式修复、Docker 跨架构构建
- v0.1.162 官方能力：客户端 IP 解析可配置、异步生图对象存储后台配置、Grok 客户端工具缓存、更新检查 GitHub Token、Codex/计费/订阅展示修复
- v0.1.163 官方能力：分组级 OpenAI 推理策略、Grok `/responses/compact`、Redis ACL 用户名、优雅关停用量清理，以及调度、计费、SSE、移动端集中修复
- v0.1.164 官方能力：聚合分组、Ollama Cloud 用量同步、支付宝移动端深链，以及 OpenAI/Grok/计费/审计修复；本 Fork 将 Kiro 补入聚合分组合法目标、模型候选和渠道定价匹配
- v0.1.169 官方能力：上游 URL 路径片段闭集校验（GHSA-vrxq-qm4h-6hgg）、代理断流熔断 fail-open、glm-5.2 定价、订阅到期标签与套餐标题可读性、SMTP/Passkey/Qwen3Guard 辅助字段等修复；本 Fork 在 `gateway.go` 保留 Access Ban 的同时挂上 `guardResponsesSubpath`
- v0.1.170 官方能力：分组级利润控制、API Key 全平台上游倍率探测与可选自动同步、`profit-preview` 离线预演、内容审核代理与账号全结果批量选择；同时合入 Anthropic 流中断部分用量计费、OpenAI/Grok 转发、订阅窗口和 SMTP 等修复；本 Fork 继续保留 Kiro 粘性 TTL、Access Ban、XorPay、Kiro→Claude 别名和自定义确认弹窗
- v0.1.171 官方能力：腾讯天御/阿里云验证码 2.0 与 Turnstile 三选一门禁、Codex 出站身份统一和版本同步、composite 推理强度、OpenAI reset credit 缓存、退款 `require_force` 与 Stripe 幂等、订阅并发锁、token refresh 单飞、计费失败用量留痕、Messages 临时错误 failover、WS/调度取消和模型广场图片价格修复；本 Fork 在这些能力上继续叠加 Kiro/XorPay/Access Ban、续期定制、提示词审计与 UI 品牌定制
- v0.1.172 官方能力：安全加固、上游响应模型审计与迁移 `194`/`195`、订阅日额度午夜重置、金额按 `NUMERIC(20,8)` 量化，以及网关、OpenAI/Grok/Antigravity 模型兼容、腾讯验证码、Ops 日志、代理超时和模型广场等修复；官方 tag 内 VERSION 仍为 0.1.171，本 Fork 按 release tag 设为 0.1.172，并完整保留 Kiro/XorPay/Access Ban、Kiro→Claude、VersionBadge、支付与订阅定制
- v0.1.173 官方能力：Grok 邮箱密码 SSO、refresh token 重认证与 Redis 跨实例会话，媒体 Voice、视频/搜索计费和调度，渠道监控 V2、邮箱域名限量注册，以及 Gemini/OpenAI 集中修复；新增迁移 `194`–`206`、`217`–`220`，默认启用 Grok 免费额度软门禁、默认隐藏 V2 吞吐量并由迁移 `220` 备份后清理非 Grok/非 composite 视频价格，渠道监控则保持 V1、V2 显式启用；本 Fork 继续保留 Kiro/XorPay/Access Ban、Kiro→Claude、VersionBadge、GLM 与支付定制
- v0.1.175 官方能力：Codex OAuth 设备指纹收敛、按上游响应模型计费、大文件备份分卷上传与恢复，以及 OpenAI/Grok/Gemini/WS/审计/调度等集中修复；官方无 v0.1.174 tag，官方 tag 内 VERSION 仍为 0.1.173；本 Fork 按 release tag 设为 0.1.175，备份页保留 ConfirmDialog/密码弹窗并叠加分卷下载，同时完整保留 Kiro/XorPay/Access Ban、Kiro→Claude、VersionBadge、GLM 与支付定制
- v0.1.176 官方能力：Grok 4.6 / JWT 订阅档位识别、分组逐模型定价与长上下文阶梯开关、原生 `POST /x_search`，以及备份 leader 锁、渠道缓存、定价冲突、Responses 探测、Realtime 音频计费等修复；新增迁移 `221`。官方 tag 内 VERSION 仍为 0.1.175；本 Fork 按 release tag 设为 0.1.176。Gateway `/x_search` 保留 Access Ban，分组 `oneof` 仍含 `kiro`，Kiro 缓存模拟 UI 与逐模型定价 UI 并存，并完整保留 Kiro/XorPay/Access Ban、Kiro→Claude、VersionBadge、GLM 与支付定制
- v0.1.177 官方能力：迁移 `222`/`223` 建立分组用量日汇总并切换为服务端配置时区，分组页与仪表盘统计提供 `today` / `yesterday` / `total`；Codex remote compaction v2 使用会话级 `remote_compaction_v2` beta header，原生 v2 保留 `/responses` 路由，`x-codex-turn-state` 跨 HTTP/SSE/WS 回传并通过来源记录阻止跨账号回显，指纹收敛默认关闭且仅显式 opt-in；同时修复 Grok 长上下文计费、带版本媒体模型识别和账号页自动刷新偏好。官方 tag 内 VERSION 是 0.1.176；本 Fork 按 release tag 设为 0.1.177，并完整保留 Kiro/XorPay/Access Ban、提示词审计、套餐续期、Ops、UI 品牌及既有支付定制
- v0.1.178 官方能力：Kimi/智谱/DeepSeek 多协议供应商、渠道监控配额模式、渠道模型分时倍率、OpenAI Team 联动熔断、OpenAI 账号批量设置，以及 Codex/Gemini/Claude/WS/Ops 集中修复；官方迁移 `224_user_platform_quotas_add_cn_providers.sql`、两个 `225_*` 和 `226_channel_monitor_quota_mode.sql` 与 Fork 订阅重置卡 `224`/`225` 按完整文件名并存；本 Fork 完整保留全部既有定制。

v0.1.160 同步时额外补齐 `securityaudit.ProviderSet` 中 `*PromptService` 到 `PromptAdminService` 的 Wire 接口绑定，确保当前 Fork 可重新生成 `wire_gen.go`；其余提示词审计实现保持官方版本。

---

## 3. Kiro 平台（最大定制模块）

### 3.1 功能范围

- 平台标识：`PlatformKiro`（`backend/internal/domain/constants.go`）
- OAuth / IDC 认证、令牌刷新、主动用量查询
- Anthropic Messages 请求/响应翻译（`backend/internal/pkg/kiro/translator.go`）
- Group 级：**缓存模拟**、**自动粘性会话**、**粘性 TTL**、**推理端点模式**（`q` / `krs`）
- 计费：`usage_logs.kiro_credits` 字段；cache_read **不重复计入** input_tokens
- 429：独立冷却（`kirocooldown`），**跳过**通用 `HandleUpstreamError`，与 DB `rate_limit_reset_at` 双向同步
- **count_tokens**：Kiro 直连账号（OAuth / 无 base_url 的 API Key）走本地 `estimateKiroInputTokens`，**不打上游**，避免 `jwt auth is not yet supported on count_tokens` → OAuth 401 临时不可调度
- **已同步 nianzs/main（上次合入之后）的 Kiro 增量**：图片视觉 token、`KiroEndpointModeAuto`、API Key 直连/中转分流、`profile_arn` 解析回填、External IdP OAuth 后端、translator tool/filePath/空 input、直连调度 `isKiroDirectModeAccount` 等（前端 External IdP 两阶段引导 UI 仍待补齐）

### 3.2 后端 — 新增文件（独立包）

```
backend/internal/pkg/kiro/
  oauth.go, oauth_test.go, oauth_invalid_grant_test.go
  signature.go, translator.go, translator_test.go
  models.go, models_test.go, fingerprint.go, fingerprint_test.go
  websearch.go, websearch_test.go, websearch_stream.go, websearch_stream_test.go

backend/internal/pkg/kirocooldown/
  store.go, store_test.go

backend/internal/pkg/anthropictokenizer/
  tokenizer.go, tokenizer_test.go, claude.json, NOTICE.md

backend/internal/handler/admin/
  kiro_oauth_handler.go
  group_handler_kiro_validation_test.go

backend/internal/service/
  kiro_oauth_service.go, kiro_oauth_service_test.go
  kiro_token_provider.go, kiro_token_provider_test.go
  kiro_token_refresher.go, kiro_usage_fetcher.go
  kiro_runtime.go, kiro_runtime_state.go, kiro_runtime_state_test.go
  kiro_runtime_state_integration_test.go
  kiro_http_helpers.go, kiro_http_helpers_test.go
  kiro_error_classifier.go, kiro_error_classifier_test.go
  kiro_cache_emulation.go, kiro_cache_emulation_test.go
  kiro_websearch.go, kiro_websearch_test.go
  kiro_alignment_test.go, kiro_alignment_unit_test.go
  kiro_credits_usage_log_test.go, kiro_mapping_fallback_test.go
  kiro_session_test.go, kiro_sticky_session_test.go
  account_test_service_kiro_test.go, account_test_service_kiro_apikey_fallback_test.go
  account_usage_service_kiro_apikey_test.go
  account_kiro_credit_unit_price_test.go
```

### 3.3 后端 — 修改的接线文件（节选）

| 文件 | 改动 |
|------|------|
| `backend/cmd/server/wire.go` / `wire_gen.go` | 注入 `kiroTokenProvider`（**在 grok 之前**） |
| `backend/internal/handler/admin/account_handler.go` | Kiro 可用模型分支 |
| `backend/internal/service/account_test_service.go` | 构造参数含 Kiro/Grok provider |
| `backend/internal/service/account_usage_service.go` | `ProvideAccountUsageService` 注入 Kiro |
| `backend/internal/handler/gateway_handler*.go` | Kiro 路由与转发 |
| `backend/internal/service/gateway_service.go` 等 | Kiro 运行时调度 |
| `backend/ent/schema/group.go` | Kiro group 字段 |
| `backend/internal/repository/usage_log_repo.go` | `kiro_credits` 读写 |

### 3.4 数据库迁移（Fork 新增）

| 文件 | 作用 |
|------|------|
| `135_add_group_kiro_cache_emulation.sql` | Group 缓存模拟字段 |
| `145_allow_kiro_user_platform_quotas.sql` | CHECK 约束加入 `kiro` |
| `151_add_group_kiro_auto_sticky.sql` | 自动粘性开关 |
| `152_add_group_kiro_sticky_session_ttl.sql` | 粘性 TTL |
| `153_add_group_kiro_endpoint_mode.sql` | 端点模式 q/krs |
| `153_add_usage_log_kiro_credits.sql` | 用量表 kiro_credits |
| `157_user_platform_quotas_add_grok.sql` | **已改**：CHECK 含 `anthropic, openai, gemini, antigravity, kiro, grok` |

### 3.5 前端

| 路径 | 作用 |
|------|------|
| `frontend/src/api/admin/kiro.ts` | Kiro 管理 API |
| `frontend/src/composables/useKiroOAuth.ts` | OAuth 流程 |
| `frontend/src/utils/platformColors.ts` | Kiro 紫色主题 |
| `frontend/src/components/common/PlatformIcon.vue` / `PlatformTypeBadge.vue` | 平台图标 |
| `frontend/src/types/index.ts` | `GroupPlatform` / `AccountPlatform` 含 `'kiro'`；Kiro 相关字段 |
| 账号/分组/用量相关 View 与 Modal | 创建编辑账号、用量统计、配额等 |

#### 用户端 Kiro → Claude 品牌别名（2026-07-14）

用户端采用展示别名，覆盖以下页面：

| 页面 | 路由 | 改动记录 |
|------|------|----------|
| 用户仪表盘 | `/dashboard` | 平台用量统计中的 `kiro` 显示为 Claude |
| API Key 管理 | `/keys` | 分组名称、描述、徽章、图标和下拉选项显示 Claude；搜索 `claude` 可命中内部 Kiro 分组 |
| 可用渠道 | `/available-channels` | 渠道、平台、分组及模型区域使用 Claude 名称、图标和配色；搜索 `claude` 可命中 Kiro 渠道 |
| 我的订阅 | `/subscriptions` | 订阅分组名称、描述、平台徽章、卡片边框和操作按钮使用 Claude 展示 |
| 购买 / 续费 | `/purchase` | Kiro 套餐分类显示为 `Claude(Max 5x)`；OpenAI 协议下名称、分组名或描述含 `GLM` 的套餐独立显示为 `智普-GLM Plan MAX`，并从普通 OpenAI 分类排除；套餐卡片、确认页、续费弹窗和已有订阅使用 Claude 名称与配色 |
| 用量—错误请求 | `/usage` | 错误列表和详情的平台字段显示 Claude |

实现边界：

- 后端、API 请求参数、数据库及前端内部筛选值继续使用 `kiro`，不得改为 `claude`。
- 管理端继续展示真实 Kiro 名称、官方 Kiro 图标和紫色主题。
- 用户侧动态分组、套餐名称和描述中的独立单词 `Kiro`（忽略大小写）替换为 Claude。
- 通用搜索选项使用隐藏搜索关键字，同时包含内部值 `kiro` 和展示值 `claude`；用户下拉中不展示 Kiro。
- 错误请求的平台字段使用 Claude 别名，但上游错误消息、响应体等诊断原文保持不变，避免影响排障。

主要实现文件：

- `frontend/src/utils/platformColors.ts`
- `frontend/src/components/common/{GroupBadge,GroupOptionItem,Select}.vue`
- `frontend/src/views/user/{KeysView,AvailableChannelsView,SubscriptionsView,PaymentView}.vue`
- `frontend/src/components/channels/AvailableChannelsTable.vue`
- `frontend/src/components/payment/SubscriptionPlanCard.vue`
- `frontend/src/components/user/dashboard/UserDashboardStats.vue`
- `frontend/src/components/user/{UserErrorRequestsTable,UserErrorDetailModal}.vue`

验证记录：Vue / TypeScript 类型检查通过；完整 Vitest 共 166 个测试文件、1094 项测试通过；品牌别名、下拉搜索和购买页定向回归共 5 个测试文件、25 项测试通过。

### 3.6 环境变量（部署）

`deploy/.env.example` / `deploy/config.example.yaml` 新增示例：

- `SUB2API_KIRO_TIME_CONTEXT` — 空=最稳定 cache 前缀；可选 `date` / `precise`

---

## 4. XorPay 支付（Fork 新增）

### 4.1 功能

- 支付类型：`payment.TypeXorPay` = `"xorpay"`
- 支付宝扫码（当面付），异步 notify + 订单查询
- Webhook：`POST /api/v1/payment/webhook/xorpay`

### 4.2 关键文件

**后端（新增）**

- `backend/internal/payment/provider/xorpay.go`

**后端（修改）**

- `backend/internal/payment/types.go` — `TypeXorPay`
- `backend/internal/payment/provider/factory.go` — 注册工厂
- `backend/internal/service/payment_config_providers.go`
- `backend/internal/service/payment_order_lifecycle.go`
- `backend/internal/handler/payment_webhook_handler.go`
- `backend/internal/server/routes/payment.go`

**前端**

- `frontend/src/components/payment/` — `ProviderCard`, `PaymentMethodSelector`, `PaymentQRDialog`, `PaymentStatusPanel`, `paymentFlow.ts`, `providerConfig.ts`
- `frontend/src/views/user/PaymentQRCodeView.vue`
- `frontend/src/views/admin/orders/AdminOrdersView.vue`
- `frontend/src/types/payment.ts` — `PaymentType` 含 `xorpay`
- `frontend/src/i18n/locales/{zh,en}.ts` — XorPay 文案

### 4.3 Fork 支付 UX 定制（相对官方 + XorPay 本身）

| 改动 | 文件/说明 |
|------|-----------|
| 用户端隐藏倍率展示 | `PaymentView.vue`、`UsageView.vue` CSV 导出等 |
| XorPay 显示名 | 前端统一为 **「XorPay（支付宝）」** |
| 二维码支付金额 | `AmountInput.vue`、`PaymentStatusPanel.vue` — `payAmount` / `creditedAmount` 分离展示 |
| 充值订阅页优化 | 提交 `9a8786ff` 等近期支付流程调整 |

---

## 5. Grok 平台口径（Fork 与官方交叉）

官方 v0.1.142 已含 Grok OAuth / gateway / media routing；本 Fork **额外**保证：

- `constants.go` 中 `PlatformGrok` 与 Kiro 并存
- `account_handler.go` 同时有 Kiro、Grok 模型分支
- `wire_gen.go` 注入顺序：**kiro → grok**
- 迁移 `157` CHECK 约束**同时**保留 kiro + grok（见第 8 节红线）

---

## 6. UI / 品牌 / 交互

| 项 | 说明 | 主要文件 |
|----|------|----------|
| 首页与认证页配色 | 靛蓝紫罗兰主题 + 动效 | `frontend/src/views/HomeView.vue`, `components/layout/AuthLayout.vue` |
| Logo | Fork 自定义 SVG | `frontend/public/logo.svg`（Flymux 风格） |
| Select 组件 | 替换原生 `<select>` | 各 admin/account View；`components/common/Select.vue` |
| ConfirmDialog | 替换原生 confirm/alert | 多处 admin/account 组件 |
| 代理广告横幅 | 已移除 | `ProxyAdBanner` 相关引用清理 |
| 版本徽章 | 隐藏「立即更新」按钮 | `frontend/src/components/common/VersionBadge.vue` |
| Kiro 截图 / 社群 | 文档与资源 | `assets/screenshots/kiro-*.png`, `assets/community/wechat-group.jpg` |

---

## 7. Ops — 管理员删除错误日志（Fork 新增）

官方 v0.1.142 **无**此 API；本 Fork 新增：

| 层 | 文件 |
|----|------|
| 路由 | `backend/internal/server/routes/admin.go` — `DELETE /ops/errors` |
| Handler | `backend/internal/handler/admin/ops_handler.go` — `DeleteErrorLogs` |
| Service | `backend/internal/service/ops_service.go` |
| Repository | `backend/internal/repository/ops_repo.go` |
| 前端 | `frontend/src/api/admin/ops.ts`, `OpsSystemLogTable.vue` 等 |

删除需带时间范围等过滤条件（见 repository 测试 `ops_repo_error_where_test.go`）。

---

## 8. 全局访问封禁 Access Ban（Fork 新增，官方 v0.1.146 无）

> **官方 v0.1.164 不存在** `ip_bans` 表、`IPBanService`、`IpBansView`。merge 上游时若出现同名模块，须逐行对比，**以本 Fork 实现为准**。

### 8.1 已提交能力（相对 v0.1.146）

- 数据表：`ip_bans`（迁移 `159_create_ip_bans.sql`）
- 规则类型（初版）：**IP / CIDR**（`pattern`）
- 中间件：`IPBanGuard`、`GatewayIPBanGuard` — 拦截注册与普通 API / Gateway 流量
- 管理端 CRUD：`GET/POST/PUT/DELETE /api/v1/admin/ip-bans`
- 前端：`IpBansView.vue`、`api/admin/ipBans.ts`
- Wire：`IPBanRepository` → `IPBanService` → `IPBanHandler`；`AuthService` 注入 `accessBanService`
- 客户端 IP：尊重 `X-Forwarded-For`（`ippkg.GetClientIP`）

**关键文件（新增）**

```
backend/ent/schema/ip_ban.go
backend/internal/handler/admin/ip_ban_handler.go
backend/internal/repository/ip_ban_repo.go
backend/internal/server/middleware/ip_ban_guard.go
backend/internal/service/ip_ban_service.go
backend/migrations/159_create_ip_bans.sql
frontend/src/api/admin/ipBans.ts
frontend/src/views/admin/IpBansView.vue
```

### 8.2 Fork 扩展（已提交，merge 时须保留）

在 IP 封禁基础上扩展为多类型 **Access Ban**：

| 规则类型 | `rule_type` | 匹配逻辑 |
|----------|-------------|----------|
| IP / CIDR | `ip` | `ippkg.MatchesPattern(clientIP, pattern)` |
| User-Agent 子串 | `ua` | UA contains pattern（不区分大小写） |
| IP + UA 组合 | `ip_ua` | 同时匹配 `pattern` + `ua_pattern` |
| 邮箱后缀 | `email_suffix` | 注册/绑邮时匹配 `@domain` 或 `*.domain` |

**新增/变更文件**

```
backend/internal/pkg/accessban/match.go          # 规则匹配核心（新增）
backend/internal/pkg/accessban/match_test.go
backend/internal/service/registration_email_format.go   # 注册邮箱格式强校验（新增）
backend/internal/service/registration_email_format_test.go
backend/internal/service/ip_ban_service_test.go
backend/migrations/160_extend_access_ban_rules.sql        # rule_type / ua_pattern / 索引（新增）
```

**联动修改（工作区）**

- `ip_ban_service.go` — 多规则类型 CRUD、`CheckClient` / `CheckEmail`
- `auth_service.go` — 注册/绑邮前 `ValidateRegistrationEmailFormat` + 邮箱后缀封禁
- `ip_ban_guard.go` — UA / 组合规则拦截
- `IpBansView.vue` — 规则类型筛选与创建表单
- `UsersView.vue` + `user_handler.go` — 管理端**批量删除用户**（`POST .../users/batch-delete`，跳过 admin 角色）

### 8.3 合并注意

- 迁移顺序：`159` 建表 → `160` 扩展列；**不可**被上游空迁移覆盖
- 中间件注册位置在 `routes` / `middleware` 链 — merge 后确认 `IPBanGuard` 仍在 Gateway 与 Auth 路由上
- 勿将 `IPBanService` 从 `AuthService` 构造中移除

### 8.4 订阅重置卡（Fork 新增，2026-08-16）

- 迁移 `224_create_user_subscription_reset_cards.sql` 新建永久明细表 `user_subscription_reset_cards`：记录订阅、`validity_days` 快照、来源类型/引用/序号、创建与消费时间；来源唯一键保证支付订单/兑换码幂等，同一订阅删除受 `ON DELETE RESTRICT` 保护。
- 迁移 `225_backfill_user_subscription_reset_cards.sql` 将历史 `manual_reset_credits > 0` 按所属 `group.default_validity_days`（限制在 1–36500 天）逐张回填为 `legacy_backfill` 卡；`ON CONFLICT DO NOTHING` 可重放，且不改旧计数。
- 对仍在有效固定期限内的支付购买、正数订阅兑换或**管理员分配**，不再顺延现有到期时间，而是发放一张以本次 `validity_days` 为快照的卡；支付/兑换用来源引用保证同一订单只发一次，管理员分配每次发一张。已过期订阅仍从当前时刻直接重开对应期限，不发卡。
- 消费接口为 `POST /api/v1/subscriptions/:id/reset-cards/consume`，按指定 `validity_days` 取一张未消费卡；同一事务内先锁订阅行，再新语句按幂等键重放或选卡消费，将日/周/月 USD 与 token 用量全部清零、窗口起点改为点击时刻，并放弃原剩余时间，从点击时重开该卡期限。旧 `POST .../:id/reset-daily` 保留并映射为消费 1 天卡。
- 用户 `/subscriptions` 接口返回按期限聚合的 `reset_cards.total/groups`；页面按期限展示按钮与数量，并标注“永久有效”。`manual_reset_credits` 继续作为兼容镜像：发卡加一、消费减一；新逻辑和 UI 不再以它作为卡期限的事实来源。
- 订阅订单退款：未消费来源卡与订单 `REFUNDED` 同一事务作废，网关成功后才落盘；`REFUNDING` 重入根据 `REFUND_GATEWAY_SUCCEEDED` 审计或查询网关状态重试本地 finalize，不重复打退款网关。已消费来源卡仍需 force，且不缩短当前周期。

**高危同步文件（订阅重置卡）**：

```
backend/internal/service/subscription_service.go
backend/internal/service/user_subscription.go
backend/internal/service/user_subscription_port.go
backend/internal/repository/user_subscription_repo.go
backend/internal/service/payment_fulfillment.go
backend/internal/service/payment_refund.go
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

---

## 9. 迁移红线（必守）

`user_platform_quotas.platform` 的 CHECK 约束会被多次 DROP + 重建：

```
142（上游）→ anthropic, openai, gemini, antigravity
145（Fork）→ + kiro
157（上游原版）→ 仅 + grok → ⚠️ 会丢掉 kiro
157（本 Fork 已修复）→ anthropic, openai, gemini, antigravity, kiro, grok
```

**今后 merge 上游若覆盖 `157`，必须重新确认约束含 `kiro`。**

---

## 10. 合并冲突高危文件（上游常改 + Fork 已改）

改动或 merge 前**必须先读现有实现**：

| 文件 | Fork 改动性质 |
|------|---------------|
| `backend/cmd/server/wire.go` / `wire_gen.go` | Kiro + Grok TokenProvider / UsageService 注入 |
| `backend/internal/domain/constants.go` | `PlatformKiro`, `PlatformGrok` |
| `backend/internal/handler/admin/account_handler.go` | Kiro + Grok 可用模型分支 |
| `backend/internal/service/account_test_service.go` | struct + 构造含 kiro/grok |
| `backend/internal/service/account_usage_service.go` | Provide 版注入 Kiro |
| `backend/internal/payment/types.go` / `provider/factory.go` | `TypeXorPay` |
| `backend/migrations/157_user_platform_quotas_add_grok.sql` | CHECK 含 kiro+grok |
| `backend/internal/i18n/*` + `frontend/src/i18n/locales/{zh,en}.ts` | Kiro / XorPay 文案 |
| `frontend/src/types/index.ts` / `types/payment.ts` | Kiro / XorPay 类型 |
| `frontend/src/views/user/PaymentView.vue` | 倍率隐藏 + 支付金额 UX |
| `frontend/src/components/common/VersionBadge.vue` | 隐藏在线更新按钮 |
| `backend/internal/service/ip_ban_service.go` | Access Ban 多规则类型 |
| `backend/internal/server/middleware/ip_ban_guard.go` | Gateway/Auth 封禁中间件 |
| `backend/internal/service/auth_service.go` | 注册邮箱校验 + 邮箱后缀封禁 |
| `backend/migrations/159_create_ip_bans.sql` / `160_extend_access_ban_rules.sql` | Access Ban 表结构 |
| `frontend/src/views/admin/IpBansView.vue` | 封禁规则管理 UI |
| `backend/internal/server/routes/gateway.go` | Access Ban 中间件；官方新增 `/x_search` 须同步挂 `ipBanAnthropic` |
| `backend/internal/handler/admin/group_handler.go` | 分组 `oneof` 含 `kiro`；合入官方字段时不可丢掉 |
| `backend/internal/service/subscription_service.go` / `user_subscription_port.go` | 固定期限重复购买发卡、过期重开、按期限消费与兼容镜像 |
| `backend/internal/repository/user_subscription_repo.go` | 重置卡幂等发放、行锁消费、用量/窗口原子重开与聚合查询 |
| `backend/internal/service/payment_fulfillment.go` / `redeem_service.go` | 支付订单与正数订阅兑换的来源引用、`validity_days` 快照 |
| `backend/internal/service/payment_refund.go` | 未消费来源卡与 REFUNDED 同事务作废；REFUNDING 重入不重复打网关 |
| `backend/internal/handler/subscription_handler.go` / `server/routes/user.go` | 新消费接口及旧 `reset-daily` → 1 天卡兼容映射 |
| `backend/migrations/224_create_user_subscription_reset_cards.sql` / `225_backfill_user_subscription_reset_cards.sql` | 永久卡明细、历史兼容计数按分组默认期限回填 |
| `frontend/src/views/user/SubscriptionsView.vue` / `api/subscriptions.ts` / `types/index.ts` | `/subscriptions` 按期限展示永久重置卡与消费动作 |

---

## 11. 构建与部署差异

| 文件 | Fork 改动 |
|------|-----------|
| `Dockerfile` | `NPM_CONFIG_REGISTRY` 默认 npmmirror；`pnpm install --ignore-scripts` |
| `.dockerignore` | 保留 `docs/`、`docs/legal/` 供前端合规页构建 |
| `deploy/docker-compose*.yml` | 与 Fork 环境相关的挂载/标签（含 SELinux `:Z` 等历史修复） |
| `deploy/.env.example` | Kiro 相关 env 示例 |

---

## 12. 完整新增文件清单（85，已提交）

<details>
<summary>backend（点击展开）</summary>

```
backend/internal/handler/admin/group_handler_kiro_validation_test.go
backend/internal/handler/admin/kiro_oauth_handler.go
backend/internal/payment/provider/xorpay.go
backend/internal/pkg/anthropictokenizer/NOTICE.md
backend/internal/pkg/anthropictokenizer/claude.json
backend/internal/pkg/anthropictokenizer/tokenizer.go
backend/internal/pkg/anthropictokenizer/tokenizer_test.go
backend/internal/pkg/kiro/fingerprint.go
backend/internal/pkg/kiro/fingerprint_test.go
backend/internal/pkg/kiro/models.go
backend/internal/pkg/kiro/models_test.go
backend/internal/pkg/kiro/oauth.go
backend/internal/pkg/kiro/oauth_invalid_grant_test.go
backend/internal/pkg/kiro/oauth_test.go
backend/internal/pkg/kiro/signature.go
backend/internal/pkg/kiro/translator.go
backend/internal/pkg/kiro/translator_test.go
backend/internal/pkg/kiro/websearch.go
backend/internal/pkg/kiro/websearch_stream.go
backend/internal/pkg/kiro/websearch_stream_test.go
backend/internal/pkg/kiro/websearch_test.go
backend/internal/pkg/kirocooldown/store.go
backend/internal/pkg/kirocooldown/store_test.go
backend/internal/service/account_kiro_credit_unit_price_test.go
backend/internal/service/account_test_service_kiro_apikey_fallback_test.go
backend/internal/service/account_test_service_kiro_test.go
backend/internal/service/account_usage_service_kiro_apikey_test.go
backend/internal/service/kiro_alignment_test.go
backend/internal/service/kiro_alignment_unit_test.go
backend/internal/service/kiro_cache_emulation.go
backend/internal/service/kiro_cache_emulation_test.go
backend/internal/service/kiro_credits_usage_log_test.go
backend/internal/service/kiro_error_classifier.go
backend/internal/service/kiro_error_classifier_test.go
backend/internal/service/kiro_http_helpers.go
backend/internal/service/kiro_http_helpers_test.go
backend/internal/service/kiro_mapping_fallback_test.go
backend/internal/service/kiro_oauth_service.go
backend/internal/service/kiro_oauth_service_test.go
backend/internal/service/kiro_runtime.go
backend/internal/service/kiro_runtime_state.go
backend/internal/service/kiro_runtime_state_integration_test.go
backend/internal/service/kiro_runtime_state_test.go
backend/internal/service/kiro_session_test.go
backend/internal/service/kiro_sticky_session_test.go
backend/internal/service/kiro_token_provider.go
backend/internal/service/kiro_token_provider_test.go
backend/internal/service/kiro_token_refresher.go
backend/internal/service/kiro_usage_fetcher.go
backend/internal/service/kiro_websearch.go
backend/internal/service/kiro_websearch_test.go
backend/migrations/135_add_group_kiro_cache_emulation.sql
backend/migrations/145_allow_kiro_user_platform_quotas.sql
backend/migrations/151_add_group_kiro_auto_sticky.sql
backend/migrations/152_add_group_kiro_sticky_session_ttl.sql
backend/migrations/153_add_group_kiro_endpoint_mode.sql
backend/migrations/153_add_usage_log_kiro_credits.sql
backend/migrations/159_create_ip_bans.sql
backend/ent/schema/ip_ban.go
backend/internal/handler/admin/ip_ban_handler.go
backend/internal/repository/ip_ban_repo.go
backend/internal/server/middleware/ip_ban_guard.go
backend/internal/server/middleware/ip_ban_guard_test.go
backend/internal/service/ip_ban_service.go
```

</details>

<details>
<summary>frontend（点击展开）</summary>

```
frontend/public/logo.svg
frontend/src/api/admin/ipBans.ts
frontend/src/api/admin/kiro.ts
frontend/src/components/account/__tests__/AccountTodayStatsCell.spec.ts
frontend/src/components/account/__tests__/OAuthAuthorizationFlow.spec.ts
frontend/src/components/common/__tests__/GroupBadge.spec.ts
frontend/src/components/common/__tests__/PlatformIcon.spec.ts
frontend/src/components/common/__tests__/PlatformTypeBadge.spec.ts
frontend/src/components/payment/__tests__/AmountInput.spec.ts
frontend/src/composables/useKiroOAuth.ts
frontend/src/utils/__tests__/platformColors.spec.ts
frontend/src/utils/__tests__/subscriptionPlanValidity.spec.ts
frontend/src/utils/subscriptionPlanValidity.ts
frontend/src/views/admin/IpBansView.vue
frontend/src/views/admin/__tests__/BackupView.spec.ts
frontend/src/views/admin/settings/__tests__/EmailTemplateEditor.spec.ts
```

</details>

<details>
<summary>deploy / 根目录（点击展开）</summary>

```
FORK_CUSTOMIZATIONS.md
assets/community/wechat-group.jpg
assets/screenshots/kiro-account-management.png
assets/screenshots/kiro-add-account.png
assets/screenshots/kiro-cache-emulation.png
```

</details>

<details>
<summary>Fork 扩展新增（7）</summary>

```
backend/internal/pkg/accessban/match.go
backend/internal/pkg/accessban/match_test.go
backend/internal/service/ip_ban_service_test.go
backend/internal/service/registration_email_format.go
backend/internal/service/registration_email_format_test.go
backend/migrations/160_extend_access_ban_rules.sql
frontend/src/components/common/__tests__/Select.spec.ts
```

</details>

---

## 13. 完整修改文件清单（215，已提交）

按目录分组；生成命令：`git diff v0.1.146..HEAD --name-only --diff-filter=M`

<details>
<summary>backend（点击展开）</summary>

```
backend/cmd/server/VERSION
backend/cmd/server/main.go
backend/cmd/server/wire.go
backend/cmd/server/wire_gen.go
backend/cmd/server/wire_gen_test.go
backend/ent/group.go
backend/ent/group/group.go
backend/ent/group/where.go
backend/ent/group_create.go
backend/ent/group_update.go
backend/ent/migrate/schema.go
backend/ent/mutation.go
backend/ent/runtime/runtime.go
backend/ent/schema/group.go
backend/ent/schema/user_platform_quota.go
backend/go.mod
backend/go.sum
backend/internal/config/config.go
backend/internal/config/config_test.go
backend/internal/domain/constants.go
backend/internal/domain/constants_test.go
backend/internal/handler/admin/account_codex_import_test.go
backend/internal/handler/admin/account_data.go
backend/internal/handler/admin/account_data_handler_test.go
backend/internal/handler/admin/account_handler.go
backend/internal/handler/admin/account_handler_available_models_test.go
backend/internal/handler/admin/account_handler_list_test.go
backend/internal/handler/admin/account_handler_mixed_channel_test.go
backend/internal/handler/admin/account_handler_passthrough_test.go
backend/internal/handler/admin/batch_update_credentials_test.go
backend/internal/handler/admin/group_handler.go
backend/internal/handler/admin/ops_handler.go
backend/internal/handler/admin/user_platform_quota_admin_test.go
backend/internal/handler/dto/mappers.go
backend/internal/handler/dto/types.go
backend/internal/handler/gateway_handler.go
backend/internal/handler/gateway_handler_chat_completions.go
backend/internal/handler/gateway_handler_responses.go
backend/internal/handler/gateway_handler_warmup_intercept_unit_test.go
backend/internal/handler/gateway_models_test.go
backend/internal/handler/handler.go
backend/internal/handler/payment_webhook_handler.go
backend/internal/handler/wire.go
backend/internal/model/error_passthrough_rule.go
backend/internal/payment/provider/factory.go
backend/internal/payment/types.go
backend/internal/pkg/usagestats/account_stats.go
backend/internal/repository/account_repo.go
backend/internal/repository/account_repo_temp_unsched_test.go
backend/internal/repository/api_key_repo.go
backend/internal/repository/group_repo.go
backend/internal/repository/ops_repo.go
backend/internal/repository/ops_repo_error_where_test.go
backend/internal/repository/simple_mode_default_groups.go
backend/internal/repository/usage_log_repo.go
backend/internal/repository/usage_log_repo_integration_test.go
backend/internal/repository/usage_log_repo_request_type_test.go
backend/internal/server/api_contract_test.go
backend/internal/server/routes/admin.go
backend/internal/server/routes/payment.go
backend/internal/service/account.go
backend/internal/service/account_service.go
backend/internal/service/account_test_service.go
backend/internal/service/account_usage_service.go
backend/internal/service/account_usage_service_test.go
backend/internal/service/account_wildcard_test.go
backend/internal/service/admin_service.go
backend/internal/service/api_key_auth_cache.go
backend/internal/service/api_key_auth_cache_impl.go
backend/internal/service/billing_service.go
backend/internal/service/domain_constants.go
backend/internal/service/gateway_forward_as_chat_completions.go
backend/internal/service/gateway_forward_as_responses.go
backend/internal/service/gateway_record_usage_test.go
backend/internal/service/gateway_request.go
backend/internal/service/gateway_service.go
backend/internal/service/gateway_streaming_test.go
backend/internal/service/gateway_websearch_emulation.go
backend/internal/service/gateway_websearch_emulation_test.go
backend/internal/service/group.go
backend/internal/service/group_test.go
backend/internal/service/openai_privacy_retry_test.go
backend/internal/service/ops_port.go
backend/internal/service/ops_repo_mock_test.go
backend/internal/service/ops_service.go
backend/internal/service/ops_upstream_context.go
backend/internal/service/payment_config_providers.go
backend/internal/service/payment_order_lifecycle.go
backend/internal/service/pricing_service.go
backend/internal/service/ratelimit_service.go
backend/internal/service/ratelimit_service_401_test.go
backend/internal/service/scheduler_snapshot_service.go
backend/internal/service/token_cache_invalidator.go
backend/internal/service/token_refresh_service.go
backend/internal/service/token_refresh_service_test.go
backend/internal/service/token_refresher_test.go
backend/internal/service/usage_log.go
backend/internal/service/usage_log_helpers.go
backend/internal/service/wire.go
backend/migrations/157_user_platform_quotas_add_grok.sql
backend/scripts/resolve-version.sh
```

</details>

<details>
<summary>frontend（点击展开）</summary>

```
frontend/src/api/__tests__/settings.authSourceDefaults.spec.ts
frontend/src/api/admin/accounts.ts
frontend/src/api/admin/index.ts
frontend/src/api/admin/ops.ts
frontend/src/api/admin/settings.ts
frontend/src/api/admin/users.ts
frontend/src/components/account/AccountStatusIndicator.vue
frontend/src/components/account/AccountTodayStatsCell.vue
frontend/src/components/account/AccountUsageCell.vue
frontend/src/components/account/BulkEditAccountModal.vue
frontend/src/components/account/CreateAccountModal.vue
frontend/src/components/account/EditAccountModal.vue
frontend/src/components/account/OAuthAuthorizationFlow.vue
frontend/src/components/account/QuotaDimensionRow.vue
frontend/src/components/account/QuotaNotifyToggle.vue
frontend/src/components/account/__tests__/AccountStatusIndicator.spec.ts
frontend/src/components/account/__tests__/AccountUsageCell.spec.ts
frontend/src/components/account/__tests__/EditAccountModal.spec.ts
frontend/src/components/admin/ErrorPassthroughRulesModal.vue
frontend/src/components/admin/account/AccountTableFilters.vue
frontend/src/components/admin/account/ReAuthAccountModal.vue
frontend/src/components/admin/payment/AdminOrderTable.vue
frontend/src/components/admin/payment/__tests__/orderCurrencyDisplay.spec.ts
frontend/src/components/admin/usage/UsageStatsCards.vue
frontend/src/components/admin/usage/UsageTable.vue
frontend/src/components/admin/usage/__tests__/UsageStatsCards.spec.ts
frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts
frontend/src/components/admin/user/UserPlatformQuotaModal.vue
frontend/src/components/admin/user/__tests__/UserPlatformQuotaModal.spec.ts
frontend/src/components/channels/AvailableChannelsTable.vue
frontend/src/components/charts/EndpointDistributionChart.vue
frontend/src/components/charts/GroupDistributionChart.vue
frontend/src/components/charts/ModelDistributionChart.vue
frontend/src/components/charts/UserBreakdownSubTable.vue
frontend/src/components/charts/__tests__/GroupDistributionChart.spec.ts
frontend/src/components/charts/__tests__/ModelDistributionChart.spec.ts
frontend/src/components/common/GroupBadge.vue
frontend/src/components/common/GroupOptionItem.vue
frontend/src/components/common/PlatformIcon.vue
frontend/src/components/common/PlatformTypeBadge.vue
frontend/src/components/common/Select.vue
frontend/src/components/common/VersionBadge.vue
frontend/src/components/layout/AuthLayout.vue
frontend/src/components/payment/AmountInput.vue
frontend/src/components/payment/OrderTable.vue
frontend/src/components/payment/PaymentMethodSelector.vue
frontend/src/components/payment/PaymentQRDialog.vue
frontend/src/components/payment/PaymentStatusPanel.vue
frontend/src/components/payment/ProviderCard.vue
frontend/src/components/payment/SubscriptionPlanCard.vue
frontend/src/components/payment/__tests__/PaymentStatusPanel.spec.ts
frontend/src/components/payment/paymentFlow.ts
frontend/src/components/payment/providerConfig.ts
frontend/src/components/user/PlatformUsageBreakdown.vue
frontend/src/components/user/UserPlatformQuotaCell.vue
frontend/src/components/user/dashboard/UserDashboardCharts.vue
frontend/src/components/user/dashboard/UserDashboardRecentUsage.vue
frontend/src/components/user/dashboard/UserDashboardStats.vue
frontend/src/composables/__tests__/useModelWhitelist.spec.ts
frontend/src/composables/useModelWhitelist.ts
frontend/src/constants/account.ts
frontend/src/i18n/locales/en.ts
frontend/src/i18n/locales/zh.ts
frontend/src/types/index.ts
frontend/src/types/payment.ts
frontend/src/utils/platformColors.ts
frontend/src/views/HomeView.vue
frontend/src/views/admin/AccountsView.vue
frontend/src/views/admin/BackupView.vue
frontend/src/views/admin/ChannelsView.vue
frontend/src/views/admin/GroupsView.vue
frontend/src/views/admin/ProxiesView.vue
frontend/src/views/admin/RiskControlView.vue
frontend/src/views/admin/SettingsView.vue
frontend/src/views/admin/SubscriptionsView.vue
frontend/src/views/admin/UsageView.vue
frontend/src/views/admin/UsersView.vue
frontend/src/views/admin/__tests__/SettingsView.spec.ts
frontend/src/views/admin/__tests__/UsageView.spec.ts
frontend/src/views/admin/__tests__/UsersView.spec.ts
frontend/src/views/admin/__tests__/groupsModelsListLayout.spec.ts
frontend/src/views/admin/ops/components/OpsDashboardHeader.vue
frontend/src/views/admin/ops/components/OpsSystemLogTable.vue
frontend/src/views/admin/orders/AdminOrdersView.vue
frontend/src/views/admin/settings/EmailTemplateEditor.vue
frontend/src/views/user/AvailableChannelsView.vue
frontend/src/views/user/KeysView.vue
frontend/src/views/user/PaymentQRCodeView.vue
frontend/src/views/user/PaymentView.vue
frontend/src/views/user/SubscriptionsView.vue
frontend/src/views/user/UsageView.vue
frontend/src/views/user/UserOrdersView.vue
frontend/src/views/user/__tests__/PaymentView.spec.ts
frontend/src/views/user/__tests__/UsageView.spec.ts
frontend/src/views/user/__tests__/paymentUx.spec.ts
frontend/src/views/user/paymentUx.ts
```

</details>

<details>
<summary>deploy / 根目录（点击展开）</summary>

```
.dockerignore
.gitignore
Dockerfile
deploy/.env.example
deploy/README.md
deploy/config.example.yaml
deploy/docker-compose.local.yml
deploy/docker-compose.yml
```

</details>

---

## 14. 与官方同步新版本（标准流程）

1. `git fetch upstream tag vX.Y.Z`
2. `git merge vX.Y.Z`（建议 `--no-commit` 先解决冲突）
3. **冲突原则**：Kiro / XorPay / Access Ban / Grok 定制**全部保留**；wire 注入顺序 **kiro 在前、grok 在后**
4. **必查**：`157` 迁移 CHECK、`159`/`160` Access Ban、`224`/`225` 重置卡迁移、`wire_gen.go`、`account_handler.go`、支付/兑换来源传递、订阅服务/仓储/路由、`SubscriptionsView.vue`、`VersionBadge.vue`、`auth_service.go`
5. 验证：
   - 后端：`go build ./... && go vet ./... && go test -tags=unit ./internal/...`
   - 前端：`pnpm install --frozen-lockfile && pnpm vitest run && pnpm typecheck`
6. 更新 `backend/cmd/server/VERSION`、本文档统计与 `AGENTS.md` 摘要
7. UTF-8 乱码复核

---

## 15. AI / 开发者禁区

- **不要**删除或弱化 Kiro / XorPay / **Access Ban** 模块以「与上游对齐」
- **不要**恢复 `VersionBadge` 的「立即更新」按钮（Fork 禁用在线升级）
- **不要**把迁移 `157` 改回仅含 `grok`
- **不要**删除 `159`/`160` 迁移或回退为仅 IP 封禁
- **不要**删除 `224`/`225` 重置卡迁移、恢复有效固定期限重复购买顺延，或把 `manual_reset_credits` 当作卡期限事实来源
- **不要**在 merge 时用上游版本覆盖 `ProvideAccountUsageService` 而丢掉 Kiro 注入
- **不要**从 `AuthService` 移除 `accessBanService` 或注册邮箱格式校验
- **不要**猜测 Kiro credits 或 cache_read 计费口径；见第 3 节
- 修改支付/配额/迁移/SQL 时，同步检查 `dowalet_db_log` 或本项目 SQL 变更约定（若适用）

---

## 16. 相关文档

| 文档 | 说明 |
|------|------|
| [PAYMENT_CN.md](./PAYMENT_CN.md) | 支付配置（官方文档；XorPay 配置见代码与第 4 节） |
| [ADMIN_PAYMENT_INTEGRATION_API.md](./ADMIN_PAYMENT_INTEGRATION_API.md) | 外部支付对接 Admin API |
| [legal/admin-compliance.zh.md](./legal/admin-compliance.zh.md) | 管理端合规文案 |

---

*最后更新：2026-08-25 · 基准：官方 v0.1.181 · 增量清单见 `FORK_CUSTOMIZATIONS.md` · 工作区：`git status --short`*
