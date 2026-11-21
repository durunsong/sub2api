# Implementation Plan

Goal: 在既有余额券向导中支持最近 7 天和 30 天活跃用户。
Architecture: 扩展共享分段 SQL 与类型，新增兼容迁移，复用现有邮件流程。
Tech Stack: Go / PostgreSQL / Vue / TypeScript。
Spec: `design.md`、`specs/redeem-email-active-audiences/spec.md`。

- [x] 添加分段校验、真实 PostgreSQL 边界/并集/资格复核/迁移与前端交互测试，先观察失败。
- [x] 扩展 models/repository、253 迁移、前端类型与选项、双语标题及提示。
- [x] 运行相关 Go 测试、隔离数据库集成测试、build/vet、前端测试/typecheck/lint。
- [x] 更新 AGENTS.md 与 FORK_VS_UPSTREAM.md，检查 UTF-8、乱码和最终差异，独立审查。

交付操作：最终验证后仅提交本任务文件，按用户授权 push 并确认远端提交；提交号和远端核验以会话交付记录为准。

Review focus: 7/30 天闭区间下界、自然月与 30 天重叠、零余额互斥、默认范围不扩大、搜索和建批谓词一致。

Execution notes: 当前环境没有 `openspec` CLI；在已有 OpenSpec 根目录按既有格式记录，使用代码与测试验证，不安装工具。
