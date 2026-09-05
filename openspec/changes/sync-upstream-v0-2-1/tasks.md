## 1. 基线与保护范围

- [x] 1.1 读取规则、定制文档、Git 状态并固定官方 SHA、82 commits、297 files 和 87 个重叠文件。
- [x] 1.2 将已有 origin/main 定制提交作为未暂存增量保留。
- [x] 1.3 运行关键 Fork 基线，并先用上游测试确认缺失行为。

## 2. 实施同步

- [x] 2.1 合入上游测试、schema/Ent、四个迁移与共享契约。
- [x] 2.2 合入网关、Codex/Astra、定价、图片、Ops、支付与前端增量。
- [x] 2.3 逐项解决三方冲突，审阅 87 个重叠文件并验证 Fork 红线。

## 3. 交付验证

- [x] 3.1 更新 VERSION、AGENTS.md、FORK_CUSTOMIZATIONS.md 和差异文档。
- [x] 3.2 运行 Go test、unit、vet/build 及前端 test/lint/typecheck/build。
- [x] 3.3 检查增量完整性、历史迁移、冲突残留、UTF-8/BOM/乱码和最终 Git diff。
- [x] 3.4 记录实际验证结果、限制与剩余风险。
