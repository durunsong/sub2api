## MODIFIED Requirements

### Requirement: 精确同步官方发布
系统 SHALL 纳入官方 v0.2.5 的允许修改路径增量，仅保留有依据的 Fork 兼容差异，VERSION 为 0.2.5。
#### Scenario: 发布范围
- **WHEN** 比较 v0.2.4 与 v0.2.5
- **THEN** 允许路径全部同步或记录兼容处理，受保护配置明确列为未同步范围。

### Requirement: 保留 Fork 可观察行为
系统 SHALL 保留 design.md 所列全部定制；共享平台集合同时支持 Kiro 与 OpenCode。
#### Scenario: 平台与订阅兼容
- **WHEN** 管理员使用 Kiro 或 OpenCode、进行订阅批量操作和有效固定期限重复分配
- **THEN** 两个平台均可用，重复分配仍发永久期限快照卡，延期月窗口对齐与事务互斥同时有效。
#### Scenario: 用户界面
- **WHEN** 用户筛选 Claude、购买 GLM 套餐或管理 API Key
- **THEN** Kiro 别名/搜索和 GLM 分类保持，并可用新增批量操作。

### Requirement: 数据与交付安全
既有 SQL SHALL 保持字节不变，新增平台约束 SHALL 包含 Kiro，真实迁移和生产部署 SHALL 另行授权。
#### Scenario: 新鲜验证
- **WHEN** 最终代码修改完成
- **THEN** 执行并记录实际测试、静态检查与构建结果，UTF-8 无 BOM/乱码/冲突标记；无法执行的检查明确披露。
