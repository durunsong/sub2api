# 同步设计

## 固定输入
- Fork 起点：352146f35，初始工作区干净。
- 官方基线 v0.2.10：2f3fed2fdb0787141294cec81487a5df30426f7f。
- 目标 v0.2.11：96f4c115c9749078f90cbf210a01d39baf3f53b6；远程 tag object 881a722349105e21aaa38e11de90dd05046ebe4a。
- 官方 VERSION 为 0.2.10，本 Fork 按发布标签设为 0.2.11。

## 合并方法
沿用前次同步的逐文件三方合并；仅导入标签增量，不整体覆盖、不带入 main 后续提交、不使用暂存区。24 个非部署重叠路径逐项复核，Wire 冲突保留 Kiro/IPBan 参数，同时前移幂等协调器并接入 Claude 重置 provider。CodeGraph/OpenSpec CLI 不可用，沿用现有 spec-driven 文件结构与 Git/rg。

## 计费、安全与兼容
余额预留按上游实现：USD 费用估算、Redis 原子准入、默认 900 秒 TTL、长请求续租，持有至 handler 和异步计费任务结束；Redis/未知价格默认 fail-open。它是并发准入估算，不是资金冻结或最终扣费，不改变既有金额精度、事务、重试、退款或历史账务。不应用于订阅模式；Kiro credits/cache_read 仍按原实现结算。保留请求 context、冻结 pricingAt、单次推理倍率与 DeepSeek 分时定价。
API Key 默认每用户最多 200 个未删除 Key、每小时创建 60 次；删除不返还次数，零值可关闭。Claude 原生重置仅供管理员对原生 Claude 账号操作，由服务端选 grant、幂等协调器及 Redis lease 保护，并要求界面确认；不触碰用户订阅永久卡。
GPT-6.1 Sol 按上游目录计价；Astra Ultrafast 模型倍率 6，与站点 Fast 定价独立；既有 Fork 历史模型兜底价格不变。
部署示例 config.example.yaml 不同步，不读取本地秘密或生产配置；功能默认值随代码提供。

## 验证与回退
先导入官方 GPT-6.1 Sol 桥接测试观察旧版本失败，再合入实现；复用上游并发预留、原生重置与 API Key 限制测试及全部现有 Fork 回归。所有修改留在工作区，可按本次差异审查/回退，不改数据库和真实服务。真实供应商、支付和数据库环境验证不在本次范围。

## 合并适配与审查
唯一文本冲突为 wire_gen.go；保留 Fork AdminHandlers 的 Kiro/IPBan 参数，采用上游的新 Claude 重置 provider。新增 gateway_inflight_reservation_test.go 与 unit 标签的 gateway_handler_claude_code_fallback_test.go 的 NewGatewayService 调用各补齐两个 Fork Kiro 参数；这两处是纯上游文件的测试适配。
独立只读审查确认 23 个非 VERSION 重叠文件的 Fork 增删行完整保留；588 个可检查非重叠定制文件及 301 个历史 SQL 不变。65 个非重叠上游文件中，63 个逐字节一致，剩下两个仅包含上述测试参数适配。
