## ADDED Requirements

### Requirement: 同步边界必须是官方 v0.1.185 到 v0.2.0 的精确 release 增量
系统 SHALL 完整包含官方 `upstream-v0.1.185..upstream-v0.2.0` 增量，并 MUST 以 peeled commits `2ac784c51a5d0925b324efef2ba6b3446c364781..aa236488351eb71e120fc2b6fb32e36b0374c918` 为可审计输入。系统 MUST NOT 引入 `v0.2.0` tag 之后的官方 main 内容，也 MUST NOT 因共同祖先较旧而重放已手工同步的旧 release。

#### Scenario: 核验同步输入
- **WHEN** 实施者生成官方增量清单
- **THEN** 增量 MUST 统计为 60 个提交、148 个文件、4,926 行新增和 543 行删除
- **THEN** `v0.2.0` tag object MUST 为 `dd07c4d8d484878e617c945cc8bacc304a5a6560`
- **THEN** peeled commit MUST 为 `aa236488351eb71e120fc2b6fb32e36b0374c918`

#### Scenario: 官方 main 在发布后前进
- **WHEN** 官方 main 包含 `v0.2.0` tag 之后的提交
- **THEN** 这些提交 MUST NOT 出现在本次工作区差异中

### Requirement: OpenAI Fast 分组策略必须覆盖请求与计费
系统 SHALL 支持 Group 级 `force_openai_fast` 和 `free_openai_fast`。强制 Fast MUST 在全局 Fast/Flex 策略前将 OpenAI/Composite 请求设置为 priority；免费 Fast MUST 仅改变用户计费档位为 Standard，不能篡改上游实际 service tier 或账户成本记录。

#### Scenario: 分组强制 Fast
- **WHEN** OpenAI 或路由到 OpenAI 的 Composite 分组启用 `force_openai_fast`
- **THEN** HTTP、Messages 兼容和 WebSocket 请求 MUST 使用 `service_tier=priority`
- **THEN** API Key auth cache 与 scheduler snapshot MUST 保留该分组策略

#### Scenario: 免费 Fast 计费
- **WHEN** 分组启用 `free_openai_fast` 且上游实际使用 priority
- **THEN** 用户费用 MUST 按 Standard 价格计算
- **THEN** 账户成本和实际 service tier 观测 MUST 继续反映真实上游档位

### Requirement: reasoning effort 必须支持按模型匹配和明确的超限策略
系统 SHALL 支持 global、exact model 和 model prefix 三种 reasoning effort 映射作用域，并 SHALL 为超过 `max_reasoning_effort` 的请求提供 `downgrade` 或 `deny` 策略。HTTP、Anthropic Messages 兼容和 WebSocket 路径 MUST 使用相同决策。

#### Scenario: 模型级映射优先
- **WHEN** 同时存在全局映射和匹配当前模型的 exact/prefix 映射
- **THEN** 系统 MUST 按官方优先级选择模型级映射
- **THEN** 不匹配当前模型的规则 MUST NOT 改写请求

#### Scenario: 超限拒绝
- **WHEN** 请求的 reasoning effort 超过 ceiling 且策略为 `deny`
- **THEN** OpenAI 兼容接口 MUST 返回官方定义的拒绝状态和错误契约
- **THEN** Anthropic Messages 兼容接口 MUST 返回对应 forbidden error
- **THEN** 被本地策略拒绝的请求 MUST NOT 调用上游或错误处罚账号

#### Scenario: 兼容降级
- **WHEN** 超限策略缺失或为 `downgrade`
- **THEN** 系统 MUST 保持旧行为，将 effort 降到 ceiling

### Requirement: v0.2.0 的网关与模型能力必须完整可用
系统 SHALL 支持 Kimi 原生 OpenAI Responses 转发、Claude Fable 5.1、无 call ID 的 Codex 自动化启动请求、OpenAI API Key 对话缓存身份以及 scheduler snapshot 中的 passthrough 配置。

#### Scenario: Kimi 原生 Responses
- **WHEN** Kimi 账号接收 OpenAI Responses 请求
- **THEN** 系统 MUST 以 Kimi 原生 Responses 协议转发并按官方契约处理响应与错误

#### Scenario: 自动化 bootstrap 无 call ID
- **WHEN** Codex automation bootstrap 的 `function_call_output` 满足官方安全 envelope 且无 call ID
- **THEN** 系统 MUST 将其规范化为可接受的用户消息
- **THEN** 非 automation_update、路径不匹配、非法 ID、非法时间或畸形 envelope MUST NOT 被放宽

#### Scenario: 模型与调度能力
- **WHEN** 请求 Claude Fable 5.1 或调度 OpenAI passthrough 账号
- **THEN** 模型目录/映射 MUST 识别 Fable 5.1
- **THEN** scheduler snapshot MUST 保留 passthrough 配置

### Requirement: v0.2.0 的稳定性和兼容性修复必须保留
系统 SHALL 修复 WebSocket 在 terminal event 前关闭仍被视为成功、模型级冷却把 `model_not_found` 404 错误转成 429、未启用 server-side-fallback beta 时仍透传 Anthropic fallback 字段，以及渠道一小时缓存写价格不能独立配置的问题。

#### Scenario: WebSocket 提前关闭
- **WHEN** 上游 WebSocket 已开始 Responses turn 但在 terminal event 前以 clean close 或 EOF 关闭
- **THEN** relay MUST 将其视为失败而不是 graceful completion

#### Scenario: 模型不可用与模型冷却并存
- **WHEN** 上游明确返回 `model_not_found` 且该模型存在冷却状态
- **THEN** 客户端 MUST 收到 404 语义，不得被改写为 429

#### Scenario: Anthropic fallback beta
- **WHEN** body 含 `fallbacks` 或 `fallback_credit_token` 但最终 beta header 不含允许 token
- **THEN** 系统 MUST 在签名前移除不被允许的字段
- **WHEN** 对应 beta token 存在
- **THEN** 系统 MUST 保留字段

#### Scenario: 一小时缓存写价格
- **WHEN** 管理员为渠道模型、时间区间或账号统计价格配置 `cache_write_1h_price`
- **THEN** API、仓储、定价解析和前端 MUST 使用 `NUMERIC(20,12)` 等价精度保存和展示
- **THEN** NULL MUST 保留旧 `cache_write_price` 回退行为

### Requirement: 所有 Fork 定制必须保持行为等价
系统 MUST 保留 `AGENTS.md`、`docs/FORK_VS_UPSTREAM.md` 和 `FORK_CUSTOMIZATIONS.md` 定义的全部定制。54 个重叠文件 MUST 使用逐 hunk 三方核对，不得用整文件 ours/theirs 覆盖任何一侧。

#### Scenario: Kiro 与官方 Group/OpenAI 新策略共存
- **WHEN** 创建、复制、更新或缓存 Group，或构建 scheduler snapshot
- **THEN** Kiro MUST 继续存在于合法平台、Composite、调度、配额和模型目录
- **THEN** 新 Fast/reasoning effort 字段 MUST 正确保存和投影
- **THEN** Kiro 独立 429、`kiro_credits` 和 cache_read 计费口径 MUST 不变

#### Scenario: 资金和订阅链路回归
- **WHEN** 用户通过 XorPay 支付、退款、兑换、重复获得固定期限订阅或消费重置卡
- **THEN** XorPay、期限快照、来源幂等、过期重开、退款作废、额度窗口清零和兼容镜像 MUST 保持现有行为

#### Scenario: 权限、安全和 UI 回归
- **WHEN** 用户注册、访问网关或管理员操作 UI
- **THEN** Access Ban、注册邮箱格式、提示词审计、批量删用户和 Ops 删除 MUST 保持可用
- **THEN** 用户侧 Kiro→Claude、VersionBadge 无在线更新、GLM 标签、倍率隐藏、金额分离和自定义交互 MUST 保持不变

### Requirement: migration 必须只追加且兼容现有数据库契约
系统 SHALL 新增四个官方 migration 和对应 schema。系统 MUST NOT 修改已发布的 Fork migration，并 MUST 允许同号不同文件名确定性共存。

#### Scenario: 新字段默认值保护现有数据
- **WHEN** migration 应用于已有数据库
- **THEN** `force_openai_fast` 和 `free_openai_fast` MUST 默认为 false
- **THEN** `max_reasoning_effort_over_limit` MUST 默认为 `downgrade`
- **THEN** `cache_write_1h_price` MUST 可空且不改变旧价格回退

#### Scenario: Fork migration 仍有效
- **WHEN** 枚举并应用全部 migration
- **THEN** `157` 平台约束 MUST 同时包含 `kiro` 与 `grok`
- **THEN** `159`/`160` Access Ban 和 `224`/`225` 重置卡 migration MUST 仍存在且内容不回退

### Requirement: 版本、文档和验证证据必须与实现一致
系统 SHALL 将运行版本更新为 `0.2.0`，并更新 Fork 差异文档。完成声明 MUST 基于最终编辑后的新鲜验证结果。

#### Scenario: 读取运行版本和同步记录
- **WHEN** 构建流程读取 `backend/cmd/server/VERSION`
- **THEN** 内容 MUST 精确为 `0.2.0`
- **THEN** Fork 文档 MUST 记录 v0.2.0 tag SHA、能力、四个 migration 和定制保留结论

#### Scenario: 执行最终验证
- **WHEN** 最终代码编辑结束
- **THEN** 后端 MUST 运行相关聚焦测试、`go test ./... -count=1`、`go vet ./...` 和 `go build ./...`
- **THEN** 前端 MUST 运行相关聚焦测试、`test:run`、`lint:check`、`typecheck` 和 `build`
- **THEN** 工作区 MUST 不含冲突标记、`.rej`、`.orig`、BOM、`U+FFFD`、受保护配置或无关生成产物
- **THEN** 任何未运行或失败的检查 MUST 在交付中说明原因和替代证据
