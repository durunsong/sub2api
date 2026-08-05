# sub2api Fork 自定义功能清单

当前整合版本为 **v0.1.171**，基于官方
[Wei-Shaw/sub2api](https://github.com/Wei-Shaw/sub2api) **v0.1.171**，并保留本 Fork
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
| 套餐续期 | 活跃套餐叠加、过期套餐续期、日卡重置额度和备注追加必须保留 |

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

*最后更新：2026-08-05*
