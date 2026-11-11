## ADDED Requirements

### Requirement: 精确同步发布增量
系统 SHALL 合入官方 v0.2.20 到 v0.2.21 的功能和修复，不带入后续 main 内容，版本 SHALL 为 0.2.21。

#### Scenario: 发布源码与本地定制重叠
- **WHEN** 官方更新已有定制的共享文件
- **THEN** 三方合并同时保留官方增量与全部定制行为，并运行对应回归

### Requirement: 定制与本地工作保持
系统 SHALL 保留 Kiro、XorPay、Access Ban、提示词审计、永久重置卡、群发余额券、批量兑换和 UI 定制；初始未提交的三个兑换文件 SHALL 字节不变。

#### Scenario: OpenAI 缓存 TTL 扩展
- **WHEN** 保存及计费 30m/12h 缓存写入 token
- **THEN** 上游分档定价生效，既有 Kiro credits/cache_read 和冻结计价时刻保持原语义

### Requirement: 数据与运行边界
历史迁移 SHALL 字节不变，新增两个迁移 SHALL 按完整文件名共存，不执行迁移或部署。

#### Scenario: 验证升级交付
- **WHEN** 完成代码整合
- **THEN** 后端 build/vet/unit、前端 lint/typecheck/Vitest/build、定制保留和编码检查均记录实际结果，环境限制明确列出

### Requirement: 各 OpenAI 入口使用同一缓存 TTL 计费语义
系统 SHALL 将上游有效的 30m/12h 缓存写入明细贯通 Responses HTTP/SSE、WS v2、Messages 和 Chat Completions 兼容入口的最终计费，保持旧聚合价格回退语义。

#### Scenario: WebSocket 或协议转换终态携带 12h 用量
- **WHEN** 终态 usage 的两个缓存桶是非负整数且其和等于缓存写入总量
- **THEN** 扣费、用量落库和账号统计 SHALL 应用既有配置的各 TTL 单价，不能因中间类型转换清零明细

#### Scenario: 新聚合快照没有有效明细
- **WHEN** 后续 usage 替换聚合量而缺失或提供不一致的 TTL 明细
- **THEN** SHALL 清除旧快照的明细并按通用缓存写入价回退，不沿用旧桶或推断请求 TTL
