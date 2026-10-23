# 同步设计

## 基线与合并
从干净 Fork HEAD 49f13ea62 开始，以官方旧标签为共同基线逐文件三方合并。先在临时目录预合并并记录范围，不修改 Git 索引。冲突逐段保留两侧行为；已合入 PR #8139 不重复追加。VERSION 明确设为 0.2.18（官方文件仍为 0.2.17）。

## 定制边界
重点检查 account 配额和模型映射、Kiro Responses 桥接缓存/请求头、账号创建编辑 UI、Kiro/Vertex 配额展示。保留 Kiro credits、缓存读取不重复计入 input、429 独立冷却、真实上游请求 ID；保留严格代理解析、提示词审计、永久卡及邮件 Worker 注入。未触及的 Fork 独有文件与历史迁移按字节验证。

## 数据和资金
唯一新 SQL 246_enable_gemini_image_generation_groups.sql 把存量 Gemini/Antigravity 分组生图开关开启以兼容升级前行为，不涉及 composite/Kiro。与 Fork 246_redeem_email_campaigns.sql 按完整文件名并存，先验证迁移加载顺序和记录键；不执行 SQL。部署前需备份并审核分组状态，二进制降级不会撤销这个数据变更。
流式断连仍记录已获取用量；沿用现有 USD 精度、请求计费幂等、事务和冻结 pricingAt，不改变 Kiro credits 或卡消费资金口径。账号配额按设置周期更新，Vertex service_account 加入现有配额链。

## 安全依赖与构建
跟随上游升级已有 Go 工具链到 1.27.3、DOMPurify 到 3.4.17、follow-redirects 到 1.16.1；不新增依赖类别。Docker/CI 同步补丁版本，保留 Fork 镜像源和构建选项。优先现有环境；若需同步依赖，仅按锁文件且禁用 lifecycle scripts。验证版本、构建和回归。

## 验收
先引入上游 Vertex 配额回归，在旧实现观察行为失败，再合入实现。后端 build/vet/unit（含 migrations），前端全量 Vitest、lint/typecheck/build；复核定制差异、历史 SQL、版本、编码/BOM/冲突标记。OpenSpec CLI 未安装，沿用仓库 spec-driven 文件格式，不声称 CLI validate 通过。未运行真实外部系统集成须注明。
