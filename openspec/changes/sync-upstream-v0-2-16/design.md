# 同步设计

## 合并与兼容
以官方 v0.2.15、Fork HEAD、官方 v0.2.16 做逐文件三方比较。与 v0.2.15 相同的 144 个路径直接采用 v0.2.16；69 个双方都改过的路径用 merge-file，冲突只出现在 8 个文件。分组 schema 在 description 后插入 `i18n`，Kiro 五个字段仍排在 `rpm_limit` 之后，生成代码的字段下标按 schema 顺序修正为 58–68。

## 资金与安全
显式缓存存储费沿用上游的纯存储路径，token 计费继续传入 Fork 的冻结 `pricingAt` 和 Kiro 兜底选项。报价接口按账号平台设置同一 Kiro 兜底。Gemini 缓存路由挂在已有 Google Access Ban 分组上。支付、永久重置卡、XorPay 与提示词审计不改语义。

## 历史数据与回退
新增 `242_group_i18n.sql` 与既有 `242_drop_platform_check_constraints.sql` 按完整文件名并存。`243`/`244`/`245` 新建 Gemini 缓存表并追加渠道计费与修改租约列。历史 SQL 不改。不自动执行迁移；上线须先备份并测试迁移。代码以未暂存差异交付。

## 验证
后端 `go build ./...`、`go vet ./...`、`go test -tags=unit ./internal/...`。前端覆盖本次冲突和新增界面的 Vitest，以及 `pnpm typecheck`。结果记录于 verification.md。
