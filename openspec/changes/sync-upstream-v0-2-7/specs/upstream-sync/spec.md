## ADDED Requirements

### Requirement: 精确同步目标版本
系统 SHALL 合入官方 v0.2.7 tag 的功能及修复，并显示版本 0.2.7。

#### Scenario: 目标版本范围
- **WHEN** 比较 v0.2.5 与 v0.2.7 的路径清单
- **THEN** 每个允许路径均同步，差异仅为已记录 Fork 兼容内容；不包含 tag 外票据模块

### Requirement: 保留定制行为
系统 SHALL 保留现有 Kiro、支付、封禁、重置卡、审计及 UI 定制。

#### Scenario: 同步网关及支付代码
- **WHEN** 新 Seedance 路由或金额输入修复启用
- **THEN** 新路由仍经过 Access Ban；非法输入恢复最近合法值，金额上限与人民币展示继续生效

#### Scenario: 数据与定制回归
- **WHEN** 执行验证
- **THEN** 298 个历史 SQL 字节不变，非重叠定制文件保持原样，相关行为测试通过

#### Scenario: 插件凭据状态边界
- **WHEN** 插件保存了账号 ID，管理员随后将该账号禁用
- **THEN** 宿主不再通过该 ID 返回凭据；仅暂停调度但仍 active 的账号保持可用

### Requirement: 验证与交付可追溯
升级 SHALL 附带实际检查结果及未执行检查的原因。

#### Scenario: 最终交付
- **WHEN** 声明升级完成
- **THEN** 记录测试/类型/lint/构建、编码与差异检查结果，不宣称已部署或完成真实供应商验证
