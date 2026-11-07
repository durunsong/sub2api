# v0.2.20 同步验证

## 来源与边界
- 官方标签：v0.2.19 `dece26153621959ccd7635814f4b2d08304e681a` → v0.2.20 `dab829a4e9b7d1cc307189aea8b0319eb29e29d0`，112 commits、459 files、+29,482/-1,457。
- 初始 Fork HEAD `c53b90760`，工作区干净。期间其他任务提交 `4996405cc` 用户端批量兑换，已增量保留；该任务之后的三份兑换源码/测试格式化和文档格式化不覆盖。
- 六个部署配置路径未同步：deploy/.env.example、deploy/config.example.yaml、四份 docker-compose 文件。其余发布源码/依赖/文档增量合入。MCP 默认关闭，启用需由部署负责人配置公开地址。
- 340 份历史迁移字节一致；仅新增 249_mcp_oauth.sql，不执行 SQL。270 个安全范围内 Fork 独有路径中，功能源码未删除；修改包括普通封禁处理链修复、Kiro 权限回归、同步文档及同期兑换文件的并发格式化。

## 适配与红灯证据
- 上游 Google API 超时测试在旧实现失败：HTTP 408 返回 UNKNOWN，期望 DEADLINE_EXCEEDED。合入后同一测试退出0。
- Kiro 请求解析保留 BodySessionID，并加入字符串复制和重试状态恢复；认证快照保留 Kiro 字段、RequirePrivacySet 与 ImageCountBilling；被动用量加入受限密钥权限检查。
- MCP transport、公开 OAuth/发现和用户授权子路由接入现有 IPBanGuard，新增14入口×5种封禁/故障情形回归。
- OAuth suffix-only 改动原会绕过 Fork 格式与邮箱封禁；6项新增场景在未修实现上均失败（错误应为 INVALID_EMAIL_FORMAT/EMAIL_BANNED/SERVICE_UNAVAILABLE，实际成功）。显式恢复两项检查，保留上游旧账号精确黑名单登录兼容。
- 旧图片计费入口漏了按张开关 guard：两项现有回归期望1/0.8 USD，实际0.018/3.15 USD。依据目标标签恢复条件，保留 Fork opts 和混合请求 token 路径。
- 上游测试构造函数需补 Fork Kiro 参数；旧 Fork 测试需适配新增 recordUsageOpts、buffered terminal Account/返回值。均只更新调用，不删除原断言。

## 前端验证
- `pnpm install --frozen-lockfile --ignore-scripts`：退出0，锁定同步 vue-i18n/message-compiler 11.4.16，未执行安装脚本。
- `pnpm exec vitest run`：406文件/3233测试，首轮404文件/3231测试通过；两项因并发兑换文案和后端目录更新时序失败。恢复文案、应用后端后，相关5文件/27测试复测全通过。
- 最终 `pnpm lint:check`、`pnpm typecheck`、`pnpm build` 均退出0；build包含3项i18n检查、vue-tsc和Vite。仅有旧Browserslist数据库和大chunk提示。
- vue-i18n v11 runtime实际中文字符串插值返回“剩余 2 张卡”，保留现有别名即可使用。

## 后端验证过程
- 初次 `go build ./...` 退出0。
- 首次全量unit发现上述测试签名问题；修复后开始定向验证。为避免多个大型Go测试包同时编译争用内存，首次全量和并发routes/handler重跑主动终止（退出143），不计作通过。
- Kiro/请求缓存/认证快照/被动用量定向回归退出0（1.090s）；受影响旧测试补充回归退出0（1.929s）。最终完整检查另列于下方。

## 检查限制
- 未启动业务服务，未执行PostgreSQL/Redis真实integration/e2e、供应商调用、生产支付、邮件发送或部署。
- OpenSpec CLI未安装；沿用已有spec-driven文档结构，使用本地结构/引用/编码校验。
- 新OAuth迁移只交付源码；部署前需备份、执行迁移并验证授权撤销/轮换。回退二进制不删除新增表，已发授权应先撤销。既有pending退款降级限制继续适用。
- 未暂存、提交、推送；其他任务的并发提交不属于本次操作。

## 最终验证结果

- `go test -p 2 -tags=unit ./... -count=1`：62个测试包通过（service 222.664s），仅 routes 新增 MCP 测试暴露原 Fork 普通 IPBanGuard 未 Abort。没有忽略该失败：新增 IP/UA 下游副作用测试，两项先红灯，再在共享 blocked 分支加 c.Abort。
- 修复后 `go test -p 2 -tags=unit ./internal/server/middleware ./internal/server/routes -count=1`：退出0（0.821s/11.772s）。合并上述结果，全部63个有测试包在最终相关源码状态下通过；其他包源码没有改变，未无意义重跑222秒的service套件。
- MCP漏接封禁的overlay定向测试：`go test -p 2 -overlay=<临时overlay.json> ./internal/server/routes -run '^TestMCPRoutesEnforceAccessBanBeforeProtocolHandlers$/^IP$/^GET_$/^api$/^v1$/^mcp$' -count=1` 退出1，明确捕获预期403却进入认证返回401；正常生产文件的完整routes套件通过。
- OAuth测试在未修实现观察6项预期拒绝均被绕过；修复后 `go test -tags=unit ./internal/service -run '^TestLoginOrRegisterVerifiedEmailOAuth_' -count=1` 的overlay验证退出0（1.268s）。初次绿灯中测试错误假定仓储初始化不创建email identity，改为前后identity数量不变后通过；未弱化拒绝断言。已测文件原样复制后纳入全量unit。
- 最终 `go build -p 2 ./... && go vet -p 2 ./... && go build -p 2 -tags embed ./...`：整体退出0；包含已构建前端的嵌入模式。
- `go test -race -p 2 ./internal/pkg/httputil ./internal/pkg/apicompat -count=1`：退出0（13.637s/2.566s），覆盖资源预算、超时和协议累计缓冲相关回归。
- 独立只读审核已复核全部重点兼容点、封禁修复和预期发布差异，无未解决Critical/Important/Minor问题；并非对459个上游文件全部逐行安全审计。
- 最终源码指纹与全量测试开始时一致，唯一后续变更为 ip_ban_guard.go/test，两包完整复测通过。前端最终验证之后未再改源码。

- 最终静态检查：467份变更/新增文本通过UTF-8、无BOM、无替换字符/乱码/冲突标记检查；340份历史SQL字节不变；版本0.2.20、OpenSpec文档结构、六个部署配置不变、空Git索引及 `git diff --check` 均通过。

## 用户授权提交与推送前复核（2026-11-07）

- 用户明确要求 push。本地 main 与 origin/main 同为 4996405cc，fetch 后无分叉。
- 445份既有验证源码指纹无变化；后续普通封禁两文件差异与已验证修复一致。
- 新鲜关键回归：service 的 OAuth、图片按张计费、Kiro认证快照/请求复用，以及 middleware/routes 的封禁与MCP入口测试全部退出0（1.444s/0.975s/1.539s）。命令为 `go test -p 2 -tags=unit ./internal/service ./internal/server/middleware ./internal/server/routes -run 'TestLoginOrRegisterVerifiedEmailOAuth_|TestCalculateRecordUsageCost_ImageCountBilling|TestAPIKeyAuthSnapshotKiroPolicyRoundtrip|TestGatewayRequestBodyReuse|TestIPBanGuardStopsBlockedRequestsBeforeHandler|TestMCPRoutesEnforceAccessBanBeforeProtocolHandlers' -count=1`。
- 提交范围为469个升级相关路径；三份同期兑换文件的已有纯格式化改动保留未暂存。作用域与UTF-8/BOM/替换字符检查通过，不包含秘密配置或构建产物。
