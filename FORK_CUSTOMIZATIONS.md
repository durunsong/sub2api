# sub2api Fork 自定义功能清单

当前整合版本为 **v0.1.182**，基于官方
[Wei-Shaw/sub2api](https://github.com/Wei-Shaw/sub2api) **v0.1.182**（官方无独立 v0.1.174 tag），并保留本 Fork
的全部定制能力。

## 必须保留的模块

| 模块 | 保留要求 |
|------|----------|
| Kiro 平台 | OAuth、调度、配额、模型映射、缓存模拟和 Web Search 链路完整保留 |
| 用户侧平台展示 | Kiro 在用户端统一显示为 Claude，管理端仍保留真实平台信息 |
| XorPay | 支付创建、查询、回调、退款及前端支付宝展示完整保留 |
| Grok/Kiro 配额 | 账号配额刷新、价格与用量口径不可回退 |
| Access Ban | 全局 IP/访问规则、管理端页面和网关中间件不可回退 |
| 提示词安全审计 | 独立扫描引擎、策略、事件、管理控制台及迁移必须保留 |
| Fork UI | 首页、品牌、VersionBadge、支付体验和管理端定制必须保留 |
| Ops | 管理端筛选删除错误日志能力必须保留 |
| 套餐续期 | 过期套餐直接重开；有效期内再分配/再购买发卡且不改到期时间 |
| 订阅重置卡 | 有效固定期限支付、兑换或管理员分配重复获得时改发按 `validity_days` 快照的永久卡；消费时放弃余期、清零日周月 USD/token 并从点击时重开；订阅退款作废未消费来源卡与 REFUNDED 同事务；旧计数仅作兼容镜像 |

## feature/subscription-reset-cards（2026-08-16）

- 新增迁移 `224`/`225`：建立 `user_subscription_reset_cards` 永久明细；将历史 `manual_reset_credits` 按 `group.default_validity_days` 回填为可重放的期限卡，保留旧计数不变。
- 有效固定期限的支付购买和正数订阅兑换重复购买不再顺延，按本次 `validity_days` 快照幂等发卡；过期购买仍从购买时直接重开，不发卡。
- 消费指定期限卡时放弃原剩余时间，清零日/周/月 USD 与 token、重置三个窗口，并从点击时开始新期限；旧 `reset-daily` 兼容映射为 1 天卡。
- 用户 `/subscriptions` 按期限返回并展示卡数量和“永久有效”；`manual_reset_credits` 保留为发卡加一、消费减一的兼容镜像，不承载期限事实。
- 高危同步范围：`subscription_service.go`、`user_subscription{,_port}.go`、`user_subscription_repo.go`、`payment_fulfillment.go`、`redeem_service.go`、订阅 handler/routes/DTO、迁移 `224`/`225`、`SubscriptionsView.vue`、订阅 API/types/i18n。

## v0.1.182 整合内容

- 合入 OpenAI Responses Lite 在 OAuth、API Key、HTTP 与 WebSocket 下的统一处理：固定并行工具调用模式，并保留数值精度。
- 合入 OAuth 图片生成原样保留用户提示词、OpenCode Go 用量重置时长解析、Anthropic 缓存创建明细去重计费。
- 合入 Antigravity Sonnet 4.5 兼容路由修正（显式 4.5 仍透传）、Composite Kimi Code K3 模型标识路由。
- 合入渠道监控 V2 将 Composite 分组错误归属到真实账号平台，以及余额充值完成后及时刷新用户余额。
- 无新增数据库迁移。官方 tag 内 `backend/cmd/server/VERSION` 仍为 `0.1.181`，本 Fork 按 release tag 设为 `0.1.182`；Kiro、XorPay、Access Ban、Kiro→Claude、提示词审计、订阅重置卡、Ops、VersionBadge、GLM 套餐、支付体验和 UI 品牌全部保留。

## v0.1.181 整合内容

- Gemini 工具 schema 会递归移除不支持的 `deprecated`，将标量 enum 规范化为字符串，并丢弃含非标量值的 enum。
- OpenAI Responses Lite 在工具搬入 `additional_tools` 后继续保留 `parallel_tool_calls=false`；Responses 拒绝字段重试会一次清理同类 input item 的无效 `status`，避免耗尽重试预算。
- Grok OAuth、模型探测和计费探针统一使用官方 CLI User-Agent，CLI 身份版本更新为 `0.2.120`。
- 无新增数据库迁移。官方 tag 内 `backend/cmd/server/VERSION` 仍为 `0.1.180`，本 Fork 按 release tag 设为 `0.1.181`；Kiro、XorPay、Access Ban、Kiro→Claude、提示词审计、订阅重置卡、Ops、VersionBadge、GLM 套餐、支付体验和 UI 品牌全部保留。

## v0.1.180 整合内容

- 合入 OpenAI 重置卡按用量阈值自动使用、Responses/Chat/WS Fast mode `service_tier` 全链路传递与按上游实际档位计费。
- 合入模型列表响应读取上限、模型广场上下文阶梯/渠道分时定价、Ops 错误详情返回列表，以及实验性 OAuth 出站传输插件。
- 新增迁移 `229_plugins.sql`、`230_plugin_artifacts.sql`；Wire 同时保留 Kiro/Kiro OAuth/IP Ban，并接入插件管理器与 OpenAI 自动重置服务。
- Kiro 继续进入共享平台目录、Composite 合法目标、渠道匹配和调度快照；Kiro 直连/中转账号表单语义、XorPay、Access Ban、Kiro→Claude、提示词审计、订阅重置卡、Ops、VersionBadge、GLM 套餐、支付体验和 UI 品牌全部保留。

## v0.1.179 整合内容

- 合入国产供应商自适应 API 协议、Composite Codex/Kimi/智谱/DeepSeek 路由、请求头覆写和 `/v1/responses/input_tokens`。
- 合入渠道 fast/flex 与上下文区间倍率、Anthropic Fast 计费、可配置代理探测目标，以及用量统计单次扫描聚合与索引优化。
- 新增迁移 `226_add_usage_log_effective_model_indexes_notx.sql`、`227_composite_routes_add_cn_providers.sql`、`228_channel_pricing_multipliers.sql`。
- 上游长上下文计费门控由“分组且账号开启”改为“任一开启”，升级后需核对存量分组计费配置。
- Kiro 继续进入共享平台目录、Composite 合法目标、渠道匹配和调度快照；XorPay、Access Ban、Kiro→Claude、提示词审计、订阅重置卡、Ops、VersionBadge、GLM 套餐、支付体验和 UI 品牌全部保留。

## v0.1.178 整合内容

- 合入 Kimi、智谱、DeepSeek 多协议供应商支持，包含分组、调度、计费、count_tokens、配额/余额监控与前端管理能力。
- 合入渠道监控配额快照模式、渠道模型分时倍率、OpenAI Team 联动熔断、OpenAI 账号批量设置、Grok 聚合用量与 Ollama 用量查询。
- 合入 Codex/Gemini/Claude/WS/Ops/i18n 等集中修复；渠道监控配额公开开关默认关闭。
- 官方迁移 `224_user_platform_quotas_add_cn_providers.sql`、`225_backfill_codex_fingerprint_seed.sql`、`225_channel_model_time_pricing.sql`、`226_channel_monitor_quota_mode.sql` 与 Fork 已部署的订阅重置卡 `224_create_user_subscription_reset_cards.sql`、`225_backfill_user_subscription_reset_cards.sql` 按完整文件名并存，禁止重命名既有迁移。
- Kiro、XorPay、Access Ban、用户端 Kiro→Claude、提示词安全审计、订阅重置卡、Ops、VersionBadge、GLM 套餐、支付体验及 UI 品牌等全部定制继续保留。
## v0.1.177 整合内容

- 合入迁移 `222_group_usage_daily_rollups.sql` / `223_group_usage_rollup_timezone.sql`，为分组用量建立日汇总并跟随服务端配置时区；分组页与仪表盘统计提供 `today` / `yesterday` / `total`，避免大数据量下重复扫描明细。
- 合入 Codex remote compaction v2：OAuth 会话请求补齐 `remote_compaction_v2` beta header，原生 v2 保留 `/responses` 路由，账号压缩测试改用 v2 原生探测。
- 合入 `x-codex-turn-state` 的 HTTP/SSE/WS 回传与来源记录，已知状态若由另一账号签发则在上游请求前剥离，防止跨账号回显；Codex 指纹收敛默认改为关闭、仅显式配置后 opt-in，并覆盖透传路径。
- 修复 Grok 长上下文阶梯仅受分组开关控制，以及带版本号的 image/video/audio 媒体模型不再误继承文本 token 价；修复账号页自动刷新偏好在页面加载时被覆盖。
- 官方 v0.1.177 tag 内 `backend/cmd/server/VERSION` 是 0.1.176；本 Fork 按 release tag 设为 0.1.177。
- Kiro、XorPay、Access Ban、用户端 Kiro→Claude、提示词安全审计、套餐续期、Ops 扩展、VersionBadge 禁用在线更新、GLM 套餐分类、支付体验及 UI 品牌等全部定制继续保留。

## v0.1.176 整合内容

- 合入 Grok 4.6（`grok-4.6` / `grok-4.6-latest`）目录、官方定价与 JWT 订阅档位识别（free / SuperGrok / Heavy / Lite）。
- 合入分组逐模型定价：`model_pricing` 与 `long_context_pricing_enabled`，解析链为 Group → Channel → 内置；关闭长上下文时 token 模型只取最低档。
- 合入原生 `POST /x_search`（仅 Grok 分组），Chat↔Responses 保留 x_search 过滤字段；Gateway 兼容路径与官方路径均挂 Access Ban（`ipBanAnthropic`）。
- 合入备份 leader 锁、渠道缓存失效、定价冲突检测、Responses 探测、Realtime 音频计费、未登记 Grok 文本模型计费等修复。
- 新增迁移 `221_group_model_pricing.sql`（`long_context_pricing_enabled` 默认 `TRUE`）。官方 tag 内 VERSION 仍为 0.1.175，本 Fork 按 release tag 设为 0.1.176。
- 迁移 `157` 继续同时包含 `kiro` + `grok`，`159`/`160` Access Ban 迁移继续保留；分组 `oneof` 仍含 `kiro`；Kiro 缓存模拟 UI 与逐模型定价 UI 并存。Kiro、XorPay、Access Ban、用户端 Kiro→Claude、VersionBadge 禁用在线更新、GLM 套餐分类和支付定制全部保留。

## v0.1.175 整合内容

- 合入 Codex OAuth 设备指纹收敛（off/device/session/full），减少上游可见的设备数和会话数。
- 合入按上游响应模型计费，渠道可选择以上游实际返回的模型作为计费基准，并保留官方对准入过宽的安全收紧。
- 合入大文件备份分卷上传与恢复；管理端备份页保留 Fork 的 ConfirmDialog / 密码弹窗，同时支持分卷下载列表。
- 合入简单模式下显示安全审计菜单、Composite 分组图片生成权限开关、API Key 配额/到期输入校验，以及运营监控内存容量显示优化。
- 合入 HTML 403 不再误罚账号、OpenAI 个人订阅到期不被 workspace 覆盖、Responses 空 completed 流 failover、确定性 400 透传、嵌套 usage 解析、TTFT、Codex 容量退避、Grok usage 守卫、User-Agent 指纹校验等修复。
- 官方无 v0.1.174 tag；官方 v0.1.175 tag 内 `backend/cmd/server/VERSION` 仍为 0.1.173，本 Fork 按 release tag 设为 0.1.175。
- 无新数据库迁移。Kiro、XorPay、Access Ban、用户端 Kiro→Claude、VersionBadge 禁用在线更新、GLM 套餐分类和支付定制全部保留。

## v0.1.173 整合内容

- 合入 Grok 邮箱密码 SSO、refresh token 重认证与 Redis 跨实例 OAuth 会话，补齐凭据脱敏、失败恢复和并发刷新链路。
- 合入 Grok 媒体 Voice、视频/搜索计费及账号调度能力，并完善免费额度软门禁、媒体账号资格与上游模型观测。
- 合入渠道监控 V2 的被动聚合、错误分类、趋势/矩阵与安全默认，支持 V1/V2 显式切换、温和回填和隐私化吞吐量展示。
- 合入邮箱域名限量注册、Gemini 图片输出计量与错误策略修复，以及 OpenAI Responses、Messages、WebSocket、图像和搜索计费/路由修复。
- 新增迁移 `194`–`206` 与 `217`–`220`，覆盖渠道监控 V2、分组视频/Voice/搜索定价，并备份后清理非 Grok/非 composite 分组的视频价格。
- 破坏性默认变化：渠道监控保持 V1、V2 改为显式启用；V2 吞吐量默认隐藏；Grok 免费额度软门禁默认启用；非 Grok/非 composite 视频价格会由迁移 `220` 备份后清空。
- Kiro、XorPay、Access Ban、用户端 Kiro→Claude、VersionBadge 禁用在线更新、GLM 套餐分类和支付定制全部保留。

## v0.1.172 整合内容

- 合入安全加固与上游响应模型审计，新增迁移 `194`/`195`，补齐响应模型记录、错配查询索引及相关用量查询能力。
- 合入订阅日额度按午夜重置、金额按数据库 `NUMERIC(20,8)` 精度量化，避免日窗口与高精度金额口径漂移。
- 合入网关、OpenAI/Grok/Antigravity 模型兼容与转发、腾讯验证码、Ops 日志、代理超时及模型广场等集中修复。
- 官方 v0.1.172 tag 内 `backend/cmd/server/VERSION` 仍为 0.1.171；本 Fork 按 release tag 设为 0.1.172。
- 迁移 `157` 继续同时包含 `kiro` + `grok`，`159`/`160` Access Ban 迁移继续保留；Kiro、XorPay、Access Ban、Kiro→Claude、VersionBadge、支付与订阅定制全部保留。

## v0.1.171 整合内容

- 合入腾讯天御、阿里云验证码 2.0，与 Turnstile 三选一，并覆盖注册、登录、找回、OAuth 与 Passkey 门禁。
- 合入 Codex 出站身份统一与版本同步、composite 推理强度、OpenAI reset credit 缓存、Messages 临时错误 failover、WS/调度取消和模型广场图片价格修复。
- 合入退款 `require_force`、Stripe 幂等、订阅行锁、token refresh 单飞与计费失败用量留痕。
- 保留 Kiro/XorPay/Access Ban、Kiro→Claude、续期日卡与备注语义、Fork 品牌/交互及 VersionBadge 禁用在线更新。
- 官方 tag 的 `backend/cmd/server/VERSION` 遗漏仍为 0.1.170；本 Fork 明确修正为 0.1.171。

## v0.1.170 整合内容

- 合入分组级利润控制、请求定价时刻锁定、槽位后二次复核、认证缓存失效迁移 `192`/`193` 与 `profit-preview` 离线预演工具；Kiro 仍保留原有粘性 TTL 与计费口径，利润控制平台范围遵循上游五个平台。
- 合入 API Key 全平台上游倍率探测、可选倍率自动同步、托管倍率编辑保护和列表同步来源提示。
- 合入 Anthropic 流中断部分用量计费、OpenAI WebSocket/Responses/Codex 修复、Grok 计费事件过滤、订阅窗口对齐和 SMTP 一致性修复。
- 合入内容审核代理、账号筛选结果全选、并发批量删除与首页精简模式；继续使用 Fork 的确认弹窗、Kiro→Claude 别名和首页品牌定制。

## v0.1.169 整合内容

- 合入上游 URL 路径片段闭集校验（`upstream_path_guard` / `guardResponsesSubpath`），修复 GHSA-vrxq-qm4h-6hgg，同时保留 Gateway 上的 Access Ban（`ipBanAnthropic` / `ipBanGoogle`）中间件。
- 合入 OpenAI 代理断流熔断 fail-open、`gateway.openai_proxy_stream_circuit.disabled` 配置项，以及容器 `no-new-privileges`。
- 合入 glm-5.2 独立兜底定价、GPT-5.6 Luna/Terra 费率更新、Anthropic count_tokens 剥离 `max_tokens`、Claude auto 分类器识别修复。
- 合入 SMTP 标准邮件格式、Qwen3Guard 辅助字段兼容、临时不可调度账号跳过 Token 刷新、可用渠道组合模型按平台展示。
- 合入订阅到期标签与套餐卡片标题可读性修复；用户端继续保留 Kiro→Claude 别名，并继续隐藏倍率展示。
- 合入 release 定价兜底资源打包修复与 Passkey 部署说明补充。

## v0.1.168 整合内容

- 合入 Passkey 登录、注册/撤销密码确认、后台开关与 WebAuthn 配置校验，同时保留 Access Ban 的认证拦截与 IP Ban 依赖注入。
- 合入模型广场、分组级模型定价展示和可选 JWT 鉴权，并保留 Kiro 在用户端统一显示为 Claude 的平台别名规则。
- 合入 Kimi K3/1M 后缀支持、Claude Sonnet 5 状态别名、模型 ID 快速复制和 GPT-5.6 `max` 推理强度兼容。
- 合入 Codex API Key Web Search、OAuth system cache breakpoint、透传优先于模型映射及 Anthropic 消息 ID 格式修复。
- 合入 OpenAI Live 会话 store 故障容错、显式 setup bypass，以及用户/API Key 限定列更新与并发丢失更新保护。
- 合入提示词安全审计配置解密恢复与保存死锁修复，同时保留 Fork 的独立扫描引擎、策略、事件和管理控制台。

## v0.1.166 整合内容

- 合入面板 API 分级限流，认证接口按用户、公开接口按安全客户端 IP 计数，同时保留 Access Ban 的认证与网关拦截链路。
- 合入 Antigravity OpenAI 兼容加固、Codex Responses/Anthropic 工具调用配对及账号切换 reasoning 清理。
- 合入 OpenAI WebSocket 分轮模型计费、最终上游模型统计、模型映射计费与 Gemini 3.6 Flash 定价修复。
- 合入系统设置部分更新保护、显式 `CONFIG_FILE`、Grok 402 暂停、Gemini 重试和计费探针时间解析修复。
- 合入多币种支付看板、请求 ID/路由用户筛选、可选推广码、移动端渠道列表和下拉框边界修复，同时保留 XorPay 与 Fork UI。
- 合入 Caddy SSE 去缓冲、图像与遥测依赖安全升级，并保留 Fork 已使用的更高版本 `golang.org/x/image`。

## v0.1.165 整合内容

- 合入 OpenAI Live WebSocket 网关、分组级 Live 开关、平台 attestation 与对应请求计费类型。
- 合入跨请求 `session_id` 持久化，同时保留 Fork 的 Kiro Credits 用量字段与写入/查询链路。
- 合入注册邮箱别名规范化和并发去重，兼容 Fork 原有邮箱格式限制与用户管理逻辑。
- 合入 Ollama Cloud 按模型刷新去抖、PostgreSQL 16 兼容和抓取下限修复。
- 合入 Responses item ID/namespace、Gemini 图片响应、Grok/OpenAI 重试冷却及公告预览样式修复。
- 合入 Claude Opus 5 定价与模型白名单，同时保留 Fork 的 Thinking、Sonnet/Haiku 和 Kiro 映射模型。

## v0.1.164 整合内容

- 合入聚合分组、模型路由规则、路由预览和按实际模型计费，并允许显式路由到 Fork 的 Kiro 平台。
- 合入 Ollama Cloud 官方用量自动刷新与后台配置。
- 合入支付宝移动端预下单深链，同时保留 XorPay 和付款金额/到账金额分离展示。
- 合入 OpenAI 账号测试/透传/代理隔离、Grok 402 冷却、渠道定价归一化、CC Switch 导入和审计脱敏修复。

## 历史整合内容

- 合入官方 v0.1.161 的安全开关、入口拒绝日志降噪和鉴权边界强化。
- 合入鉴权缓存失效 outbox、健康检查及相关数据库迁移。
- 合入 Grok 受保护视频内容代理、媒体模型映射和免费账号探测修复。
- 合入 OpenAI WebSocket turn 生命周期、流式错误处理和 Responses 兼容修复。
- 合入过期套餐重新分配修复，并与 Fork 的套餐叠加和手动重置逻辑合并。
- 合入 Docker 跨架构构建、Redis 启动参数和上游计费倍率展示修复。
- 合入官方 v0.1.162 的客户端 IP 配置、异步生图对象存储后台配置、Grok 工具缓存与 Codex/计费修复。
- 合入官方 v0.1.163 的分组级 OpenAI 推理策略、Grok `/responses/compact`、Redis ACL 用户名与优雅关停用量清理。
- 合入官方 v0.1.163 的调度、SSE、故障转移计费、移动端布局、套餐周期与币种展示修复，同时保留用户端倍率隐藏和 Kiro→Claude 别名。

## 后续同步红线

- 三方合并必须以上一次官方版本为基线，不能用官方整树覆盖 Fork。
- 账号处理器的依赖注入参数必须保持 `kiro` 在前、`grok` 在后；其他服务按各自函数签名传参，两者都必须存在。
- `157_user_platform_quotas_add_grok.sql` 相关结构必须同时兼容 Kiro 与 Grok。
- Access Ban 的服务、路由和网关中间件不得因官方安全中间件更新而被移除。
- `wire.go`、`wire_gen.go`、网关路由、套餐服务、Ops 服务和设置页属于高冲突文件，合并后必须运行对应测试。
- 同步订阅链路时必须保留迁移 `224`/`225`、购买来源幂等键、有效期快照、过期重开、旧 `reset-daily` → 1 天卡映射，以及 `manual_reset_credits` 兼容镜像。

*最后更新：2026-08-25*
