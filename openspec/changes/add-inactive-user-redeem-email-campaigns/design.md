## Context

参考 Dowsure LLM Router 的「兑换码 + 群发邮件」链路（批次表 + 后台租约 Worker + 单 leader 限速），但本 Fork 的需求是「系统按收件人自动生成余额券」，因此不让前端粘贴或配对兑换码。

## Decisions

### 1. 筛选口径
- 使用 `usage_logs`（已有 `(user_id, created_at)` 索引）取每人 `MAX(created_at)` 作为 `last_used_at`，不新增冗余列。
- 分段互斥：`last_used_at` 为空 → `never_used`；`>= t1` → 活跃（永不入选）；`[t2, t1)` → `inactive_1_2m`；`[t3, t2)` → `inactive_2_3m`；`< t3` → `inactive_3m_plus`。`tN = now.UTC().AddDate(0, -N, 0)` 均在服务端计算，前端只传 `segments`（非空子集，取并集）。默认选前三段。
- 搜索结果附带 `last_used_at`、`segment` 与 `unsubscribed`（逐个查询退订偏好，已退订者前端默认不勾选）。

### 2. 数据模型（迁移 246，原生 SQL）
- `redeem_email_batches`：筛选快照（`segments TEXT[]` CHECK 为合法非空子集、`reference_time`）、券参数（`coupon_value NUMERIC(20,8)`、`code_expires_at`）、五种状态计数（pending/sending/sent/failed/skipped，CHECK 求和等于总数）、`created_by` + `idempotency_key_hash` 唯一、`request_hash`。不保存模板。
- `redeem_email_items`：收件人快照（用户、邮箱、名称）、`redeem_code_id`/`redeem_code`、投递状态、租约与错误信息。`redeem_code_id` 全局唯一，保证一个兑换码只属于一条明细。用户或兑换码被硬删除时外键置空，保留发送历史。

### 3. 创建事务
1. `pg_advisory_xact_lock` 锁定「管理员 + key 哈希」，串行化同 key 并发。
2. 查找已有批次：`request_hash` 相同则回放，不同则 409。
3. `FOR SHARE` 锁定选中用户行，再用与搜索相同的分段 SQL 重新查询，数量不一致即 `REDEEM_EMAIL_RECIPIENT_INELIGIBLE`。
4. 插入批次，逐个用户生成 32 位兑换码，`INSERT ... ON CONFLICT (code) DO NOTHING RETURNING id`（冲突时重新生成，最多 5 次，避免唯一冲突中止事务），备注 `redeem_email_batch:<id>`，再插入明细。
5. 提交。任何错误回滚。

`request_hash` = SHA-256（规范化 JSON：按固定顺序去重的分段、升序用户 ID、金额 8 位小数字符串、有效天数）。基准时间不参与哈希，因此重放时以原批次为准。

### 4. 投递 Worker
- `tryAcquireSingletonLeaderLock`（Redis，失败回退 PG advisory lock）保证跨实例单发送者；新 leader 先等待一个发送间隔，维持全局 1 封/秒。
- `ClaimNext` 使用 `FOR UPDATE SKIP LOCKED` 领取一条并写入 attempt token 与 45 秒租约；`CompleteAttempt` 以 token 做 CAS，迟到的结果不会覆盖。
- `RecoverExpiredSending` 把过期租约标记为 `DELIVERY_OUTCOME_UNKNOWN` 失败，不自动重发（SMTP 可能已投递），由管理员判断后重试。
- 发送前查询退订偏好：已退订 → 明细 `skipped` / `RECIPIENT_UNSUBSCRIBED`（不计失败、不被重试）；查询出错 → `failed` / `UNSUBSCRIBE_CHECK_FAILED`。
- 通过 `NotificationEmailService.Send` 渲染并发送：语言由 `ResolveRecipientLocale` 决定，变量 `coupon_value`、`redeem_code`、`expires_at`（长期有效按语言显示「长期有效」/“Never expires”），`unsubscribe_url` 由通知服务签名生成；不传 SourceType/SourceID，避免为每封邮件写 settings 投递键。模板错误记 `TEMPLATE_INVALID`，SMTP 配置错误记 `SMTP_NOT_CONFIGURED`。

### 5. 邮件模板
- 事件 `redeem.inactive_coupon`：`category=marketing`、`Optional=true`，占位符为通用变量 + `coupon_value`、`redeem_code`、`expires_at`、`unsubscribe_url`；官方 zh/en 模板沿用 `notificationEmailCard` 视觉外壳。
- 在独立 Fork 文件中由 `init` 追加到 `notificationEmailEventOrder` / `notificationEmailEventDefinitions` / `notificationEmailOfficialTemplates`，上游文件只在 `sampleVariables` 中调用 `notificationEmailForkSampleVariables` 以便设置页预览显示示例券面。
- 批次不快照模板：发送时取当前保存的版本，管理员可在批次进行中修正模板。向导预览调用 `POST preview`（`locale` + 券参数），以保存的模板和示例兑换码渲染，前端放入 `sandbox` iframe。

### 6. 接口
`/api/v1/admin/redeem-codes/email-batches`：`GET defaults`、`POST recipients/search`、`POST preview`、`POST`（202）、`GET`、`GET :batch_id`、`GET :batch_id/items`、`POST :batch_id/retry-failed`、`DELETE :batch_id`。均在 admin 鉴权、面板限流与审计中间件之下；创建路由的请求体整体不入审计库。

## Risks / Trade-offs

- 同一用户可在多个批次中各收到一张券；本期不做跨批次去重，管理员需注意活动节奏。
- 已退订用户若被手动勾选，仍会生成兑换码但不发邮件；码保留在列表中可按备注禁用。
- 搜索时逐个查询退订偏好（最多 1000 次 settings 读取），可接受但不是批量查询。
- 批次进行中修改模板会影响尚未发送的邮件，这是有意的（与其他通知事件一致）。
- 结果未知的投递需要人工确认后重试，可能导致重复邮件，但同一兑换码只能核销一次，资金风险受控。
- `usage_logs` 很大时搜索为相关子查询，单次最多返回 1001 行；已有 `(user_id, created_at)` 索引支撑。

## Rollback

迁移只新增两张表；回滚应用时这两张表可保留不影响旧版本。已发出的兑换码是普通 `balance` 码，可在兑换码列表按备注筛查或禁用。
