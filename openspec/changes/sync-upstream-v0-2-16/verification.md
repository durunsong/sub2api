# 验证记录

日期：2026-10-13。工作目录 `personal_code/sub2api`。未执行迁移、部署、支付或提交。

## 命令与结果

- `go build ./...`：退出码 0。
- `go vet ./...`：退出码 0。
- `go test -tags=unit ./internal/...`：除 `internal/handler` 外全部通过。`internal/service` 189.516s 通过。handler 仅 `TestUserFacingGroupDTOsAreLocalized` 失败，原因是消费重置卡响应未调用 `Localized`。
- 修复后 `go test -tags=unit -count=1 ./internal/handler/`：退出码 0，39.813s。
- 修复 Select 组字中的 Escape 后，`pnpm exec vitest run`：387 个文件、2977 个测试通过，耗时 38.53s，退出码 0。
- `pnpm typecheck`（`vue-tsc --noEmit`）：退出码 0。

## 边界检查

- 历史 SQL 相对 HEAD 无修改。新增 `242_group_i18n.sql`、`243_gemini_cached_contents.sql`、`244_gemini_cached_contents_channel_usage.sql`、`245_gemini_cached_contents_mutation_lease.sql`。
- 变更文本无 BOM、无 U+FFFD、无冲突标记。
- `backend/cmd/server/VERSION` 为 0.2.16。官方 tag 内该文件仍是 0.2.15。
- 改动未暂存、未提交。
