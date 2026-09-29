# v0.2.9 同步验证记录

## 范围与定制保留

- Fork 起点：`f9385c45a`；官方 v0.2.8 `fd80b08c90b55edcad5b00171b53f08721d30da1` → v0.2.9 `4c00df2e0183e2c70b7fa8ba45914205e36aad0c`。
- 官方增量 70 commits / 117 files / +3,133 / -378；114 个源码/测试路径核对同步，3 个部署配置路径排除。
- 92 个非重叠路径逐字节等于官方目标；22 个重叠路径保留 Fork 差异或 VERSION 修正。账号成本实现/测试两个冲突文件手工合并，其他路径三方合并。
- 301 个历史 SQL 保持原样，没有新增迁移；577 个非重叠、受版本管理且允许检查的 Fork 文件保持原样。此前已从 Git 删除的截图不纳入受版本管理文件检查，也不触碰本地忽略文件。
- Kiro/XorPay/Access Ban/提示词审计/订阅重置卡/active_available/Ops/批量删用户/品牌别名/GLM 分类/支付 UX/VersionBadge 均保留。Wire、平台 CHECK 和重置卡历史迁移没有变化。
- 额外修改一个已有测试文件 `backend/internal/service/channel_monitor_quota_mode_test.go`，去除两项参数校验对真实 DNS 的依赖；未修改生产 SSRF 校验。
- 两轮独立只读审查覆盖合并冲突、定制保留及上述测试修正。

## 行为测试证据

- `go test ./internal/pkg/antigravity -run TestBuildToolsPreservesStringConst -count=1`：先导入上游测试，旧实现的 existing_enum/conflicting_enum 两项按预期失败；合入实现后通过。
- 首轮 `go test -tags=unit ./... -count=1`：service 包两项渠道监控参数测试失败，其他包通过。定向重跑可复现。测试域名 DNS 解析包含 IPv6 ULA，先命中既有 endpoint 私网拦截，未进入目标缺字段校验。
- 测试输入改为与既有 endpoint 测试一致的公网 IP 字面量，不进行 HTTP 连接，不更改错误断言。
- `go test -tags=unit ./internal/service -run '^(TestValidateCreateParams_CheckModeMatrix|TestValidateMonitorEndpoint_BasePath)$' -count=1`：通过，参数校验及私网拒绝仍生效。

## 已完成命令

以下命令在 backend 或 frontend 对应项目根目录执行，均读取实际退出状态。

- 后端 `go test ./... -run '^$'`：通过，所有默认标签测试可编译。
- 后端 `go test ./... -count=1`：通过，53 个有测试的包通过，service 用时 180.156s。
- 后端 `go build ./...`、`go vet ./...`：通过。
- 后端 `golangci-lint run ./...`：通过，`0 issues`。
- 后端 `go build -tags embed -o <临时输出文件> ./cmd/server`：通过，包含本次构建的前端资源；未启动服务。
- 根目录 `sh backend/scripts/resolve-version.sh`：输出 `0.2.9`。
- 前端 `pnpm run test:run`：348 个文件、2,606 项测试全部通过。
- 前端 `pnpm run typecheck`、`pnpm run lint:check`：通过。
- 前端 `pnpm run build`：通过，内置 i18n 检查 3 项通过；Vite 提示既有的大 chunk 警告。

## 最终检查

- 修正测试后 `go test -tags=unit ./... -count=1` 全量重跑通过：61 个有测试的包通过，service 用时 193.987s。
- `go vet -tags=unit ./internal/service` 通过；额外测试文件 `gofmt -l` 无输出。
- `git diff --check` 通过；最终 123 个修改/新增文本文件 UTF-8 解码、BOM、异常替换字符与冲突标记检查通过。
- 最终范围仅包含 114 个上游路径、1 个测试稳定性修正、3 个 Fork 文档和 5 个 OpenSpec 文档；没有受保护路径或暂存区改动。
- 独立审查未发现升级引入的明确问题，测试修正未弱化断言或 SSRF 防护。

## 限制

- 三个 `deploy/docker-compose{,.dev,.local}.yml` 未同步，官方 Redis command 的 exec 列表调整需在部署配置维护时单独审核；未读取或修改部署配置/秘密。
- 没有执行真实 PostgreSQL/Redis 集成、供应商 OAuth/模型调用、支付网关联调、数据库迁移或部署。
- OpenSpec CLI 未安装，使用仓库已有 spec-driven 结构并检查文件、引用和内容，没有宣称 CLI validate 成功。
- 所有变更留在工作区，未暂存、提交、推送或创建 PR。
