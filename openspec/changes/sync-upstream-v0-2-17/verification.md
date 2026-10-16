# v0.2.17 同步验证记录

日期：2026-10-16。范围：本地源码同步，不含部署、数据库迁移、真实供应商请求、支付或邮件发送。

## 发布与修改范围

- 官方基线：v0.2.16，01842576b3d5c1e2545b5a99ad555d36e28ed306。
- 官方目标：v0.2.17，060d0160417782cd962a0c7f460f6cd4565d4916；tag object 6d6c4e701a463dcf83c261e8428e93423d6c0b3e。
- Fork 起点：b76424c80，初始工作区干净。上游增量 48 commits、153 files、+7,035/-808。
- 三方预合并：107 个未定制路径采用上游，39 个自动合并、5 个冲突文件人工处理、VERSION 旧值与上游相同、1 个部署示例保留。
- 152 个非部署发布路径交付；版本明确设为 0.2.17（官方 VERSION 仍为 0.2.16）。
- 新文档 docs/OPENAI_GPT55_RETIREMENT.md 被既有 docs/* 忽略，已在 .gitignore 加精确白名单，避免后续提交漏掉发布说明。
- 同步修改 AGENTS.md、FORK_CUSTOMIZATIONS.md、docs/FORK_VS_UPSTREAM.md 及本 OpenSpec；额外调整账号全选测试的异步收尾，不改业务逻辑。

## 红灯与兼容修复

1. 先单独纳入上游 pricing_service_test.go，在旧实现运行 Gemini 4 Argon 与 Claude 5.5 档位计价测试。退出 1：缺 Gemini 价格、Claude 档位误匹配旧价卡。合入上游定价并保留 Fork 别名/历史兜底后，同一组定向测试通过。
2. 新增 TestPendingRefundPreservesResetCardUntilSettlement。初始退出 1：新服务从 pending 审计重建计划后，退款成功但来源卡未作废。审计增加 ResetCardID/ResetCardToVoid 并恢复到计划，作废继续位于成功事务和 deductionHeld 提前返回之前。重跑退出 0；成功/失败、兼容计数和重复终结均覆盖。
3. 新增 TestShouldEmulateWebSearch_AnthropicOAuthRetainsNativeSearch。有效渠道开启时初始退出 1，旧 Fork OAuth 边界在共享代码合入后丢失。wrapper 恢复 Anthropic APIKey/Kiro 直连边界，保留 OpenAI APIKey 新能力；最终 service 全量单测通过。测试准备阶段纠正了清理 nil atomic.Value 和缺少搜索 provider 的测试夹具，然后才观察到真实行为红灯。
4. 前端首次全量 390 文件/2991 断言通过但退出 1：AccountsView.selectAllResults 测试未等待防抖，跨测试触发已重置 mock。补等待第三次加载、flushPromises 和卸载，定向 5 项通过，最终全量退出 0。
5. 上游新增两个 Codex import 测试缺少 Fork Kiro 构造参数；补齐测试参数，不改实际 handler/Wire 签名。重跑 admin 包和 vet 通过。

## 实际命令与结果

工作目录 backend：

- `go test -tags=unit ./internal/service -run 'Test(BillingService_Gemini4ArgonFallbacksAreBillable|DefaultPricingIncludesGemini4ArgonRates|PricingService_AntigravityClaude55TiersUseBaseModelPricing)' -count=1`：旧实现预期失败（退出 1）。
- 定向计价、搜索、退款组合测试：退出 0，3.849s。
- `go test -tags=unit ./internal/service -run '^TestPendingRefundPreservesResetCardUntilSettlement$' -count=1`：先退出 1，修复后退出 0，1.229s。
- `go build ./...`：最终生产代码修改后退出 0。
- `go test -tags=unit ./internal/... -count=1`：53 个包通过（service 194.962s）；仅 admin 包因两个新增测试构造参数缺失编译失败，整体退出 1。
- 修复后 `go test -tags=unit ./internal/handler/admin -count=1`：退出 0，1.535s。其余包生产代码/测试未再改，不重复运行。
- 修复后 `go vet ./...`：退出 0，无输出。

工作目录 frontend：

- `pnpm exec vitest run src/views/admin/__tests__/AccountsView.selectAllResults.spec.ts`：5 项通过，退出 0。
- 最终 `pnpm exec vitest run`：390 文件、2991 测试通过，退出 0，77.06s，无 unhandled errors。
- 最终 `pnpm lint:check && pnpm typecheck`：退出 0。
- `pnpm run build`：退出 0；Vite 大于 500 kB 的 chunk 提示仍存在，不阻断构建。之后仅调整测试异步收尾，无前端生产代码变化。

仓库根目录：

- `git diff --check`：退出 0。
- Python 范围/编码检查：所有变更文本 UTF-8 解码成功，无 BOM、U+FFFD 或合并冲突标记。
- `git diff --name-only HEAD -- backend/migrations deploy`：空；309 个历史 SQL、部署文件未改。
- 240 个相对上游新增的 Fork 路径仅有三份同步文档发生变化，其余独有文件保持原样。重叠共享代码由三方合并、专项测试和只读复核确认。
- 最近邮件样式提交 370228bc7 的全部文件未改；群发提交 b76424c80 仅两份同步文档改动，功能文件不变。代理失败拒绝直连保留在共享转发文件与 proxy 测试中。
- `git diff --cached --name-only` 为空，未暂存、提交或推送。

## 复核

只读审查覆盖生产重叠文件，发现并闭环 OAuth 原生搜索回归；最终复核未发现其他明确的合并引入缺陷。重点确认 Kiro/cache/credits、冻结计价时刻与别名、XorPay、Access Ban、订阅卡事务作废和邮件定制。

## 未运行与上线注意

- 未运行需要真实 PostgreSQL/Redis/供应商的 integration/e2e；单测采用项目现有本地夹具，不能证明线上账号权限、网关支付状态或真实邮件投递。
- 没有安装依赖；Go 使用项目要求的现有 1.27.2 toolchain。无新增依赖或数据库迁移。
- OpenSpec CLI 不可用，未执行其 CLI validate；按仓库既有格式检查了 proposal/spec/design/tasks。
- deploy/config.example.yaml 保持原样；gateway.min_output_tokens 使用应用默认 128，最低 16。
- 新版 pending 退款审计携带“扣减已保留”状态。降级到旧程序前须结清此类退款，避免旧程序重复扣减；本次不执行生产退款或降级。


## 用户要求的第二轮审核

- 发布覆盖与规范两个只读复核完成；152 个非部署路径覆盖、309 个历史 SQL 不变、Fork 独有功能文件保留。
- 再次执行 `go test -tags=unit ./internal/... -count=1`：退出 0，54 个包通过，已包含此前修正的 admin 测试。
- 再次执行 `pnpm exec vitest run`：退出 0，390 文件 / 2991 测试，45.14s；`pnpm lint:check && pnpm typecheck` 退出 0。
- 再次执行 `go build ./... && go vet ./...`：退出 0。
- 复核发现新增 OpenAI 搜索入口复用了旧的宽松代理解析：有 ProxyID 但缺 Proxy 时可能使用供应商代理或直连。共享 `doWebSearch` 改用已有严格 helper，解析失败包装 `websearch.ErrProxyUnavailable`，由响应层返回 502 failover。
- `TestWebSearchEmulationRejectsUnavailableAccountProxy` 在修复前退出 1，覆盖 OpenAI、Anthropic、Kiro；用未初始化 manager 证明校验先于供应商选择，且验证响应层 failover 分类，无真实网络请求。

- 代理修复后执行 `go test -tags=unit ./internal/service -run 'Test(WebSearchEmulationRejectsUnavailableAccountProxy|ShouldEmulateWebSearch|PendingRefundPreservesResetCardUntilSettlement)' -count=1 && go build ./... && go vet ./...`：整体退出 0，定向测试 1.235s。全量后端检查在此最小修复前完成，修复后重跑相关回归与编译/静态检查。

### 本次范围之外的既有风险

- `userRepository.UpdateBalance` 对正数同时增加 `total_recharged`，旧版本退款失败恢复余额也使用此方法，可能使累计充值统计偏高。本次不改变既有资金口径或历史数据。
- 旧 OpenAI 普通转发也有宽松代理解析，不应把此前 Fork 部分转发路径的严格校验解释成全协议覆盖。本次封住新增搜索能力及共享搜索路径，未全仓改造其他出站链路。
- 上述风险不是此次合并新增的回归；本次审核不构成真实支付、生产部署或全协议出站安全验收。
