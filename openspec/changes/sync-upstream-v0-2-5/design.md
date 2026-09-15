# 设计

## 固定输入
Fork 起点 `107b1a256b23b206751503ae0a1928a51a9bbdb9`，工作树干净。官方 v0.2.5 tag object `4af0e80db1b0bc7626dfb8fb76ccaffc6bb0dc17`，commit `86f93c28ee34cc74b629dafb748bd5ac5ca8c5ea`。不带入后续 main。

## 方法
使用上次官方 tag、当前 Fork、目标官方 tag 逐文件三方合并，先在临时目录预演，按 hunk 解决 30 个冲突文件。保持 HEAD 和暂存区不动。127 个重叠路径单独审查；非重叠 Fork 文件按 Git blob 校验。
OpenSpec CLI、CodeGraph 当前不可用，沿用项目已有 spec-driven 文件结构，使用 Git/rg 调查，不安装工具。

## 必须保留
Kiro credits/cache_read、独立 429 冷却、上游 ID、管理端真名/用户端 Claude 别名；XorPay；Access Ban 与 admin 登录失败封禁；提示词审计；重置卡期限明细/来源幂等/退款事务/旧接口兼容；active_available 默认筛选；Ops 删除；GLM 分类、支付金额分离/隐藏倍率、VersionBadge 禁用在线更新、品牌、Select/ConfirmDialog。
新 OpenCode 平台与 Kiro 同时保留于共享平台、配额、Composite、Wire 和调度快照。上游批量删用户与 Fork 已有能力整合，使用受保护的 Fork 批量接口。
订阅延期采用上游事务行锁，并保留 Fork 的可选月窗口对齐和重复购买发卡，不改业务金额/精度/币种。

## 数据与回滚
296 个既有 SQL 字节不变。新增 `238_opencode_go_platform.sql` 配额和 Composite CHECK 补回 Kiro，监控 provider 沿用官方范围。新增 `238_purge_unlimited_user_platform_quotas.sql` 仅删除三档限额全 NULL 的行；不执行该迁移。上线前须备份这些历史行并验证已有 Kiro 数据、约束和迁移耗时；删除不可由降级二进制恢复。包含新平台数据后，回退须同时处理数据兼容性。
代码交付保持未提交差异，可经审查后单独撤销；不得强制重置工作区。

## 验证
先导入官方 Codex UA 安全测试观察旧实现预期失败。迁移 Kiro 约束和自定义订阅/支付/别名行为使用现有测试与最小兼容回归。最终执行 Go 普通与 unit 套件、vet/build，前端 Vitest/typecheck/lint/build，部署脚本 fixture 测试，以及版本、历史 SQL、编码和差异检查。真实外部服务与数据库不由本次单元测试代替。
