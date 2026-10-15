## Purpose

规定管理员向长期未使用 API 的用户群发余额兑换券的筛选口径、发券事务、邮件投递与重试行为。

## ADDED Requirements

### Requirement: 未活跃分段按 API 计费用量判断
系统 MUST 以用户最后一条 `usage_logs.created_at`（`last_used_at`）划分互斥分段，边界 `tN` 由服务端当前 UTC 时间减去 N 个自然月（`time.AddDate(0, -N, 0)`）得到：`[t2, t1)` 为 `inactive_1_2m`，`[t3, t2)` 为 `inactive_2_3m`，`< t3` 为 `inactive_3m_plus`，无用量为 `never_used`；`>= t1` 视为活跃，MUST NOT 入选。请求 MUST 指定非空分段子集，结果为所选分段的并集。面板访问或登录 MUST NOT 计入活跃。向导 MUST 默认选中前三段且不选 `never_used`。

#### Scenario: 分段边界
- **WHEN** 当前时间为 2026-10-15T00:00:00Z，用户最后一条用量为 2026-09-15T00:00:00Z
- **THEN** 该用户为活跃，任何分段下 MUST NOT 出现
- **WHEN** 最后一条用量为 2026-09-14T23:59:59Z
- **THEN** 该用户属于 `inactive_1_2m`
- **WHEN** 最后一条用量为 2026-07-14T23:59:59Z
- **THEN** 该用户属于 `inactive_3m_plus`，只选 `inactive_1_2m` 时 MUST NOT 出现

#### Scenario: 从未使用作为可选分段
- **WHEN** 未选择 `never_used`
- **THEN** 没有任何用量记录的用户 MUST NOT 出现在结果中
- **WHEN** 选择 `never_used`
- **THEN** 这些用户 MUST 出现在结果中，`last_used_at` 为空、`segment=never_used`

### Requirement: 收件人范围
结果 MUST 只包含 `status=active`、`deleted_at IS NULL`、`role=user`、邮箱非空且不是 `.invalid` 合成域名的用户。单次搜索或单个批次最多 1000 人，超过时 MUST 返回 `REDEEM_EMAIL_RECIPIENT_LIMIT_EXCEEDED`。

#### Scenario: 管理员和禁用用户被排除
- **WHEN** 管理员或禁用用户满足未活跃条件
- **THEN** 他们 MUST NOT 出现在结果中，也 MUST NOT 被创建为批次收件人

### Requirement: 事务化一人一码
创建批次时系统 MUST 在服务端重新校验所有选中用户的资格，并在同一数据库事务中创建批次、每人一个 `type=balance`、`status=unused` 的新兑换码以及对应投递明细。任一用户不合格、金额/有效期非法或任一步骤失败时 MUST 整体回滚，不留下兑换码或批次。

#### Scenario: 选中用户期间变为活跃
- **WHEN** 用户被选中后在提交前产生了新的用量
- **THEN** 创建 MUST 失败并返回 `REDEEM_EMAIL_RECIPIENT_INELIGIBLE`，数据库中 MUST NOT 新增任何兑换码

#### Scenario: 金额校验
- **WHEN** 金额 ≤ 0、非有限数或超过 8 位小数或超过 1,000,000
- **THEN** 创建 MUST 被拒绝

### Requirement: 创建幂等
创建请求 MUST 携带 8–128 字符的 `Idempotency-Key`。同一管理员使用同一 key 与相同请求内容重复提交 MUST 返回原批次且不新增兑换码；内容不同 MUST 返回 409 `REDEEM_EMAIL_BATCH_CONFLICT`。

#### Scenario: 网络重试
- **WHEN** 同一请求因超时被前端用相同 key 重发
- **THEN** 系统 MUST 只存在一个批次，响应头 `Idempotency-Replayed: true`

### Requirement: 异步限速投递
批次创建后 MUST 立即返回 202，邮件由后台 Worker 投递。跨实例 MUST 只有一个 Worker 在发送，发送间隔默认 1 秒。发送前 MUST 确认兑换码仍为未使用且未过期，否则该项标记失败 `REDEEM_CODE_UNAVAILABLE`。租约超时（结果未知）MUST 标记失败 `DELIVERY_OUTCOME_UNKNOWN` 且 MUST NOT 自动重发。

#### Scenario: 已退订收件人
- **WHEN** 收件人已通过邮件中的退订链接退订 `redeem.inactive_coupon`
- **THEN** 搜索结果 MUST 标记 `unsubscribed=true`；若仍被选入批次，Worker MUST NOT 发送邮件，明细 MUST 标记为 `skipped` / `RECIPIENT_UNSUBSCRIBED`，且不计入失败、不被「重试失败项」重置

#### Scenario: SMTP 未配置
- **WHEN** 未配置 SMTP 时提交创建
- **THEN** 系统 MUST 返回 `REDEEM_EMAIL_SMTP_NOT_CONFIGURED` 且不创建任何数据

### Requirement: 邮件模板在设置中维护
券邮件 MUST 使用可退订通知事件 `redeem.inactive_coupon`，模板 MUST 能在 `/admin/settings` 邮件模板中按 zh/en 编辑、预览和恢复官方模板，允许的占位符为通用变量与 `coupon_value`、`redeem_code`、`expires_at`、`unsubscribe_url`。发送时 MUST 按收件人语言选择当前保存的模板。批次创建请求 MUST NOT 携带模板。

#### Scenario: 管理员修改模板后预览
- **WHEN** 管理员在设置中保存了自定义英文模板，随后在向导中选择英文预览
- **THEN** 预览 MUST 使用该自定义模板，并以本次券面值、有效期和示例兑换码渲染

### Requirement: 追踪与重试
管理员 MUST 能分页查看批次及各状态计数、查看明细与失败原因，并将失败项重置为待发送。终态批次可删除；删除批次 MUST NOT 删除已生成的兑换码。

#### Scenario: 部分失败
- **WHEN** 批次中部分邮件发送失败
- **THEN** 批次状态 MUST 为 `partial_failed`，成功项保持 `sent`，重试只影响失败项

### Requirement: 审计脱敏
批次创建、重试、删除 MUST 写入审计日志；创建与预览请求体 MUST NOT 以明文进入审计日志。
