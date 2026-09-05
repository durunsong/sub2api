## Context

- 初始 HEAD：`3f050cea2`，main 工作树干净，落后已有 origin/main 一个提交。
- Fork 保护基线：`e316f5189`，新增管理员登录失败 IP 自动封禁及重置卡提示。以未暂存增量纳入本次工作，不移动分支或创建提交。
- 官方旧基线：`aa236488351eb71e120fc2b6fb32e36b0374c918`（v0.2.0）。
- 官方目标 tag object：`adc26f68f687685e847bfb997559f48e79cac475`。
- 官方目标 peeled commit：`578785ee7fb35030b094b69624efe25670a36f5f`（v0.2.1）。
- 官方增量与 Fork 基线有 87 个重叠文件。官方 tag 内 VERSION 为 0.2.0，按本 Fork 惯例显示 release 版本 0.2.1。
- 现有 Go 1.27.0、Node 22.23.2、pnpm/Vitest 可用；OpenSpec CLI 不在 PATH，沿用仓库 spec-driven 文件结构，不安装依赖。

## Goals / Non-Goals

目标是完整移植 release 增量，证明上游行为和定制可共存。验收包括增量覆盖、重叠文件审阅、Fork 回归、Go test/vet/build、前端 test/lint/typecheck/build 和 UTF-8/差异检查。

不引入 tag 后 main 内容，不更改本地或秘密配置，不访问生产，不运行迁移，不创建提交或推送，不增加无关功能或依赖。

## Decisions

### Tag 增量与三方合并

使用旧 tag、Fork 基线和目标 tag 做逐文件三方合并，避免共同祖先过旧带来的重复合并。可直接应用的 hunk 使用 Git 原生补丁；文本冲突逐项保留双方语义，禁止整文件选 upstream/ours。导入的 Ent 文件是与 schema 配套的已跟踪源码。

### 金额、状态与数据

计费精度与币种沿用官方 v0.2.1 和当前 Fork 的 USD/token 规则；max 推理倍率来自渠道配置或官方 Fable 5.1 默认值，Kiro credits 与 cache_read 去重口径保留。Alipay 待支付订单补偿沿用已有支付确认/幂等发货路径，XorPay 工厂、回调、退款和订阅发卡来源保持不变。

追加官方 `232_add_usage_log_upstream_request_id.sql`、`233_add_usage_log_upstream_request_id_index_notx.sql`、`234_channel_max_reasoning_effort_multiplier.sql`、`234_group_codex_models_manifest_config.sql`，按完整文件名与旧同号迁移并存。历史请求 ID 为 NULL，新模型目录配置默认关闭，渠道 max 倍率可空且必须为正，索引使用非事务并发创建。保留 Fork 157 的 Kiro/Grok 配额、159/160 访问封禁和 224/225 重置卡原文。

本次不执行生产数据变更。回滚在部署前通过审阅未提交 diff 决定；实际生产迁移与回滚需另行确认，不删除或重写历史迁移。

### 定制保护矩阵

| 范围 | 保留要求 | 验证 |
| --- | --- | --- |
| Kiro | OAuth、模型、配额、独立 429、credits、cache_read、Composite、Wire 顺序 | pkg/service/handler/Wire 与品牌测试 |
| XorPay | 支付、回调、退款、支付宝文案、金额分离 | provider/payment tests |
| Access Ban | 四类规则、可信 IP 自动封禁、Auth/Gateway 中间件 | accessban/IPBan/auth/middleware tests |
| 订阅 | 永久卡期限快照、来源幂等、消费与退款事务、active_available、卡提示 | repository/service/frontend tests |
| 提示词审计/Ops | 扫描、策略、控制台、错误筛选删除 | securityaudit/Ops tests |
| UI | Claude 别名、GLM 分类、品牌、隐藏在线更新、自定义对话框 | Vitest 与增量审阅 |

### 验证顺序

先运行关键 Fork 基线，再导入上游测试观察缺失行为，移植实现后重跑。最终运行 Go 全量默认与 unit 测试、vet/build，前端全量 Vitest、lint、typecheck/build。数据库/Docker 集成验证按本机实际可用性报告，不将编译或静态检查等同于数据库执行成功。

逐文件确认最终结果可反向识别所有官方 hunk，同时保留 Fork 相对旧 tag 的扩展。最终检查无冲突标记、补丁残留、秘密/依赖变更及乱码。
