## 1. 规格
- [x] 1.1 proposal / spec / design / tasks

## 2. 后端
- [x] 2.1 迁移 `246_redeem_email_campaigns.sql` 与迁移内容测试
- [x] 2.2 服务层模型、校验、请求哈希、模板渲染及单元测试
- [x] 2.3 仓储：搜索、事务化创建、列表/明细、重试、删除、领取/完成/租约恢复，及 PostgreSQL 集成测试（sqlmock 单测已运行；`-tags integration` 集成测试需 Docker，本地仅编译检查）
- [x] 2.4 Worker 及单元测试
- [x] 2.5 Handler、路由、审计、Wire 接线与 handler 测试

## 3. 前端
- [x] 3.1 API、类型、i18n
- [x] 3.2 群发向导、发送记录、批次详情组件与 RedeemView 入口
- [x] 3.3 组件与 API 测试

## 4. 文档与验证
- [x] 4.1 更新 `docs/FORK_VS_UPSTREAM.md`、`AGENTS.md`
- [x] 4.2 后端 go test / vet / build，前端 vitest / typecheck / build，编码检查

## 5. 追加：分段多选与设置页模板
- [x] 5.1 迁移 246 改为 `segments` / `reference_time` / `skipped_count` 与明细 `skipped` 状态（未发布，原地修改）
- [x] 5.2 服务层分段校验/边界/哈希；仓储分段 SQL 与 skipped 计数；handler 绑定 `segments`、预览 `locale`
- [x] 5.3 注册 `redeem.inactive_coupon` 可退订事件与 zh/en 官方模板；Worker 经通知服务发送并跳过已退订；搜索标记已退订
- [x] 5.4 前端分段多选 Tab、退订标记、设置模板 iframe 预览与深链，历史/详情 skipped 展示；EmailTemplateEditor 展示文案与预选
- [x] 5.5 更新文档并重新运行全部验证
