# v0.2.7 同步验证记录

## 固定范围

- Fork 起点 `0c99c78fb268f1f5c34be222457bdeccd6f3058e`；初始工作区干净，回退后的树与 v0.2.5 Fork 提交相同。
- 旧官方基线 v0.2.5：`86f93c28ee34cc74b629dafb748bd5ac5ca8c5ea`。
- 目标官方 v0.2.7：tag object `7484192016807acf55c6ef4f2827d371eb61ec1c`；commit `aea725f2ea644d5592d0bbb1d63b607efa7e200a`。
- 精确增量 71 commits / 131 files / +7,283 / -446。直接从 v0.2.5 三方合并；不恢复此前已回退的 v0.2.6，也不带入 v0.2.7 未包含的 Codex 票据模块。
- 官方 VERSION 仍是 0.2.5，Fork 设置为 0.2.7。

## 合并与定制

131 个 release 路径全部核对；27 个重叠路径、6 个文本冲突。103 个路径与官方目标字节一致，其余保留 Fork、版本或测试兼容差异。无受保护路径需要同步。
298 个历史 SQL 与起点逐字节一致，无新增 SQL。558 个非重叠且允许读取的 Fork 文件中，554 个不变；变化仅为三份定制说明和插件源接线修复。

- Kiro OAuth/credits/cache_read/独立 429 冷却/真实上游 ID/精简状态 DTO/用户 Claude 别名、XorPay、Access Ban/admin 登录失败封禁保持。
- 重置卡期限快照、来源幂等、退款事务、消费/旧接口/兼容镜像、提示和 active_available 筛选保持。兑换分页只增加按当前用户查询路径，不改变发卡事务。
- 提示词审计、Ops 删除、GLM 分类、支付金额分离/隐藏倍率、品牌/Select/ConfirmDialog、VersionBadge 禁止在线更新保持。
- Seedance 四个 URL 前缀、创建/查询/删除均使用含 Access Ban 的 rootRoute；账号能力默认仍不启用 Seedance。
- 插件 KV 与账号目录注入同时保留 Kiro/Grok/IPBan。独立审查发现上游只在生成文件手写 SetAccountDirectory，重新生成会丢失；增加 ProvidePluginManager 源 provider，并用现有 Wire 生成器验证可重现。
- gRPC/x/crypto/x/net 等使用官方安全升级，保留 tokenizer、x/image 0.45.0 与 direct x/text；前端依赖不变。Wire 补充其既有 google/subcommands 的两条校验和，无新增业务依赖。

## 回归及诊断

- 合并金额测试先在旧实现运行：5 fail / 3 pass，证明非法字符/超两位小数不能恢复；修复后 8 pass。人民币、快捷金额和 max 限制保留，未把拒绝值保存成恢复目标。
- 插件 provider 回归在缺少 setter 时得到 Unavailable（账号目录不可用），补上注入后通过；同一测试验证未声明能力的插件仍被拒绝。
- 初轮冲突处理脚本对标记行使用过宽匹配，造成依赖/Wire/监控测试截断；Go 解析检查和独立审查发现后，从原始三份重新逐块合并，检查完整差异并重新运行验证。无截断内容留在最终文件。
- 独立只读审查检查重叠路径及 Fork 新增行；唯一接线问题修复后复核无新增发现。

## 已验证的命令

Go 命令在 backend，pnpm 命令在 frontend，其余在仓库根。Go 1.27.0。

- `pnpm exec vitest run src/components/payment/__tests__/AmountInput.spec.ts`：8 项通过，退出 0；此前旧实现 5 项预期失败。
- `pnpm run test:run --maxWorkers=4 --minWorkers=2`：312 文件 / 2346 项通过，退出 0。
- `pnpm run typecheck && pnpm run lint:check && pnpm run build`：通过，退出 0，包含 i18n 校验、vue-tsc 与 Vite 构建。
- `go test ./internal/service -run TestProvidePluginManager_AccountDirectoryCapability -count=1`：修复后通过，退出 0；缺少注入时预期失败。
- `GOPROXY=https://goproxy.cn,direct go generate ./cmd/server`：两处生成指令均成功，退出 0。默认 Go proxy 首次获取工具所需模块元数据超时，改用已有模块代理后成功，未更改用户配置。
- `go vet ./... && go build ./... && go build -tags embed ./cmd/server`：生成后通过，退出 0，包括内嵌前端资源的服务端构建。
- `sh backend/scripts/resolve-version.sh`：输出 0.2.7，退出 0。

- `go test ./... -count=1`：最终 52 个测试包通过，退出 0；service 140.353 秒。

- `go test -tags=unit -p 1 ./... -count=1`：最终 60 个测试包通过，退出 0；service 203.589 秒。
- `git diff --check`、新增文件 no-index whitespace 与变更 Go 文件 `gofmt -l`：无诊断。
- 最终 139 个变更文本文件均为 UTF-8，无 BOM、异常替换字符、合并冲突标记或新增绝对个人路径；OpenSpec 文件与 requirement/scenario 结构检查通过。
- 最终再次确认 HEAD 不变、暂存区为空、298 个历史 SQL 字节不变。

## 限制

- 本机无 Docker/PostgreSQL 可执行文件，未运行 integration 标签的数据库/容器测试；未进行真实迁移或部署。此次没有新增 SQL，298 个历史 SQL 字节不变。
- 未调用真实模型、OAuth、Ark、插件进程或支付网关进行外部联调。Seedance 使用官方文档中的查询结算机制，真实启用前需配置账号能力、模型映射和输出 token 价格；本次不代改业务配置。
- 前端构建有大 chunk 提示，部分测试有既有 i18n runtime 警告；不阻断检查，不扩大到无关依赖或拆包调整。
- OpenSpec CLI 与 CodeGraph 不可用，沿用现有规划结构与 Git/rg；未安装工具，未运行 CLI validate，改以文件/requirement/scenario 检查验证规划完整性。
- 实现阶段未读取或修改秘密配置，未暂存、提交、推送或发布。用户随后明确授权审查完善并提交推送。


## 提交前再次审查

- Standards 审查发现 1 项：插件身份解析缺少 active 状态校验。真实目录回归在 disabled/error 两种状态均复现继续返回测试凭据；现已在读取 token 前拒绝非 active。仍 active 但暂停调度的账号保持可用，不把调度暂停当作撤销凭据。
- Spec 审查确认上游 131 路径、定制保留、298 SQL、兑换用户隔离与新网关鉴权链符合要求。提交清单检查额外发现 `docs/seedance-api.md` 被 Fork 的 `docs/*` 忽略，已加精确例外，文档与官方字节一致。
- 安全修复后 102 个 release 路径与官方字节一致，其余 29 个包含已记录 Fork、版本、测试或安全修复差异。最终交付为 141 个文本文件。
- 本轮前端定向 Vitest：金额输入、用户平台配额、账号批量编辑与账号编辑，4 文件 / 143 项通过，退出 0。前端源码未在前次全量验证后改变。
- 远端核验：fetch 后本地 HEAD 与 origin/main 的 ahead/behind 均为 0。按用户授权只提交本次文件并普通 push 到 main，不推 release tag、不执行部署。
- 修复后 `go test ./internal/service -run 'TestResolvePluginOutboundIdentityRejectsDisabledAccount|TestProvidePluginManager_AccountDirectoryCapability|TestPlugin|TestBuildHostServices' -count=1`：通过，退出 0。
- 修复后 `go test -tags=unit ./internal/handler ./internal/service ./internal/server/routes -run 'TestResolvePluginOutboundIdentityRejectsDisabledAccount|TestSeedance|TestProvidePluginManager_AccountDirectoryCapability|TestPlugin|TestBuildHostServices|TestKiro|Test.*ResetCard|Test.*AdminLogin' -count=1`：3 个包通过，退出 0。
- 修复后 `go vet ./... && go build ./...`：通过，退出 0。Standards 独立复核并重跑新增状态回归通过，当前无待修复发现。
- 最终 UTF-8/BOM/异常替换字符/冲突标记、whitespace、gofmt 与提交路径检查通过；不包含秘密或生成产物。
