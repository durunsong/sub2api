## 1. 固定基线、环境和重叠清单

- [x] 1.1 重新确认仓库根、`main`/`HEAD`、工作区改动和本地 tag SHA；保留实施期间出现的用户改动。
- [x] 1.2 用 `git rev-list`、`git diff --shortstat` 和 `git diff --name-only` 复核 60 commits / 148 files / 4,926 insertions / 543 deletions，并生成 54 个重叠文件的执行清单。
- [x] 1.3 确认现有 Go 1.27.0、Node、pnpm 运行入口；不得修改依赖或工具链声明迎合错误环境。
- [x] 1.4 运行当前 Fork 的 Kiro、XorPay、Access Ban、提示词审计、订阅重置卡和 UI 品牌关键测试基线，记录既有失败。

## 2. 同步 migration、Ent 和共享契约

- [x] 2.1 先移植 `group_force_openai_fast_migration_test.go`、`group_free_openai_fast_migration_test.go` 及 schema integration 断言，确认缺少四个 migration/字段时预期失败。
- [x] 2.2 新增四个官方 `232/233` migration，并移植 Group schema、Channel pricing schema、Ent 生成文件、DTO 和 mapper 差异。
- [x] 2.3 在 Group/DTO/仓储重叠文件中保留 Kiro/Composite 合法平台、Fork 字段、Access Ban 和订阅相关契约。
- [x] 2.4 运行 migration、Ent、DTO、API contract 聚焦测试，复核默认值、精度、同号排序、`157`、`159/160` 和 `224/225`。

## 3. 同步按模型 reasoning effort 策略

- [x] 3.1 先移植后端 `openai_reasoning_effort_policy_test.go`、handler/WS 相关测试和前端 `ReasoningEffortPolicyFields.spec.ts`、`groupsReasoningEffort.spec.ts`，观察 global/exact/prefix 与 deny 行为缺失导致的失败。
- [x] 3.2 合入 domain、Group CRUD/cache、HTTP、Messages 和 WS 的模型作用域映射及 `max_reasoning_effort_over_limit` 决策。
- [x] 3.3 合入前端分组类型、表单转换、验证、组件和中英文 i18n；保留当前 GroupsView 的 Fork 交互和 Kiro 选项。
- [x] 3.4 运行 reasoning effort 后端/前端聚焦测试，并复核拒绝不会调用上游或处罚账号。

## 4. 同步 OpenAI Fast 与计费策略

- [x] 4.1 先移植 Group Fast handler/DTO/cache/fixture、HTTP/WS gateway、billing 和前端 `groupsOpenAIFast.spec.ts` 测试，观察预期失败。
- [x] 4.2 合入 `force_openai_fast` 的 Group 生命周期、API Key auth projection、scheduler snapshot、HTTP/Messages/WS 请求策略。
- [x] 4.3 合入 `free_openai_fast` 的用户 Standard 计费与真实账户成本/service tier 观测分离。
- [x] 4.4 合入管理端 Group Fast 开关和类型/i18n，并保留 Kiro/Composite、利润控制和 Fork 分组页面行为。
- [x] 4.5 运行 Group、auth cache、scheduler、gateway、billing、account stats 和前端聚焦测试。

## 5. 同步网关、模型和兼容性修复

- [x] 5.1 先移植 Kimi native Responses、automation bootstrap、passthrough snapshot/API Key cache identity、WS terminal event、model_not_found、Anthropic fallback beta、Fable 5.1 测试，观察预期失败。
- [x] 5.2 合入 Kimi 自适应协议和 Responses 转发，保留现有 Kiro/Kimi/智谱/DeepSeek Composite 路由与计费。
- [x] 5.3 合入受严格 envelope 校验的 Codex automation bootstrap、OpenAI cache identity 和 scheduler passthrough projection。
- [x] 5.4 合入 WS terminal event 判定、模型冷却 404、Anthropic fallback 字段清理和 Claude Fable 5.1 常量/映射。
- [x] 5.5 运行 handler、gateway、WS、scheduler、Claude/Antigravity、Kiro translator/websearch 和提示词审计入口聚焦测试。

## 6. 同步渠道定价与前端页面

- [x] 6.1 先移植渠道 `cache_write_1h_price` 仓储/服务测试和前端 account/channel/model plaza 测试，观察预期失败。
- [x] 6.2 合入 channel pricing repository/service/API/types 和 `NUMERIC(20,12)` 精度、NULL 回退行为。
- [x] 6.3 合入 AccountStatus/Usage、账号表单、Channel 定价组件、SupportedModel、UseKeyModal、模型广场和相关 i18n 差异。
- [x] 6.4 在重叠页面中保留 Kiro→Claude、XorPay、GLM 命名、VersionBadge、倍率隐藏、金额分离、Select/ConfirmDialog 和当前品牌。
- [x] 6.5 运行本切片官方 Vitest、Fork 品牌/支付/订阅测试、`lint:check` 和 `typecheck`。

## 7. 完成 54 个重叠文件和 Fork 红线审计

- [x] 7.1 对每个重叠文件核对官方 185、当前 Fork、官方 2.0，确认每个官方 hunk 已移植或有明确等价实现。
- [x] 7.2 复核 Kiro wire/provider 顺序、平台目录、Composite、调度、配额、独立 429、`kiro_credits` 和 cache_read 计费。
- [x] 7.3 复核 XorPay factory/routes/refund/UI、Access Ban 路由/auth/管理页、提示词审计和 Ops 删除。
- [x] 7.4 复核订阅重置卡 migration、期限快照、来源幂等、过期重开、退款作废、消费清零和兼容镜像，以及 `active_available` 筛选。
- [x] 7.5 检查无冲突标记、`.rej`、`.orig`、意外删除、受保护配置、依赖或生成产物改动。

## 8. 更新版本和 Fork 文档

- [x] 8.1 将 `backend/cmd/server/VERSION` 更新为 `0.2.0`，确认版本解析脚本输出一致。
- [x] 8.2 更新 `FORK_CUSTOMIZATIONS.md`，记录官方 2.0 能力、四个 migration、tag SHA 和定制保留结论。
- [x] 8.3 更新 `docs/FORK_VS_UPSTREAM.md` 的基线、当前版本、统计、迁移和同步说明。
- [x] 8.4 更新 `AGENTS.md` 的已同步基线、VERSION、官方 2.0 摘要和后续同步边界。
- [x] 8.5 检查新增/修改文本为 UTF-8 无 BOM，且不含异常替换字符。

## 9. 最终验证与交付证据

- [x] 9.1 最终编辑后重新运行全部受影响 Go 聚焦测试和前端聚焦测试，读取退出状态。
- [x] 9.2 运行 `go test ./... -count=1`、`go vet ./...` 和 `go build ./...`。
- [x] 9.3 运行 `pnpm --dir frontend run test:run`、`lint:check`、`typecheck` 和 `build`。
- [x] 9.4 运行 `git diff --check`、冲突/临时文件扫描、版本/migration/Fork 符号断言和 UTF-8/BOM/乱码检查。
- [x] 9.5 检查最终 Git diff 仅包含 v0.2.0 同步、测试和规划/文档文件，不含 `.env*`、秘密、本地配置、依赖或生成目录。
- [x] 9.6 在交付说明中记录实际命令、版本、退出状态、失败/跳过原因、替代证据和剩余风险。
