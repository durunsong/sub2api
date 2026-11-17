# v0.2.22 整合设计

## Context
初始 HEAD e9c30cb40，工作区干净。使用 v0.2.21 作为共同内容基线逐文件三方合并；不合并上游 main。OpenSpec CLI 未安装，沿用仓库 spec-driven 文件结构，不安装额外工具。

## Decisions
- 243 路径无 Fork 差异，79 路径自动三方合并，16 路径需解决冲突，VERSION 已等于上游但须按发布标签改为 0.2.22，deploy/config.example.yaml 不覆盖。
- 冲突合并同时保留 Kiro 直连/credits/粘性 TTL/代理失败拒绝直连、冻结 pricingAt/context、用户端成本隐藏和上游新增行为。
- Ent 与 Wire 基于合并后的 schema/provider 重新生成，保留 Kiro 顺序、封禁注入和邮件 Worker 清理；新增扣费通道纳入停机。
- 扣费按上游同用户合并事务与幂等规则，保留 Fork 订阅全期用量：monthly 字段不能恢复为每 30 天补额。Vertex 1.1 倍仅适用真实多区域且默认模型价，不叠加自定义价；Kiro credits 不参与该倍数。
- 两份新 251 SQL 与既有 251_subscription_quota_reset_history.sql 按完整文件名共存。历史 SQL 字节保持不变，不执行迁移。上线前备份并验证新列及非阻塞汇总重建；回退不删除表/索引，订阅全期口径及旧 pending 退款降级限制保持。
- Go 声明跟随上游 1.27.4；jsonschema-go 仅由间接改为直接依赖，版本不变。前端无需更新依赖。保留 Fork 镜像源与安装策略。

## Validation
上游回归先观察旧代码失败，随后前端全量 Vitest/lint/typecheck/build，后端 unit/build/vet/embed，关键并发计费/调度 race；检查历史 SQL、Fork 独有文件、最新 7 个定制提交相关路径、UTF-8/BOM/冲突标记和 git diff。真实数据库/外部供应商测试未运行时明确列明。
