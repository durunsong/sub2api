# 升级设计

## 固定输入
- Fork 起点：b92ce2353，初始工作区干净。
- 基线 v0.2.9：4c00df2e0183e2c70b7fa8ba45914205e36aad0c。
- 目标 v0.2.10：2f3fed2fdb0787141294cec81487a5df30426f7f；tag object e8cbebf37505922d30ffccd0a5bc19c4beec60ac。

## 合并方法
按官方标签增量逐文件三方合并，不覆盖整树，不暂存。42 个重叠路径中，Wire、billing_service、Responses 转换和 Dashboard 测试有文本冲突，逐段合并并复核 Fork 差异。使用已有依赖；CodeGraph/OpenSpec CLI 当前不可用，使用 Git/rg 和既有 spec-driven 文档结构。

## 兼容与安全
Claude 原生重置额度查询是管理端只读查询，与 Fork 用户订阅重置卡独立。风控白名单仅按显式平台用户 ID 配置，默认空；保留证据且不处罚，不绕过 Fork Access Ban。Sonnet 5.5 按官方价格目录增加模型；保留请求级 context、冻结 pricingAt、单次推理倍率、DeepSeek 峰谷与 Kiro credits/cache_read 计费。不改金额精度、币种、交易/退款幂等、订阅事务与历史账务，无迁移或历史数据重算。

## 验证与回退
先导入 Sonnet 签名思考回传测试观察旧实现失败，再合入实现。保留既有 Fork 回归并核对非重叠文件。工作区源码可按本次差异回退，不自动回退，不连接真实数据库、支付或供应商。

## 审查后的兼容说明
独立 PromptEngine 与 Access Ban/登录防护继续执行，新白名单仅影响官方内容审核与 cyber_policy。修订双语白名单提示，避免误解为绕过全部本地安全规则；不改变安全行为。因此 76 个非重叠路径中，74 个与官方目标逐字节一致，两个 settings.ts 仅调整该提示。

Sonnet 5.5 依据官方目录采用 USD/百万 token：输入 2、输出 10、缓存写入 5 分钟 2.5/1 小时 4、缓存读取 0.2；不重算历史记录，历史 Sonnet 4.5/4.6 与 Haiku 4.5 的 Fork 兜底价格保持原样。
