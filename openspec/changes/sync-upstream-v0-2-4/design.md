## Context
初始 main / HEAD 0e5fd73f6，工作树干净。旧基线 v0.2.3 commit 8fa67d477d6651a744754392a8982ea589c26ae6；目标 v0.2.4 tag object d681d0798064ee0ffff376d19687d12f09fe600f、commit 5de5e2bed035d43591a2e10e51f420ef6a84eb98。官方 VERSION 仍为 0.2.3，本 Fork 设为 0.2.4。

## Decisions
使用旧 tag / 当前 HEAD / 目标 tag 逐文件三方合并，在临时目录预演后写入当前工作树；不移动 HEAD、不暂存、不安装无关依赖。80 个文件含不同于官方的内容，25 个产生文本冲突；逐块保留 Fork 增量并叠加 MiniMax 等官方变化。通过全部非重叠文件逐字节检查、重叠文件相对 tag 的差异检查及回归测试共同验证。

Kiro credits/cache_read、真实上游请求 ID、429 独立冷却、XorPay、Access Ban 与登录自动封禁、提示词审计、订阅重置卡/来源幂等/退款事务、active_available、Ops 删除、Claude 别名、GLM 筛选、VersionBadge/UI 均须保留。

237 重建 user_platform_quotas 与 composite_model_routes 的 CHECK 时加入 kiro；监控 provider 维持原有平台范围并加 MiniMax，不凭空加入 Kiro。历史 SQL 全部原文不动，237 使用现有迁移事务/校验和机制；不改金额精度币种及重试补偿，不执行真实迁移。上线需确认并备份，回滚须与数据库已有平台数据兼容，另行授权。

Redis v9.17.2 升级至上游指定 v9.22.0，沿用 go.sum 验证，修复 nil context 连接池崩溃；保留 Fork 已有更高安全依赖。前端无需新依赖。部署受保护文件不读不写，应用日志限制默认值由官方代码提供，文档列出可选参数。

## Integration Fixes

全量回归检出插件包关闭逻辑语义冲突：Fork 原有显式关闭与上游新关闭 helper 叠加。统一为返回错误、只关闭一次的 helper，在两次 rename 前显式关闭，defer 仅兜底；保留关闭失败中止及临时文件清理。所有安装/重复上传/签名/不兼容插件测试恢复，数据库原件恢复共用同一入口。

管理端 PlatformTypeBadge 对 kiro 保留真实名称分支，其余平台使用共享映射，避免用户端 Claude 别名污染管理端。测试适配 Kiro 额外依赖参数、确认弹窗和 MiniMax 平台数量；监控原四项竞态测试保留并新增隐藏排行不请求测试。

## Validation
先导入官方有效订阅分组可见性测试观察预期失败，再合入实现。237 测试要求 Kiro 与 MiniMax 同时存在并观察原版迁移失败。Go default 和 unit 全量串行执行，避免历史 .entc 并发临时目录冲突；运行 vet/build，前端全量 Vitest/typecheck/lint/build。复核最终文件范围、历史迁移、编码/BOM/乱码/冲突标记。数据库 integration 和外部联调以可用隔离环境为限。

OpenSpec CLI/CodeGraph 不可用，沿用仓库 spec-driven 结构与 Git/rg，不声称 CLI validate 已通过。
