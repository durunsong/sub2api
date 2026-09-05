# v0.2.1 同步验证记录

## 固定输入

| 输入 | 值 |
| --- | --- |
| 初始工作区 | main，HEAD `3f050cea2`，工作树干净 |
| Fork 保护基线 | `e316f5189`，已作为未暂存增量纳入，未移动 HEAD |
| 官方 v0.2.0 | `aa236488351eb71e120fc2b6fb32e36b0374c918` |
| 官方 v0.2.1 tag object | `adc26f68f687685e847bfb997559f48e79cac475` |
| 官方 v0.2.1 commit | `578785ee7fb35030b094b69624efe25670a36f5f` |
| 官方增量 | 82 commits / 297 files / +12,622 / -1,033 |
| Fork 重叠 | 87 文件，21 文件存在文本冲突 |
| 工具链 | Go 1.27.0，Node 22.23.2，Vitest 2.1.9 |

## 合并处理

- 266 个文件直接应用 tag 增量，2 个先行测试已应用，7 个文件三方无冲突合并，21 个冲突逐项处理，VERSION 单独设置为 0.2.1。
- Ent 先按 Fork schema 合并，再使用既有 Ent 库重新生成，保留 Kiro 字段位置和 IPBan mutation。生成结果仅涉及本次 Group 契约的 7 个生成源码文件。
- 账号统计保留 Fork `ctx/pricingAt` 和统一管线，将最终 reasoning effort 传入，仅乘一次 max 倍率。保留 DeepSeek 峰谷测试。
- 用量 SQL 在所有插入、批量插入及扫描路径并存 `kiro_credits` 与 `upstream_request_id`；两条静态 INSERT 均为 63 列及 `$1..$63`。
- API Key auth snapshot 升为 23，模型目录取自 Fork 已规范化的 group snapshot，并继续包含 Kiro 字段。
- Kiro 冷却恢复和渠道模型限制检查并存；Ops 保留 Kiro 模型诊断与错误删除，新增代理归因和队列边界。
- 精简账号 DTO 补齐 Kiro 六状态字段，保持凭据脱敏；真实上游响应头快照贯穿 Kiro Messages 和两种协议桥接。
- 所有官方新增前端行都存在，唯一文本例外是合并 Fork 测试 imports。后端逐行核对的文本例外均对应 Ent 字段偏移、SQL 增加 Kiro 列、冻结计价时刻、快照变量名及去重倍率。
- 464 个未与本次上游重叠的 Fork 独有/修改文件，与 `origin/main` 逐字节一致。额外改动仅为版本/文档、Kiro 请求 ID 接入及相应测试适配。
- 289 个历史 SQL 迁移逐字节不变，四个新增迁移与目标 tag 逐字节一致。

## 测试与构建

Go 和版本脚本命令在 `backend/` 目录执行；带 `--dir frontend` 的 pnpm 命令及 Git 检查在项目根目录执行。

| 命令/检查 | 结果 |
| --- | --- |
| Kiro/冷却/支付 provider 基线 Go tests | 通过 |
| Fork 订阅及 i18n 基线 Vitest | 3 文件 / 34 tests 通过 |
| Astra 模型测试先行 | 预期 RED：旧模型目录缺少 Astra；合入后 GREEN |
| ultrafast 标签测试先行 | 预期 RED：旧标签未翻译；合入后 GREEN |
| Kiro 精简 DTO 回归 | 预期 RED：缺少六字段；补齐后 GREEN，并验证凭据仍脱敏 |
| Kiro 请求 ID 回归 | 6 场景预期 RED；非流式/流式及三种响应头补齐后 GREEN |
| `go test ./... -count=1` | 首轮完整同步通过；后续 Kiro 补齐由最终 unit 全量覆盖 |
| `go test -tags=unit ./... -count=1` | 最终全量通过：59 个测试包，0 失败，退出 0；service 包耗时 170.531 秒 |
| `go vet ./...` | 通过 |
| `go build ./...` | 通过 |
| `pnpm --dir frontend run test:run` | 269 文件 / 1969 tests 通过 |
| `pnpm --dir frontend run lint:check` | 通过 |
| `pnpm --dir frontend run typecheck` | 通过 |
| `pnpm --dir frontend run build --outDir /tmp/sub2api-v021-web` | 通过，产物隔离在临时目录 |
| `sh scripts/resolve-version.sh` | 输出 0.2.1 |
| `git diff --check` | 通过 |
| UTF-8 / BOM / U+FFFD / 冲突标记 | 329 个文本文件检查通过，无 BOM、替换字符或冲突标记；中文修改内容复核无乱码 |
| 补丁残留 | 无 `.rej` / `.orig` 文件 |
| 依赖与暂存区检查 | go.mod/go.sum/package.json/pnpm-lock.yaml 无变化，暂存区为空 |

## 环境限制及恢复

- 原生 Ent CLI 的 readonly 执行因 CLI 专用依赖缺失而失败；改用仓库已有 `entc.Generate` API 和相同 features/idtype 成功生成，没有安装新依赖或修改锁文件。
- 版本脚本当前没有执行权限，直接执行返回 126；使用 `sh` 运行成功，不修改无关文件模式。
- 首轮 unit 全量发现旧 Fork 测试固定 62 列，已按实际 63 列更新，并增加正数 Kiro credits、非空上游 ID 和 session ID 同行参数断言。
- 本机没有 Docker，未执行 PostgreSQL/Redis 容器集成测试、真实迁移、支付网关或外部 OAuth 联调。已有默认/unit、SQL shape 和迁移测试提供本地证据。
- 前端构建存在 Browserslist 数据陈旧、混合静态/动态导入及大 chunk 警告，不影响构建退出状态。
- 前端预览使用独立本机端口 3011，返回 HTTP 200。未启动后端或连接生产数据，登录及业务操作需配置本地后端。
- 未执行 git add/commit/push、生产迁移或部署。后续生产上线需先备份并按既有迁移流程验证四个新迁移，特别是非事务并发索引。
