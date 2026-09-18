## MODIFIED Requirements

### Requirement: 精确同步官方发布
系统 SHALL 纳入官方 v0.2.6 的允许路径增量，仅保留有依据的 Fork 兼容差异，VERSION 为 0.2.6。
#### Scenario: 发布范围
- **WHEN** 对照 v0.2.5 到 v0.2.6 的文件列表
- **THEN** 131 个允许路径全部纳入或有兼容记录，受保护配置明确列为未同步范围，不引入后续 main。

### Requirement: 保留 Fork 可观察行为
系统 SHALL 保留 design.md 所列全部定制与现有测试。
#### Scenario: 支付输入兼容
- **WHEN** 输入非法字符、超两位小数或超过 max 的充值金额
- **THEN** 恢复最近合法显示值且不发出错误金额，人民币符号和快捷金额保留。
#### Scenario: 定制与上游功能共存
- **WHEN** 使用 Kiro/XorPay/Access Ban/订阅重置卡及新 Codex 设置、兑换分页
- **THEN** 定制链路保持原行为，新功能依照上游契约可用；票据私密状态不暴露到普通 DTO 或导出。

### Requirement: 数据与交付安全
历史 SQL SHALL 保持字节不变，真实迁移及生产部署 SHALL 另行授权。
#### Scenario: 最终检查
- **WHEN** 最终编辑完成
- **THEN** 执行新鲜测试、静态检查、构建和编码检查，并记录未运行项与原因，HEAD 与暂存区不变。
