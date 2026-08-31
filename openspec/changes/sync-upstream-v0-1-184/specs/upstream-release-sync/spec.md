## ADDED Requirements

### Requirement: 同步必须以官方 v0.1.184 release tag 的精确增量为边界
系统 SHALL 完整包含官方 `v0.1.183..v0.1.184` release 增量，并 MUST 以 peeled commits `e8cb019fabf8b55199436229044cbf9aa7a82564..e98ef32eb29aecd30d1def615912ec4dc93173f3` 为可审计输入。同步 MUST NOT 引入官方 `v0.1.184` tag 之后的 `main` 内容，也 MUST NOT 因本地共同祖先较旧而重放已手工同步的官方 `v0.1.183` 整版。

#### Scenario: 确认同步边界
- **WHEN** 实施者生成上游增量清单
- **THEN** 清单 MUST 来自 `upstream-v0.1.183..upstream-v0.1.184`
- **THEN** 目标 tag peeled commit MUST 为 `e98ef32eb29aecd30d1def615912ec4dc93173f3`
- **THEN** 增量 MUST 统计为 170 个提交和 342 个文件

#### Scenario: 官方 main 在 release 后继续前进
- **WHEN** 官方 `main` 包含 `v0.1.184` tag 之后的提交
- **THEN** 这些提交 MUST NOT 进入本次工作区差异

### Requirement: 官方 v0.1.184 的新增能力必须完整可用
系统 SHALL 包含官方 release 描述和 tag 源码中的全部新增能力，包括 Codex 实际路由模型目录与能力同步、精确账号模型别名、原生 compaction 用量统计与映射前推理强度、用户公共分组访问限制、智谱团队 GLM Coding Plan 用量查询、国产三家平台下的 Ollama Cloud 用量窗口、后台可配置的 OpenAI 图像工具冷却和管理员可配置的 TTFT 指标模式。

#### Scenario: Codex 路由模型目录
- **WHEN** 管理端或兼容 API 查询 Codex 可用模型
- **THEN** 系统 MUST 按实际路由和可调度账号生成模型目录及能力
- **THEN** 精确账号模型别名 MUST 路由到对应账号模型
- **THEN** Kiro MUST 继续存在于 Fork 已支持的共享平台、Composite 和调度目录中

#### Scenario: 用量记录新字段
- **WHEN** 请求是原生 compaction 或包含映射前 reasoning effort
- **THEN** 后端 MUST 按官方 184 契约保存并返回相应字段
- **THEN** 管理端和用户端 MUST 按官方权限语义展示这些字段

#### Scenario: 管理员限制用户公共分组
- **WHEN** 管理员为用户配置可访问公共分组
- **THEN** 用户的分组绑定、列表和监控查询 MUST 遵守该限制
- **THEN** 管理员的部分更新 MUST NOT 意外清空未提交的配额或其他用户属性

#### Scenario: 新的账号用量与运行配置
- **WHEN** 管理员查询智谱团队套餐或将 Ollama Cloud 用量窗口挂载到国产平台账号
- **THEN** 系统 MUST 使用官方 184 的查询和展示契约
- **WHEN** 管理员配置图像工具冷却或 TTFT 模式
- **THEN** 网关运行时 MUST 使用保存后的配置，且读取/测试接口 MUST NOT 覆盖其他已保存设置

### Requirement: 官方 v0.1.184 的兼容性和稳定性修复必须保留
系统 SHALL 合入官方 release 中 OpenAI 多渠道 service tier、WebSocket 隔离与大请求桥接、配额重置、图像能力冷却、流式失败判断、Anthropic/Bedrock 故障转移、Anthropic-to-Responses 流生命周期、Grok 工具结果、支付回调、SMTP TLS 测试、账号过期时间、DeepSeek 价格和 Claude 归因/工具参数等修复。

#### Scenario: OpenAI 请求跨渠道或跨会话
- **WHEN** 请求经历 Responses/Chat fallback、HTTP/WS bridge、流式故障转移或正常客户端断开
- **THEN** service tier、模型范围、会话隔离、故障归因和计费 MUST 符合官方 184 测试契约
- **THEN** 正常客户端关闭 MUST NOT 被记录为账号故障

#### Scenario: Anthropic 或 Bedrock 传输失败
- **WHEN** 上游发生官方定义的可故障转移传输错误
- **THEN** 系统 MUST 触发故障转移并按官方规则临时摘除持久故障账号
- **THEN** Fork 的 Kiro 独立 429 冷却和计费语义 MUST NOT 被通用错误处理覆盖

#### Scenario: 支付和设置修复
- **WHEN** EasyPay 返回相对 pay URL/QR URL、用户选择充值币种或管理员测试 SMTP TLS
- **THEN** 地址解析、币种展示和设置保存 MUST 符合官方 184 契约
- **THEN** XorPay、支付金额/入账金额分离和用户侧倍率隐藏 MUST 保持不变

### Requirement: 所有 Fork 定制必须在同步后保持行为等价
系统 MUST 保留 `AGENTS.md`、`docs/FORK_VS_UPSTREAM.md` 和 `FORK_CUSTOMIZATIONS.md` 记录的全部定制。任何上游冲突 MUST 采用逐 hunk 合并，MUST NOT 使用整文件 ours/theirs 覆盖来跳过一侧行为。

#### Scenario: Kiro 和 Grok 共存
- **WHEN** 服务启动、生成 Wire、创建/更新分组或刷新平台配额
- **THEN** Kiro provider MUST 在 Grok provider 之前注入
- **THEN** 平台约束、Composite、模型目录、调度和计费 MUST 同时包含 Kiro 与 Grok
- **THEN** 用户侧 Kiro MUST 继续显示为 Claude，管理端 MUST 保留真实平台信息

#### Scenario: 访问控制和网关路由更新
- **WHEN** 官方 184 修改 auth、admin 或 gateway 路由
- **THEN** Access Ban 的 IP/UA/IP+UA/邮箱后缀规则和中间件 MUST 继续覆盖既有入口
- **THEN** 管理端批量删用户和 Ops 错误日志删除能力 MUST 保持可用

#### Scenario: 订阅和支付链路更新
- **WHEN** 官方 184 修改订阅窗口、支付或用户 DTO
- **THEN** 固定期限重复购买发卡、过期重开、退款作废来源卡、旧 `reset-daily` 到 1 天卡映射和 `manual_reset_credits` 兼容镜像 MUST 全部保留
- **THEN** XorPay 创建、查询、回调和退款链路 MUST 保持可用

#### Scenario: Fork UI 更新
- **WHEN** 官方前端组件和页面被同步
- **THEN** VersionBadge MUST 继续隐藏立即更新按钮
- **THEN** Fork 品牌、Select/ConfirmDialog、GLM 套餐命名和支付体验 MUST 保持不变

### Requirement: 数据库迁移必须可追加且不回退现有数据契约
系统 SHALL 新增官方三个 `231_*` migration 及对应 schema 变更。迁移 MUST 仅追加，MUST NOT 改写已发布迁移，且 MUST 与 Fork 同号 migration 共存。

#### Scenario: 枚举 migration 文件
- **WHEN** migration runner 按当前项目规则扫描文件
- **THEN** 三个官方 `231_*` 文件 MUST 均被发现且按确定顺序执行
- **THEN** Fork 的 `224_create_user_subscription_reset_cards.sql`、`225_backfill_user_subscription_reset_cards.sql`、`159` 和 `160` Access Ban migration MUST 仍存在

#### Scenario: 校验平台约束
- **WHEN** 全部 migration 应用完成
- **THEN** `user_platform_quotas.platform` 的有效集合 MUST 同时包含 `anthropic`、`openai`、`gemini`、`antigravity`、`kiro` 和 `grok`

### Requirement: 版本、文档和验证证据必须与实际实现一致
系统 SHALL 将运行版本更新为 `0.1.184`，并更新 Fork 差异文档。完成声明 MUST 基于最终编辑后的新鲜验证结果。

#### Scenario: 读取运行版本和差异文档
- **WHEN** 构建流程读取 `backend/cmd/server/VERSION`
- **THEN** 内容 MUST 精确为 `0.1.184`
- **THEN** Fork 文档 MUST 记录官方 184 内容、新增 migration 和定制保留结论

#### Scenario: 执行最终验证
- **WHEN** 最终代码编辑完成
- **THEN** 后端相关测试、`go test ./...`、`go vet ./...` 和 `go build ./...` MUST 使用 Go 1.27.0 运行并成功
- **THEN** 前端聚焦测试、`pnpm run test:run`、`pnpm run lint:check`、`pnpm run typecheck` 和 `pnpm run build` MUST 在兼容 Node/pnpm 环境运行并成功
- **THEN** 工作区 MUST 不包含冲突标记、`.rej` 文件、异常替换字符或 BOM
