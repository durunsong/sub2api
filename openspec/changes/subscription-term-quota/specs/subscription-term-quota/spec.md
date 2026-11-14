## ADDED Requirements

### Requirement: 周期额度持续到订阅到期
系统 SHALL 以 expires_at 作为订阅周期额度结束时间，保留 monthly_* 字段兼容，不在有效期内自动重新发放该额度。
#### Scenario: 90 天卡重开一小时后
- WHEN 用户消费 90 天卡，已过去一小时
- THEN 两端倒计时为 89 天 23 小时，进度 API 结束时间等于订阅到期时间
#### Scenario: 第 30 或 60 天额度已用尽
- WHEN 90 天订阅仍有效且周期额度已耗尽
- THEN 列表和进度保留用量，自动维护不清零，网关继续拒绝，active_available 不包含该订阅
#### Scenario: 期限结束
- WHEN 到达 expires_at
- THEN 订阅过期，不自动获得新额度或新的有效期

### Requirement: 显式重置及其他限额保持边界
系统 SHALL 保留卡消费原子性、来源期限快照、幂等以及管理员单条和批量重置边界。
#### Scenario: 重复消费请求
- WHEN 同一幂等键重试
- THEN 不重复扣卡、不改变已确定的到期时间、不清掉消费之后的新用量
#### Scenario: 管理员清额度
- WHEN 管理员重置周期额度
- THEN 指定额度清零且产生一条记录，到期时间不变
#### Scenario: 其他额度
- WHEN 日/周订阅额度或钱包平台月额度进入新周期
- THEN 继续沿用各自原有规则
#### Scenario: 延期及旧参数兼容
- WHEN 管理员调整订阅有效期，或旧客户端传入 align_monthly_reset / shift_monthly_reset
- THEN 周期额度结束时间跟随新的到期时间，保留真实用量和窗口起点；旧参数的 0 天请求保持兼容
