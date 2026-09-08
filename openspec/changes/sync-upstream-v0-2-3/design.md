## Context

Fork HEAD ebd2d484f，初始工作树干净。官方旧基线 v0.2.2 commit 5485f368b29d05adb95a00f71801c7c23d8f48af；目标 v0.2.3 tag object fe2b5c04b1c9503fba7e01a099b206f14867cfc1，commit 8fa67d477d6651a744754392a8982ea589c26ae6。官方 VERSION 仍为 0.2.2，Fork 按 release tag 设为 0.2.3。

## Decisions

使用旧 tag / 当前 Fork / 新 tag 逐文件三方合入；不移动分支或暂存。4 个代码重叠文件是 account_test_service.go、gateway_anthropic_passthrough.go、gateway_count_tokens.go、gateway_upstream_request.go，必须保留 Kiro provider、直连分流、credits 与按账号心跳。其余官方增量原样纳入，其余 Fork 代码逐字节不变。

输出上限和鉴权按实际上游 URL 判断，保持官方精确主机/协议限制与模型映射后判断，防止仅凭平台或用量配置影响其它供应商。Kiro credits、cache_read、订阅来源幂等、USD/token、支付事务、登录自动封禁及品牌规则不变。

236 使用 search_path/regclass 定位 groups：仅旧列则重命名，两列并存仅回填空新列，两列皆无则补列，最终 NOT NULL DEFAULT '{}'。旧非空白名单优先保留；全部历史 SQL 原文不变。本次不执行迁移。真实升级应先备份并核对 schema/白名单；应用迁移采用现有事务及 checksum 重试机制，不通过修改历史迁移回滚；回滚需要与数据库结构匹配，另行确认。

## Validation

导入账号模型显示名测试，观察旧版字段为空的预期失败，再合入实现。运行 Go 默认和 unit 测试、vet/build、前端 Vitest/typecheck；核对全部官方增量、Fork 文件及历史迁移完整性，检查 UTF-8/BOM/替换字符、冲突标记及最终差异。PostgreSQL 集成测试取决于本地隔离测试工具可用性。

OpenSpec CLI 与 CodeGraph 当前不可用，沿用已有 spec-driven 目录结构，以 Git/rg 读取和核对。
