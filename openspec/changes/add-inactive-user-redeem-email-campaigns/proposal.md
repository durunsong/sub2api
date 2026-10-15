## Why

运营需要召回长期不使用站点的用户。目前 `/admin/redeem` 只能生成兑换码，管理员要手动导出、配对邮箱再逐封发送，容易出现一码多发、漏发或错发，且无法追踪发送结果。

## What Changes

- `/admin/redeem` 新增「群发余额券」向导和「发送记录」入口。
- 管理员按最后一次 API 计费用量时间的互斥分段筛选用户：「1–2 个月未用」「2–3 个月未用」「3 个月以上未用」「从未使用」，多选 Tab 取并集，默认选前三段（从未使用的用户通常意向较低，默认不选）。只包含 active、未删除、普通角色且邮箱可投递的用户。
- 邮件模板注册为可退订通知事件 `redeem.inactive_coupon`，与其他通知邮件一样在 `/admin/settings` 邮件模板中按 zh/en 编辑；发送时按收件人语言渲染，已退订的收件人被跳过并标记。
- 管理员设置单券金额（USD，正数，最多 8 位小数）与兑换码有效天数（1–3650 天或永不过期），确认后服务端在一个事务内创建发送批次，并为每个收件人生成唯一的 `balance` 兑换码。
- 后台 Worker 以全局约 1 封/秒的速度投递邮件，记录每封邮件的状态和失败原因；失败项可由管理员手动重试。
- 非目标：直接给用户加余额、向非注册邮箱群发、注册自动赠码、OpenAPI 发券。

## Capabilities

### New Capabilities
- `inactive-user-redeem-email-campaigns`：未活跃用户筛选、事务化一人一码发券、异步限速邮件投递与管理端追踪。

### Modified Capabilities
<!-- 无 -->

## Impact

- **数据库**：新增迁移 `246_redeem_email_campaigns.sql`（`redeem_email_batches`、`redeem_email_items`），不修改历史 SQL；表只通过原生 SQL 访问，不新增 Ent schema，避免改动上游高冲突的 `user.go` / `redeem_code.go` schema。
- **后端**：新增 `service/redeem_email_*.go`、`repository/redeem_email_repo.go`、`handler/admin/redeem_email_handler.go`；在 `routes/admin.go`、handler/service/repository ProviderSet、`cmd/server/wire.go` 清理流程中接线。
- **管理端 API**：`/api/v1/admin/redeem-codes/email-batches/*`。
- **通知邮件**：新增 `service/redeem_email_notification.go` 以 `init` 追加事件与官方模板；上游 `notification_email_service.go` 只增加一行示例变量钩子。
- **前端**：`RedeemView.vue` 入口、`components/admin/redeem/` 下三个对话框组件、`api/admin/redeem.ts`、类型与中英文 i18n；`EmailTemplateEditor.vue` 增加事件展示文案、营销分类与 `initialEvent` 预选；`SettingsView.vue` 支持 `?tab=email&email_event=` 深链。
- **资金影响**：每个批次的潜在发放额为「收件人数 × 单券金额」；兑换仍走现有 `RedeemService` 事务核销。
