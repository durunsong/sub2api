# v0.2.18 同步验证记录

日期：2026-10-23。范围：本地源码升级，不含部署、真实迁移、供应商调用、支付或邮件发送。

## 发布范围

- 基线 v0.2.17：060d0160417782cd962a0c7f460f6cd4565d4916。
- 目标 v0.2.18：ac1b8f1271be0d674fa6f074403c1410f30a3a81；tag object de86961b627e48db75d61fff8c05535d44a00cbe。
- Fork 起点：49f13ea62，初始工作区干净。45 commits / 114 files / +7482 / -566。
- 78 路径采用官方版本、30 自动三方合并、4 处冲突人工兼容、2 路径已含目标变化。所有发布路径均交付；VERSION 显式更新为 0.2.18。
- 冲突处理：account.go 保留 Fork fmt，删除上游已不用的 fnv；Responses 保留 Kiro credits 提取并采用新版流式观察/收尾；创建和编辑账号页合并两侧 imports。
- 全部 311 份历史 SQL 保持字节不变；新 Gemini 生图迁移与 Fork 邮件迁移按完整文件名共存，不执行。deploy 路径未修改。
- 独立只读审查确认：未重叠发布路径与官方相同；共享文件的 Fork 增删内容除 VERSION 和已被上游吸收的 PR #8139 外保留。后续仅增加测试适配及 credits 回归。

## 测试发现与修正

1. 先单独合入上游 account_quota_schedulable_test.go，在旧实现运行 TestAccountIsSchedulable_QuotaExceeded，退出 1：service_account 日/总配额耗尽仍可调度。合入实现后同组和 Kiro/Vertex 专项测试退出 0。
2. 上游新增两处 Handler 测试未提供 Fork Kiro provider/cooldown 参数，编译失败；只补测试参数，不改构造函数或 Wire。Handler 全包复跑通过。
3. 既有 API 契约期望未包含近期 nianzs Kiro 增加的 cache mode/create/read ratio、gateway_kiro_tokenizer；核对既有 DTO getter 与 setting_parse 后补齐，未改实际接口。server 全包复跑通过。
4. 既有 Kiro 图片测试假定 noisy PNG 比 flat PNG 大，当前编码结果 5962 < 6662。改为只要求编码长度不同，继续断言两张同尺寸图均为 350 tokens；不改变图片计量实现。kiro 全包复跑通过。
5. 新增客户端断开后保留 Kiro credits 的 Responses 回归；用 Go overlay 模拟漏合 credits 钩子，不修改工作区业务源码。该 overlay 下测试退出 1，精确捕获 credits 从 0.17 丢为 0。

## 已完成命令

backend：
- `go test -tags=unit ./internal/service -run '^TestAccountIsSchedulable_QuotaExceeded$' -count=1`：旧实现预期失败，退出 1。
- Vertex/Kiro/兼容流断连定向 service 测试：退出 0，1.882s。
- `go build ./... && go vet ./...`：退出 0。
- 首次 `go test -tags=unit ./... -count=1`：退出 1，仅上述 handler 编译、kiro 图片前提和 server 期望失败；service 全包通过，230.955s。
- `go test -tags=unit ./internal/handler -count=1`：退出 0，50.701s。
- `go test -tags=unit ./internal/pkg/kiro ./internal/server -count=1`：退出 0，分别 3.370s / 2.772s。

frontend：
- `pnpm install --frozen-lockfile --ignore-scripts`：退出 0；按锁文件同步已有依赖，禁用安装脚本。
- `pnpm exec vitest run`：退出 0，392 文件 / 3004 测试，48.94s。
- `pnpm lint:check && pnpm typecheck && pnpm run build`：退出 0；构建有既有的大 chunk 提示，无阻断错误。
- `pnpm list dompurify --depth 0`：确认 3.4.17；锁文件 follow-redirects 为 1.16.1。

## 验证边界

- Go 使用 1.27.3；首次按 go.mod 自动下载工具链。不新增业务依赖类别。
- 未运行 PostgreSQL/Redis/真实供应商 integration/e2e，单测不能证明线上配置或外部服务状态。
- 没有 OpenSpec CLI，未运行 CLI validate；沿用现有 spec-driven 格式检查文档。
- 生图迁移会开启旧 Gemini/Antigravity 分组开关；部署前备份并核验。二进制降级不会撤销该数据更新。
- 原 v0.2.17 pending 退款降级限制仍适用；本次不改退款/重置卡资金口径。
- 没有暂存、提交、推送、部署或真实数据库操作。

## 最终验证（所有编辑后）

- `go test -tags=unit ./... -count=1`：退出 0，全部 61 个有测试包通过；217.136s（service）。
- `go test -tags=unit ./internal/service -run '^TestHandleResponsesStreamingResponse_ClientDisconnectPreservesKiroCredits$' -count=1 && go build ./... && go vet ./...`：整体退出 0，credits 回归 4.247s。此前 overlay 红灯精确证明漏合 credits 钩子会失败。
- 前端生产代码在 392 文件 / 3004 测试及 lint/typecheck/build 通过后未再修改。
- 四处测试适配经独立只读补审，无发现；没有删除断言或调整实际业务规则来掩盖测试失败。图片测试保留不同编码长度、相同尺寸均为 350 tokens 的行为约束。
- UTF-8/BOM/替换字符/冲突标记检查覆盖 121 个变更文本文件，均通过；`git diff --check` 通过。三份同步文档及五份 OpenSpec 文档引用存在。
- 263 个相对旧标签独有 Fork 路径中，只有三份同步文档及 Kiro 图片测试发生变化，独有功能源码不变。新增上游测试参数适配和 API 契约补全均为测试修改。
- Git 索引为空；311 份历史 SQL、部署配置未变；没有提交、推送或执行迁移。

## 用户授权推送前复核（2026-10-23）

- 用户要求先审核再 push；本轮以 49f13ea62 为基线，按项目规范和升级规格分别独立只读审核，两个维度均未发现阻断问题。复核范围包括发布完整性、Kiro credits/缓存/代理、账号配额、迁移共存及资金相关定制保留。
- `git fetch origin main` 成功；`git rev-list --left-right --count origin/main...HEAD` 为 0/0。
- 提交范围为 123 个已核对文件，仅含发布增量、必要测试适配与同步文档。除本次审核记录外，文件哈希与审核开始时一致。
- 新鲜验证：Vertex 配额、断连 Kiro credits、pending 退款重置卡、搜索代理保护定向测试退出 0（1.292s），随后 `go build ./... && go vet ./...` 退出 0；`pnpm lint:check && pnpm typecheck` 退出 0。
- 同一代码状态此前全量后端 61 包、前端 392 文件/3004 项及前端构建通过，本轮未重复无变更的全量套件。
- `git diff --check`、UTF-8/BOM/替换字符/冲突标记检查均通过；311 份历史 SQL、部署配置未变。真实迁移与生产联调未执行。
