## Context

- Fork 保护基线：`1f38262d1`，main 工作树干净。
- 官方 v0.2.1 commit：`578785ee7fb35030b094b69624efe25670a36f5f`。
- 官方 v0.2.2 tag object：`86495464d82f78da55792db6aa5568ee1302716a`。
- 官方 v0.2.2 commit：`5485f368b29d05adb95a00f71801c7c23d8f48af`。
- 上游 VERSION 仍为 0.2.1；沿用 Fork 惯例按 release tag 显示 0.2.2。
- CodeGraph 和 OpenSpec CLI 在当前会话不可用；使用 Git、rg 和既有 spec-driven 文档结构。

## Goals / Non-Goals

验收标准：全部 release 增量已核对，Fork 定制在重叠文件内保留，既有 SQL 原文不变，Go 默认/unit 测试、vet/build、前端 Vitest/lint/typecheck/build 通过，最终差异无冲突标记、UTF-8 错误、BOM 或替换字符。环境受限检查必须记录真实限制。

不访问生产、外部支付/OAuth 或秘密配置，不执行迁移，不改变原有金额/币种、Kiro credits、重置卡期限与幂等规则，不做无关重构。

## Decisions

### 三方增量合并

使用旧 tag、当前 Fork 和目标 tag，逐文件合并。没有重叠的文件应用官方增量；重叠文件通过 Git 三方文本合并，再审阅 Fork 差异。重命名文件必须带上 Fork 原文件中的额外测试。Ent 生成源码与 Fork schema 一致，不能丢掉 Kiro/IPBan 字段或错位索引。

### 网关与权限

新增模型白名单校验位于 API Key 鉴权之后、Composite 模型改写之前，使用客户端模型名。所有路由保留 Access Ban；Kiro 仍保留在平台枚举、简易模式分组、模型列表与 Composite 路由中。鉴权缓存契约随新字段升级。

### 支付、计费与数据

支付/管理端发货使用受信任专用兑换入口，隔离公共兑换失败计数，但保留输入校验、分布式锁、事务及返佣规则。余额发货验证兑换码金额（原精度 1e-8）、状态和归属；支付订阅继续透传订单来源创建重置卡，兑换订阅继续使用兑换来源幂等。保留 USD/token 口径、冻结 pricingAt、DeepSeek 峰谷和仅一次 max 倍率计算。

235 迁移在事务内将 groups.models_list_config 重命名为 model_allowlist，原数据不变。非空配置升级为准入白名单是上游明确行为，升级前须审核。Fork 的 157、159/160、224/225 以及其他历史迁移不得重写。本次只交付迁移源码；部署和数据库回滚另行确认，不能直接用旧二进制回退到已重命名 schema。

## Validation

先运行关键 Fork 基线，再导入上游行为测试观察 RED，合入实现后观察 GREEN。最终运行全量本地默认/unit 与前端检查，重点检查 Kiro、白名单、Access Ban、支付/兑换/订阅、成本、Wire 和数据库 schema 映射。记录测试覆盖范围，不把静态检查等同于真实 PostgreSQL/Redis 或支付联调。
