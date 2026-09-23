# 设计

## 固定输入
起点为 Fork `HEAD cf2b9c670`（官方 v0.2.7 之上的定制提交）；目标为官方 tag `v0.2.8`，peeled commit `fd80b08c90b55edcad5b00171b53f08721d30da1`。官方相对 v0.2.7 增量为 473 个文件、+28,638/-2,096。

## 合并策略
以官方 v0.2.8 为新基线，使用 `git diff v0.2.7..HEAD --binary` 三方回放 Fork 定制，避免官方整树覆盖定制。真实重叠点逐 hunk 合并：

- Wire 重新生成，保留 Kiro 在 Grok 前、IPBan 注入 AuthService、插件账号目录注入。
- 网关保留 Kiro 直连、缓存/credits、Access Ban 和 Fork 的简单模式限额/成本口径，同时吸收上游最终 reasoning effort、OpenCode Go 和协议兼容修复。
- 配额前端/迁移保持 Kiro、MiniMax、OpenCode Go；订阅重置卡、XorPay、VersionBadge、批量用户删除按现有 Fork 实现保留。
- 插件凭据解析要求账号 `active`；仅暂停调度但仍 active 的账号继续可用。

新增上游 SQL 原样纳入，不执行真实数据库操作；历史 Fork 迁移字节不改。版本、Fork 差异文档和本变更记录同步更新。
