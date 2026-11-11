# 同步设计

## 合并方法
以官方 v0.2.20 为 base，初始 Fork HEAD 1e8e2321696c9d41d625931a7455f12f8eb70425 的当前工作区为 ours，v0.2.21 为 theirs，在临时目录逐文件三方预合并，不改变 Git 索引。检查所有重叠文件；确认无冲突后分批写回。版本按目标 release 设为 0.2.21。

## 定制兼容
Kiro credits、cache_read 不重复计入 input、冻结 pricingAt、真实上游请求 ID、独立 429 冷却、平台目录和 Wire 注入顺序继续保留。缓存 TTL 扩展不得改变 Kiro 的缓存统计。Access Ban、邮箱格式校验、登录失败自动封禁与上游刷新令牌轮换并存。重置卡来源幂等、期限快照、支付退款事务及邮件 Worker 清理不变。
前端保留用户侧 Claude 别名、GLM 分类、XorPay、倍率隐藏、禁用在线更新、客服二维码和批量兑换。

## 数据和风险
新增 249_add_usage_logs_api_key_cost_index_notx.sql 与既有 249_mcp_oauth.sql 按完整文件名共存；250_openai_cache_write_ttl.sql 扩展缓存 TTL 计数。只交付文件，历史 SQL 不改。上线前备份并验证大表并发索引和新增列；不得在本任务启动服务自动执行迁移。保持 USD/token/credits 既有精度、事务及幂等规则，按上游规则区分 30m/12h，并保留既有通用缓存写入价覆盖语义。

## 验证策略
先引入上游回归测试观察旧版失败，再合入对应实现。运行完整前端回归与后端 unit/build/vet；对缓存计费、令牌轮换、封禁、Kiro、XorPay、永久重置卡和并发槽补充重点核验。不依赖真实数据库、供应商或生产数据。
OpenSpec CLI 未安装，沿用仓库 spec-driven 文件结构并进行本地结构、路径和编码检查。

## 审查发现的上游计费缺口
v0.2.21 的新缓存 TTL 分桶在 WS v2 Usage 和 ResponsesUsage 类型转换中未贯通，导致 Messages/Chat 兼容终态覆盖已解析的桶、WS 逐轮计费丢失桶。补齐这些既有入口属于本次缓存计费能力的完整性修复，不新增计价策略。
将目标版本既有严格解析规则共用到 HTTP、WS 和 typed usage；聚合量、30m 和 12h 必须作为同一快照替换，缺失、非法或不相符明细继续回退既有通用价。保持 Kiro credits 的独立传递和缓存计量。用真实 handler、WS adapter → RecordUsage 的回归验证 token、扣费和账号统计成本，并覆盖逐轮累计与终态替换。
