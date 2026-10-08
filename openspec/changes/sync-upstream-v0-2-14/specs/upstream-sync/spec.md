## ADDED Requirements

### Requirement: 同步固定发布增量
系统 SHALL 同步官方 v0.2.11 到 v0.2.14 的非部署功能、测试与依赖安全修复，版本文件为 0.2.14。

#### Scenario: 发布范围可核验
- WHEN 比较上游目标和最终工作区
- THEN 非重叠路径与目标一致，重叠路径有明确 Fork 兼容差异，不带入 main 后续内容。

### Requirement: 定制兼容
系统 SHALL 保留 docs/FORK_VS_UPSTREAM.md 与 FORK_CUSTOMIZATIONS.md 记录的所有定制。

#### Scenario: 平台与路由共存
- WHEN 增加 TypeSafe 平台及 System One 路由
- THEN Kiro 仍可配置、调度与计费，新入口受 Access Ban 和原有审计链保护，历史迁移不变。

#### Scenario: 支付和订阅共存
- WHEN 使用余额充值优惠或原有订阅支付
- THEN 余额订单按官方优惠规则快照计价，XorPay 和实付/到账展示保留，订阅仍按 Fork 永久卡及退款幂等口径处理。

### Requirement: 可验证交付
同步 SHALL 包含新的安全回归和 Fork 兼容回归，并报告实际检查结果及运行限制。

#### Scenario: 安全修复有效
- WHEN 重放 EasyPay 下单签名伪造支付成功回调
- THEN 拒绝该回调而正常合法回调仍可通过。

#### Scenario: 发布检查
- WHEN 完成最终编辑
- THEN 后端与前端验证、历史 SQL 完整性和 UTF-8/冲突标记检查完成，未运行项明确说明。
