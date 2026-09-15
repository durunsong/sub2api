# v0.2.5 同步验证记录

## 固定范围

- 起点：Fork `main` / `107b1a256b23b206751503ae0a1928a51a9bbdb9`，初始工作树干净。
- 上游旧基线 v0.2.4：`5de5e2bed035d43591a2e10e51f420ef6a84eb98`。
- 目标 v0.2.5：tag object `4af0e80db1b0bc7626dfb8fb76ccaffc6bb0dc17`，commit `86f93c28ee34cc74b629dafb748bd5ac5ca8c5ea`。
- 精确增量：197 commits / 447 files / +23,038 / -1,538；未纳入目标 tag 后的 main。
- 官方 VERSION 仍为 0.2.4，Fork 设置为 0.2.5，版本解析脚本返回 0.2.5。

## 合并与定制兼容

临时目录逐文件三方预演后写入工作树；127 个重叠路径，30 个文本冲突文件。未移动 HEAD、未暂存。
允许同步的 441 个 release 路径全部纳入；315 个与官方目标逐字节一致，126 个包含版本、Fork 或测试兼容差异。456 个非重叠且允许读取的 Fork 路径中，453 个逐字节不变，其余只有三份版本/定制说明文档。296 个历史 SQL 均与初始 HEAD 逐字节一致。

- Kiro 与 OpenCode 同时保留于共享平台、配额、Composite、渠道、调度快照（11 个平台）；Wire 保留 Kiro/Grok 顺序、Access Ban 和新增 Ollama 限流重置依赖。
- `238_opencode_go_platform.sql` 的 quota/Composite CHECK 保留 Kiro、MiniMax、OpenCode，监控 provider 沿用官方范围。无限额清理不改任何非 NULL 限额（含零）的行；真实 SQL 未执行。
- 订阅延期按上游使用事务行锁，并保留 Fork 月窗口对齐。固定期限重复支付/兑换/管理员分配仍发期限快照卡；到期重开、来源幂等、退款事务、旧接口兼容、USD/token 重置和 active_available 保留。
- 批量删用户整合为一个入口，复用 Fork 批量接口及管理员保护；保留确认目标快照、提交锁和未删除项选择。
- 用户端 Kiro→Claude、GLM 套餐筛选、支付金额分离/隐藏倍率、ConfirmDialog、VersionBadge 禁止在线更新及品牌保留。新 API Key 服务商筛选将 Kiro 归 Claude；批量编辑分组支持 Claude 搜索并传原内部 ID。
- 认证仅明确刷新 401 才清会话，临时失败返回实际刷新状态或 0，防止 auth store 将临时故障误判为登出。
- 两位独立只读审查覆盖前后端重叠文件与关键调用链；发现的问题均已修复并复核。

## 测试驱动与失败诊断

- 先导入官方 `request_ua_validation_test.go`，旧实现对六种非法控制字符 UA 预期失败；合入官方校验后通过。
- 新增 OpenCode 迁移的 Kiro 约束断言对原官方 SQL 失败，补回后通过。
- Kiro 的服务商归类测试先得到 other，补入 anthropic 后通过。
- 订阅锁读测试发现合并仍用真实时间而未采纳上游冻结时钟，改用 `s.now()` 后通过。
- 配额 UI 新测试硬编码 5 个平台，按 Fork 实际含 Kiro 的 6 平台 fixture 校正，保留三窗口禁用行为断言。
- 前端首轮全量仅认证文件失败（6 项）：合并真实刷新状态并保留 Fork 清会话边界后，定向 40 项与最终全量通过。真实 Select 回归验证搜索 Claude 找到 Kiro，并向批量 API 发送原分组 ID。
- unit 第二轮唯一失败为未修改的 `TestServerTimingConnectorRecordsDriverCallsWithoutRowLifetime`：真实计时出现 app=30.8ms、db=63.0ms；测试以多个 2ms sleep 的 DB 累计耗时与 30ms 应用 sleep 比较，受调度负载影响。该测试和实现对 HEAD 无差异，单独 `-count=3` 通过。保留原断言，`go test -tags=unit -p 1 ./... -count=1` 完整复核通过。
- 两个新增上游测试分别遗漏 Gateway/Auth 的 Fork Access Ban 构造参数；仅适配测试接线，不移除生产依赖。

## 已执行验证

Go 在 backend，pnpm 在 frontend，其余在仓库根执行。Go 1.27.0，pnpm 9.15.9，Vitest 2.1.9。

| 命令 | 结果 |
| --- | --- |
| `go test ./internal/pkg/openai -run TestPairCodexClientIdentity -count=1` | 旧实现预期失败；合入后定向通过 |
| `go test ./migrations -run TestOpenCodeGoPlatformMigration -count=1` | 原官方 SQL 的 Kiro 断言预期失败；兼容后通过 |
| `go test ./migrations ./internal/pkg/openai ./internal/service -run 'TestOpenCodeGoPlatformMigration\|TestPairCodexClientIdentity\|TestExtendSubscriptionUsesLockedCurrentRow\|TestAdjustSubscription' -count=1` | 三个包通过；月窗口对齐测试另由 unit 全量执行 |
| `pnpm exec vitest run src/api/__tests__/client.spec.ts src/components/keys/__tests__/BulkEditKeysModal.spec.ts` | 2 文件 / 40 项通过 |
| `pnpm run test:run --maxWorkers=4 --minWorkers=2` | 最终 300 文件 / 2279 项通过，退出 0 |
| `go test -tags=unit -p 1 ./... -count=1` | 最终 60 个测试包通过，退出 0；service 176.520s |
| `go test ./... -count=1` | 最终 52 个测试包通过，退出 0；service 144.616 秒 |
| `pnpm run typecheck && pnpm run lint:check && pnpm run build` | 最终全部通过，退出 0；i18n、vue-tsc 和 Vite 构建通过，存在大 chunk 提示 |
| `go test -tags=unit ./internal/repository -run TestServerTimingConnectorRecordsDriverCallsWithoutRowLifetime -count=3` | 连续 3 次通过，退出 0 |
| `go test -tags=unit ./internal/server/middleware -run TestJWTAuth_UserLookupErrors -count=1` | 测试构造参数适配后通过，退出 0 |
| `go vet ./... && go build ./...` | 通过，退出 0 |
| `bash deploy/tests/apple-container-test.sh` | 通过，退出 0，仅使用 fixture 和临时目录，无真实容器操作 |
| `sh backend/scripts/resolve-version.sh` | 0.2.5，退出 0 |

最终 unit 使用包并发 1 完整运行成功；未跳过测试、未修改计时断言。

## 文件检查与限制

最终改动为 449 个文本文件，UTF-8，无 BOM、异常替换字符或冲突标记。`git diff --check`、未跟踪文件逐个 no-index whitespace 检查和改动 Go 文件 `gofmt -l` 均无诊断；HEAD 未移动，暂存区为空。最终文档落盘后已再次复核，无异常。

- 六个部署配置路径未读未改：`deploy/.env.example`、`deploy/config.example.yaml`、`deploy/docker-compose.dev.yml`、`deploy/docker-compose.local.yml`、`deploy/docker-compose.standalone.yml`、`deploy/docker-compose.yml`。应用默认值与功能代码已同步，既有部署配置需上线前单独核对；不宣称整树与官方相同。
- 本机没有 Docker、psql/postgres、Apple container，未执行 PostgreSQL integration、真实迁移或部署。新增清理迁移会删除三档限额全 NULL 的历史配额行，上线前先备份并验证 Kiro 数据、CHECK、迁移耗时及回滚策略。数据删除不能只靠二进制降级恢复。
- 前端构建存在大 chunk 提示和 Browserslist 数据陈旧提示，未阻断构建；不扩大到无关拆包或依赖更新。
- 未调用真实 OAuth、模型供应商、支付网关联调；自动化测试不能替代这些集成验证。
- OpenSpec CLI 与 CodeGraph 不可用，使用项目已有 spec-driven 结构及 Git/rg，未安装依赖或执行 CLI validate；已检查规划文件存在及 requirement/scenario 结构。
- 未执行 git add/commit/push、生产操作或修改秘密配置。
