# v0.2.8 同步验证记录

## 固定范围

- 起点 Fork `cf2b9c670`，目标官方 `v0.2.8` commit `fd80b08c90b55edcad5b00171b53f08721d30da1`。
- 官方增量：473 个文件，+28,638/-2,096；新增迁移 `238b`、`239`、`240`。
- 采用官方树 + Fork 定制三方回放；未执行数据库迁移、部署、真实供应商或支付联调。

## 合并检查

- Kiro、XorPay、Access Ban、订阅重置卡、批量用户删除、提示词审计、Ops、VersionBadge、GLM、品牌和支付 UX 保留。
- `157`/`237`/`238` 平台 CHECK 保留 Kiro、MiniMax、OpenCode Go；`159`/`160` 与 `224`/`225` 迁移保留。
- Wire 重新生成；插件目录读取凭据前检查 `StatusActive`，账号元数据会移除代理凭据并递归过滤敏感 Extra 字段。

## 命令结果

- 后端 `go test ./... -run '^$'`：通过。
- 后端 `go test ./internal/handler/admin ./migrations -count=1`：通过。
- 后端 `go test ./internal/service -count=1`：通过（139.457s）。
- 后端 `go vet ./...`、`go build ./...`、`go build -tags embed ./cmd/server`：均通过。
- `sh backend/scripts/resolve-version.sh`：输出 `0.2.8`。
- 前端 `pnpm run test:run`：通过（347 个测试文件、2590 个测试）。
- 前端 `pnpm run typecheck`、`pnpm run lint:check`：均通过。
- 前端 `pnpm run build`：通过；内置 i18n 完整性检查 3 项通过，构建仅有既有大 chunk 警告。
- `git diff --check`（含 staged 与工作树差异）：通过；UTF-8 异常替换字符扫描：通过；未发现未解决合并标记。
- 未执行数据库迁移、部署、真实 PostgreSQL/Redis、供应商 OAuth 或支付网关联调。

## 限制

本地未执行真实 PostgreSQL/Redis/模型 OAuth/支付网关和部署；上游新增 SQL 仅纳入源码，需上线前按项目迁移流程备份、审核和执行。
