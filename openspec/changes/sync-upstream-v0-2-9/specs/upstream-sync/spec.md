## ADDED Requirements

### Requirement: 精确同步目标版本
系统 SHALL 合入官方 v0.2.9 的允许路径并显示版本 0.2.9。

#### Scenario: 发布范围
- **WHEN** 对比官方 v0.2.8 与 v0.2.9
- **THEN** 114 个非部署路径均核对同步，所有偏离仅为 Fork 兼容内容或版本修正，不带入 tag 后变更

### Requirement: 保留 Fork 行为
系统 SHALL 保留 Kiro、XorPay、Access Ban、重置卡、审计、Ops 及现有 UI 定制。

#### Scenario: 账号统计成本合并
- **WHEN** 同步账号长上下文计费开关
- **THEN** 同时保留请求级 context、冻结计价时刻、单次推理倍率及 DeepSeek 峰谷计价，Kiro credits 和订阅 token 计数保持原样

#### Scenario: 数据及非重叠模块
- **WHEN** 检查最终差异
- **THEN** 历史 SQL、部署配置、非重叠 Fork 模块字节不变，封禁/重置卡/支付/品牌相关回归通过

### Requirement: 可追溯验证
升级 SHALL 记录实际测试、构建、编码检查和环境限制。

#### Scenario: 交付
- **WHEN** 完成升级
- **THEN** 记录前后端新鲜验证结果，不将源码验证表述为已部署、真实迁移或真实供应商联调成功
