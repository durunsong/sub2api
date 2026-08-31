## 1. 固定基线与执行环境

- [x] 1.1 确认真实仓库根、当前分支、工作区状态和用户已有改动；不得覆盖规划后出现的用户改动。
- [x] 1.2 校验官方 `v0.1.183` / `v0.1.184` tag object 与 peeled commit SHA，重新生成 170 commits / 342 files 的精确增量清单。
- [x] 1.3 恢复项目要求的 Go 1.27.0 与兼容 Corepack/pnpm 的 Node 运行时；记录实际版本，不修改项目依赖来绕过环境问题。
- [x] 1.4 运行当前 Fork 的关键定制测试基线，区分既有失败与同步引入的回归。

## 2. 同步数据库、Ent schema 和共享契约

- [x] 2.1 先移植三个 `231_*` migration 的官方测试/契约检查并观察缺少 184 schema 时的预期失败。
- [x] 2.2 合入三个官方 `231_*` migration、User schema 和 usage log schema/runtime 变更；保留全部 Fork migration 和字段。
- [x] 2.3 合入 DTO、API contract、usage request kind/reasoning effort 的共享类型与 mapper 变更。
- [x] 2.4 运行 migration、Ent schema、DTO 和 API contract 聚焦测试，复核同号 migration 的确定顺序及 Kiro/Grok 平台 CHECK。

## 3. 同步 Codex 路由模型目录、账号和公共分组访问

- [x] 3.1 先移植 Codex routed catalog、capability intersection、exact account alias、user public group restriction 和 partial update 回归测试，观察预期失败。
- [x] 3.2 合入 Codex 模型目录、能力缓存/同步、模型发现和 exact alias 路由实现。
- [x] 3.3 合入管理员限制用户公共分组、用户绑定/列表/监控范围和部分更新保护。
- [x] 3.4 在 account/group handler、service、Wire 和平台目录重叠文件中保留 Kiro、Access Ban、批量删用户及管理端真实平台语义。
- [x] 3.5 运行 Codex、account、group、user、monitor 和 Kiro/Composite 聚焦测试。

## 4. 同步 OpenAI 传输、WebSocket、配额和 service tier

- [x] 4.1 先移植 OpenAI HTTP/WS bridge、WS session isolation、client-close attribution、raw stream truncation、terminal failure、quota reset、Spark window、service tier 和 image cooldown 测试，观察预期失败。
- [x] 4.2 合入 HTTP/SSE/WS 转发、会话隔离、大请求 bridge、keepalive、流式失败与正常断开判定。
- [x] 4.3 合入多渠道 service tier 传递/观测/计费、配额重置原子化、Spark 模型范围和图像工具冷却配置。
- [x] 4.4 在 gateway、scheduler、ratelimit、usage billing 和 Wire 重叠文件中保留 Kiro 独立冷却、Kiro 计费、粘性会话和 Access Ban。
- [x] 4.5 运行 OpenAI service/handler、gateway route、billing、ratelimit、Kiro 和提示词审计入口聚焦测试。

## 5. 同步协议适配器和其他 provider 修复

- [x] 5.1 先移植 Anthropic-to-Responses/Chat 生命周期、Claude effort、Anthropic/Bedrock failover、Grok tool union/image output、Antigravity mixed tools、Ollama/智谱用量测试，观察预期失败。
- [x] 5.2 合入 apicompat、Claude、Anthropic/Bedrock transport、Grok、Antigravity、Ollama Cloud 和智谱团队套餐实现。
- [x] 5.3 合入 Claude attribution header、Anthropic tool arguments 和 cross-provider reasoning replay 修复。
- [x] 5.4 运行 apicompat、provider、account usage、gateway 和 Fork Kiro translator/websearch 聚焦测试。

## 6. 同步用量、计费、支付、订阅和设置修复

- [x] 6.1 先移植 native compaction、requested reasoning effort、DeepSeek peak/off-peak、渠道后缀模型定价、EasyPay 相对 URL、币种展示、SMTP TLS 和订阅窗口测试，观察预期失败。
- [x] 6.2 合入 usage repository/service、pricing、billing probe、payment provider、settings 和 subscription reset time 实现。
- [x] 6.3 在 payment/subscription/user DTO 重叠文件中保留 XorPay、重置卡来源幂等/期限快照/退款作废/兼容镜像和 Fork 支付 UX。
- [x] 6.4 运行 usage、pricing、payment、settings、subscription、reset-card、XorPay 和退款聚焦测试。

## 7. 同步前端 API、类型、组件和页面

- [x] 7.1 先移植 Codex catalog、reasoning effort、本地过期时间解析、账号用量、设置、分组、用量和支付相关 Vitest，观察预期失败。
- [x] 7.2 合入前端 Codex API/配置工具、usage API/types、账号/分组/设置/用量/注册页面和 i18n 变更。
- [x] 7.3 在 184 重叠页面中保留 Kiro→Claude 用户侧品牌、XorPay、GLM 套餐命名、VersionBadge、支付金额 UX、倍率隐藏和自定义 Select/ConfirmDialog。
- [x] 7.4 运行本切片官方测试及 Fork `client.spec.ts`、`PaymentView.spec.ts`、SubscriptionPlanCard、VersionBadge、branding 和平台显示测试。
- [x] 7.5 运行前端 typecheck 和 lint，修复同步导致的类型/模板问题，不格式化无关文件。

## 8. 完成 Fork 红线审计

- [x] 8.1 对 89 个重叠文件逐项对照官方 183、当前 Fork 和官方 184，确认每个官方 hunk 已合入或有明确 Fork 等价实现。
- [x] 8.2 检查 Wire 注入顺序、Kiro/Grok 平台目录、迁移 CHECK、Gateway Access Ban、XorPay factory/routes 和 VersionBadge。
- [x] 8.3 检查提示词审计、订阅重置卡 `224/225`、来源幂等、有效期快照、退款事务、旧接口映射和 `manual_reset_credits` 兼容镜像。
- [x] 8.4 检查 `v0.1.183` 后本地 transient refresh session 和 Claude 套餐标签修复仍有测试覆盖且通过。
- [x] 8.5 检查无冲突标记、`.rej`、`.orig`、意外删除、生成产物或受保护配置改动。

## 9. 更新版本和 Fork 文档

- [x] 9.1 将 `backend/cmd/server/VERSION` 更新为 `0.1.184`，确认版本解析脚本返回一致值。
- [x] 9.2 更新 `FORK_CUSTOMIZATIONS.md`，记录官方 184 功能、修复、三个 migration 和定制保留结论。
- [x] 9.3 更新 `docs/FORK_VS_UPSTREAM.md` 的基线、当前版本、能力摘要、迁移和同步说明。
- [x] 9.4 更新 `AGENTS.md` 的已同步基线、VERSION、官方 184 摘要和后续同步边界。
- [x] 9.5 检查所有新增/修改文本为 UTF-8 无 BOM，且不含异常替换字符。

## 10. 最终验证与交付证据

- [x] 10.1 最终编辑后运行全部相关 Go 聚焦测试并读取退出状态。
- [x] 10.2 使用 Go 1.27.0 运行 `go test ./... -count=1`、`go vet ./...` 和 `go build ./...`。
- [x] 10.3 运行 `pnpm --dir frontend run test:run`、`lint:check`、`typecheck` 和 `build`。
- [x] 10.4 运行 `git diff --check`、冲突/临时文件扫描、版本/迁移/Fork 符号断言和 UTF-8/BOM/乱码检查。
- [x] 10.5 检查最终 Git diff 仅包含本次同步和规划文件，不包含秘密、本地配置、依赖或生成目录。
- [x] 10.6 在实施证据中记录每条命令、版本、退出状态、失败/跳过原因和替代证据；不得把未运行检查报告为通过。
