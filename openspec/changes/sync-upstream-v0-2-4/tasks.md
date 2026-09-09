# v0.2.4 同步执行计划

**Goal:** 精确同步上游 v0.2.4 并保留全部 Fork 功能。
**Architecture:** 旧 tag/Fork/新 tag 逐文件三方合并，冲突只做兼容所需改动。
**Tech Stack:** Go 1.27、Ent/PostgreSQL/Redis、Vue 3/TypeScript/Vitest。
**Spec:** specs/upstream-release-sync/spec.md；design.md。

- [x] 固定旧/新 tag 与 HEAD，读取 Fork 清单并生成差异/冲突清单。
- [x] 导入 api_key_group_visibility_test.go，运行 `go test -tags=unit ./internal/service -run TestGetUserGroupVisibility -count=1` 观察失败。
- [x] 合入官方增量并逐块解决冲突，保留 Kiro 平台及全部定制。
- [x] 为 237 平台约束保留 Kiro 增加回归断言，观察原版失败并修复。
- [x] 运行 Go default/unit（串行）、vet/build，前端 Vitest/typecheck/lint/build及部署脚本离线测试。
- [x] 独立复核重叠差异与全部非重叠定制、历史迁移；处理实际发现的问题。
- [x] 更新 VERSION/三个 Fork 文档/验证记录；检查最终差异和 UTF-8/BOM/乱码。
