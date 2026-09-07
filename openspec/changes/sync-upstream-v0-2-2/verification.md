# v0.2.2 同步验证记录

## 固定输入

| 项目 | 值 |
| --- | --- |
| 初始 Fork | main，HEAD `1f38262d1`，工作树干净 |
| 官方旧基线 | v0.2.1 / `578785ee7fb35030b094b69624efe25670a36f5f` |
| 官方目标 tag object | `86495464d82f78da55792db6aa5568ee1302716a` |
| 官方目标 commit | v0.2.2 / `5485f368b29d05adb95a00f71801c7c23d8f48af` |
| release 增量 | 102 commits / 261 files / +13,430 / -2,005；不识别重命名时为 265 路径 |
| 工具链 | Go 1.27.0，Node 22.23.2，Vitest 2.1.9 |

## 合并与定制保留

- 使用 v0.2.1、Fork HEAD、v0.2.2 逐文件三方合并；没有提交、暂存或移动分支。176 路径直接合入、69 路径三方无文本冲突、9 个文本冲突、1 个定制测试删除/重命名冲突、6 个直接删除、3 个已经相同路径；VERSION 单独设为 0.2.2。
- 82 个上游/Fork 重叠路径逐项核对。账号成本的官方峰谷修复已由 Fork 等价实现，保留请求 ctx、冻结 pricingAt 和统一定价管线，新增上游峰谷/标准价回归保留，避免重复计算 max 倍率。
- Gateway 新 rootRoute helper 和 Codex 链继续携带 IPBan；白名单位于鉴权后、Composite 改写前。新增运行时路由回归覆盖 13 个封禁入口和 4 个 Kiro 模型准入入口。
- Wire 保留 Kiro 在 Grok 前、AuthService 的 IPBan，以及新增 config 注入。鉴权快照升级 v24，保留规范化后的 groupForSnapshot 和全部 Kiro 字段。
- Ent 的 Group 字段更名完整保留 Kiro/IPBan。纠正三方文本合并误把 Group ResetField 插入 IPBan ResetField 的位置；最终 mutation 的变化只涉及白名单更名，runtime 字段索引保持原位。
- 支付/管理端受信任兑换入口隔离公共失败计数；支付/兑换的订阅 PurchaseSource、重置卡期限与来源幂等保持原样。XorPay、订阅核心服务、登录自动封禁、提示词审计及其历史迁移未被覆盖。
- 账号创建/编辑在简易模式也显示分组选择，保留 Fork 的 Select、边框样式和 Kiro 直连/中转逻辑。模型列表布局测试重命名时迁移 Kiro 颜色断言；修复新 Codex manifest 测试遗漏 auth store 和候选 API 改名。合并重复中英文键时仅保留一份语义等价文案。
- 481 个不与 release 重叠的 Fork 文件中，478 个逐字节不变；另 3 个仅为本次要求更新的 AGENTS.md、FORK_CUSTOMIZATIONS.md、docs/FORK_VS_UPSTREAM.md。
- 上游新增行核对仅剩已审阅的文本例外：Fork 注入参数/IPBan 链、ctx/pricingAt 签名及格式、规范化快照变量、等价翻译文案和测试 API 改名。没有遗漏未解释的官方逻辑。
- 293 个历史 SQL 迁移逐字节不变；新增 `235_group_model_allowlist.sql` 与官方 tag 逐字节相同。go.mod、go.sum、pnpm-lock.yaml 及依赖版本无变化，package.json 只新增上游 i18n 构建门禁。

## 验证结果

Go 与版本脚本命令在 backend 目录执行；pnpm 和 Git 命令在项目根目录执行。

| 命令 / 检查 | 实际结果 |
| --- | --- |
| Kiro、冷却、支付 provider 基线 | 通过；Access Ban 的行为测试由最终 unit suite 覆盖 |
| 用户重置卡及 Fork i18n 基线 | 2 文件 / 31 tests 通过 |
| `go test ./internal/pkg/apicompat -run FunctionArgumentsDone -count=1`（测试先行） | 旧版两个场景预期失败：最终参数丢失；移植后由最终全量通过确认修复 |
| `go test ./internal/server/routes -count=1` | 通过，包含新增 Access Ban/Kiro 运行时回归 |
| `go test ./... -count=1` | 最终 52 个测试包通过、退出 0；service 125.100 秒 |
| `go test -tags=unit ./... -count=1` | 最终 60 个测试包通过、退出 0；service 180.347 秒 |
| `go vet ./...` | 最终通过，退出 0 |
| `go build ./...` | 最终通过，退出 0 |
| `pnpm --dir frontend run test:run --maxWorkers=4 --minWorkers=2` | 277 文件 / 2008 tests 通过，退出 0，无未处理错误 |
| `pnpm --dir frontend run lint:check` | 通过，退出 0 |
| `pnpm --dir frontend run typecheck` | 通过，退出 0 |
| `pnpm --dir frontend run build --outDir /tmp/sub2api-v022-web` | 通过，i18n 门禁 3 tests 通过；产物在临时目录 |
| `sh scripts/resolve-version.sh` | 输出 0.2.2 |
| `git diff --check` / 暂存后完整检查 | 原有跟踪文件通过；暂存新增文件后发现官方 `instructions_gpt6_astra.txt` 含 11 处行尾空格，已确认与 v0.2.2 逐字节相同并保留原文；排除该文件后完整暂存差异检查通过 |
| UTF-8 / BOM / 替换字符 / 冲突标记 | 修改和新增文本文件检查通过；中文改动已复核 |
| 本机预览 | http://127.0.0.1:3012/home 返回 HTTP 200；浏览器确认首页、品牌图像和中文文本正常渲染 |

## 修正过程与限制

- 首轮编译发现上游新测试未包含 Fork Kiro/IPBan 注入参数、一个 Fork 模型列表测试仍用旧字段、翻译重复和创建账号页遗留 simple-mode 条件；均已按实际接口修正，最终检查通过。
- 首轮 unit 只有新上游静态路由断言漏算 IPBan；适配断言并补运行时覆盖后，最终完整 unit 通过。
- 首轮前端测试发现新 Codex manifest 测试缺少 auth store、以及高并发全量下已有账号选择测试出现一次异步异常。账号选择定向复跑通过；完整复跑限制 2–4 workers 后全部通过，无需修改账号选择业务代码。
- 本机无 Docker、PostgreSQL、Redis 命令，未运行容器集成测试、真实 SQL 迁移、支付/OAuth 或登录后端联调。后端未启动，预览仅验证静态前端；登录后的真实数据工作流需要本地后端。
- 前端构建保留既有 Browserslist 数据陈旧、混合静态/动态导入及大 chunk 警告，未为此扩大升级范围或安装依赖。
- OpenSpec CLI/CodeGraph 在当前会话不可用，采用现有 spec-driven 文档结构与 Git/rg，未运行 OpenSpec CLI validate。
- 独立只读复核未发现可行动的升级回归或明确官方功能遗漏；复核涵盖 Access Ban/Kiro 准入顺序、支付和兑换 PurchaseSource、Ent Group 更名与 IPBan 隔离、冻结计价时刻、鉴权快照 v24、Wire 顺序及账号创建/编辑定制，并核对上游新增行例外清单。复核执行 Git/rg/文件读取与 `git diff --check`，未重复运行完整测试；真实迁移及后端联调仍受上述环境限制。

## 上线边界

本次未执行 git add/commit/push、迁移或部署。部署前应备份数据库并检查已启用的旧 models_list_config：迁移 235 保留内容但升级为实际模型准入白名单；未列出的模型可能被拒绝。数据库列更名后不能直接启动旧二进制作为回退，实际迁移及回滚需另行确认。
