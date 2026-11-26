# Fork 相对官方仓库差异说明

> **上游官方仓库**：[Wei-Shaw/sub2api](https://github.com/Wei-Shaw/sub2api)
> **本 Fork 远程**：`origin` → `durunsong/sub2api`（中转/部署用）
> **对比基准**：官方 tag **`v0.2.25`**（同步于 2026-11-24；官方无 v0.1.174 tag）
> **本 Fork 当前版本**：`backend/cmd/server/VERSION` = **0.2.25**
> **最新同步统计**：`v0.2.24..v0.2.25` 共 15 个提交、91 个路径；目标 `348997267d756eab5729bb6b78a275a15d18796a`。请求头身份变量及错误详情、上游分组名称/备注手动与默认关闭的自动同步、billing include=group、OpenAI 透传错误规则/结构化 400 与 compact 过载重试修复。
> **当前工作区**：已合入官方 v0.2.25，并保留 Kiro / XorPay / Access Ban / 提示词审计 / 订阅重置卡 / Ops / UI 品牌等全部定制，含登录失败自动封禁及重置卡提示
> **维护**：新增 Fork 定制后，请同步更新本文与根目录 `AGENTS.md` 摘要。

v0.2.21 整合时另补齐目标版本的缓存 TTL 传递缺口：WS v2 逐轮/累计结果与 Messages/Chat typed usage 共同保留 30m/12h 分档，沿用上游严格校验和无明细通用价回退，不新增计价规则；Kiro credits 独立保留。实际 handler/WS adapter 到扣费、账号统计均有回归，详见 [v0.2.21 验证记录](../openspec/changes/sync-upstream-v0-2-21/verification.md)。

---

## 1. 仓库关系与阅读顺序

| 角色                                   | 说明                                                          |
| -------------------------------------- | ------------------------------------------------------------- |
| **Wei-Shaw/sub2api**                   | 官方上游，功能与版本以该仓库 Releases / tag 为准              |
| **durunsong/sub2api**                  | 本 Fork 的 `origin`，在官方之上叠加 Kiro / XorPay / UI 等定制 |
| **本文档**                             | AI 与开发者改代码前的**完整差异清单**                         |
| **`AGENTS.md`**（项目根）              | AI 协作**强制摘要**与禁区，改代码前必读                       |
| **`FORK_CUSTOMIZATIONS.md`**（项目根） | 历史清单，已收敛到本文；保留作快捷索引                        |

**注意**：本 Fork 已同步官方至 **v0.2.25**，但 **`upstream/main` 仍可能领先**。与官方同步时以目标 release **tag** 为准，不带入 tag 后的 `main` 内容；merge `main` 前务必先读本文 Fork 定制章节，禁止 blindly 采用上游覆盖 Kiro / XorPay / Access Ban 等模块。

---

## 2. 差异总览（相对 v0.1.146）

- **上游 PR #8139 定点合入（2026-10-21）**：完整应用 `e068c3d02ff8664c1142d5f7c51a05a82f75ad3d` 的 3 文件增量（+15/-1），将 `response.queued` / `response.metadata` 纳入共享元数据分类，保留实际输出前的过载重试与切号，覆盖普通 SSE、透传及 WS→HTTP 回归测试。未知事件与实际输出后禁止重放的边界保持不变；未引入其余上游提交，不调整重试预算、调度策略、版本或迁移。

| #   | 模块                           | 类型      | 一句话                                                                                               |
| --- | ------------------------------ | --------- | ---------------------------------------------------------------------------------------------------- |
| 1   | **Kiro 平台**                  | 新增      | AWS Kiro 全链路：OAuth、翻译、缓存模拟、粘性会话、计费 credits、429 冷却                             |
| 2   | **XorPay 支付**                | 新增      | 支付宝及微信 Native 扫码聚合支付，含 webhook、管理端配置、用户端 QR 流程                                           |
| 3   | **Grok + Kiro 配额口径**       | 修改      | 平台常量、默认配额、`157` 迁移 CHECK 约束同时含 `kiro` + `grok`                                      |
| 4   | **UI / 品牌**                  | 修改      | 首页/登录注册靛蓝紫罗兰主题、`logo.svg`、Kiro 紫色平台图标                                           |
| 5   | **组件规范化**                 | 修改      | 原生 `<select>` → `Select.vue`；原生 `confirm/alert` → `ConfirmDialog`                               |
| 6   | **支付 UX**                    | 修改      | 隐藏用户端倍率；XorPay 前端展示为「XorPay（支付宝）」；二维码金额展示优化                            |
| 7   | **版本弹窗**                   | 修改      | 管理员版本下拉**隐藏「立即更新」**，保留版本提示与更新日志链接                                       |
| 8   | **Ops 错误日志**               | 新增      | 管理员按条件批量删除 ops error logs                                                                  |
| 9   | **全局访问封禁（Access Ban）** | 新增      | v0.1.146 之后 Fork 自研：IP/CIDR 封禁 + 中间件拦截注册/API；扩展 UA / IP+UA / 邮箱后缀               |
| 10  | **代理广告**                   | 删除/禁用 | 移除 `ProxyAdBanner` 相关展示                                                                        |
| 11  | **构建与部署**                 | 修改      | Docker 使用 npmmirror、Kiro 环境变量、`.dockerignore` 文档打包调整                                   |
| 12  | **资源与文档**                 | 新增      | Kiro 截图、微信群二维码；`docs/FORK_VS_UPSTREAM.md`、`AGENTS.md`；`docs/legal/` 合规文案             |
| 13  | **管理端批量删用户**           | 新增      | `POST /api/v1/admin/users/batch-delete`，UsersView 多选批量删除（跳过 admin）                        |
| 14  | **订阅重置卡**                 | 新增      | 固定期限支付/正数订阅兑换重复购买改发永久重置卡；按购买时 `validity_days` 快照消费并从点击时重开周期 |

**以下能力已在官方 v0.1.142+ 中，本 Fork 通过同步拥有，不算 Fork 独有开发**（合并时保留了 Kiro/XorPay 定制）：

- **v0.2.23**：GPT-6.2 Sol/Astra 模型与套餐档位、渠道 Ultra Fast 倍率、缓存写入与读取分桶，以及 OpenAI 首段输出前的空闲切换预算。
- **v0.2.22**：Ultra Fast 分组策略、Vertex us/eu 多区域与 Gemini Carbon、透传模型白名单、按用户合并扣费、调度快照与 JSON 热路径优化，以及协议、配额、退款和统计修复。官方目标 `b0fa9f3e6eeab1968a53e1289b69daa4bb40ca2b`，187 commits / 340 路径 / +20,708/-1,652；官方 VERSION 为 0.2.21，Fork 设为 0.2.22。保留全部定制，新增两份 251 SQL 与订阅重置历史 251 并存；历史 SQL 不改，不执行迁移。验证见 `openspec/changes/sync-upstream-v0-2-22/`。
- **v0.2.21**：Claude Fable/Mythos 5.2、OpenAI 30m/12h 缓存写入计费、自适应请求体内存预算、账号/API Key 页面性能、EasyPay 退款成功码及认证/协议/调度修复。官方目标 10ba8f9b4dc2e6023d4a9879ab77b21841babd10，82 commits / 200 files / +6,699 / -541；官方 VERSION 仍为 0.2.20，Fork 设为 0.2.21。同步 199 个非部署配置路径（含版本），保留全部定制和初始 3 个未提交兑换文件；新增两个 SQL 仅交付、不执行，315 份历史 SQL 不变。验证见 openspec/changes/sync-upstream-v0-2-21/。
- **v0.2.20**：管理 MCP/OAuth、管理员密钥权限扩展、精确邮箱黑名单、图片尺寸分流与 GPT Image 2.6、资源预算/JSON 热路径优化及协议兼容修复。官方目标 `dab829a4e9b7d1cc307189aea8b0319eb29e29d0`，112 commits / 459 files / +29,482 / -1,457；官方 VERSION 仍为 0.2.19，Fork 设为 0.2.20。保留全部 Fork 定制及同期批量兑换码功能；新增 `249_mcp_oauth.sql` 仅交付、不执行，历史 SQL 与部署配置不变。验证见 `openspec/changes/sync-upstream-v0-2-20/`。
- **v0.2.19**：管理员 API Key、按张图片计费、GPT-6.1 Terra、Grok 模型诊断及 Gemini/计费修复；新增 247/248 SQL，与 Fork 同号迁移共存。
- **v0.2.18**：Vertex service_account 配额、Gemini/Antigravity 生图门禁、流式断连计费与错误收尾、Codex 过载恢复和身份收敛、调度快照/模型映射优化、Grok Fast 定价及图片/工具兼容修复。目标 `ac1b8f1271be0d674fa6f074403c1410f30a3a81`，45 commits / 114 files / +7,482 / -566。官方 VERSION 仍为 0.2.17，Fork 设为 0.2.18；Go 1.27.3、DOMPurify 3.4.17、follow-redirects 1.16.1 随上游升级。新增 `246_enable_gemini_image_generation_groups.sql` 与 Fork 246 邮件迁移并存，历史 SQL 不改、不执行迁移。保留全部 Fork 定制，含 nianzs Kiro 增量、零余额群发和 PR #8139；Kiro Responses credits 继续在新流式收尾中累计。验证见 `openspec/changes/sync-upstream-v0-2-18/`。
- **v0.2.17**：Gemini 4 Argon、思考档位/摘要透传、OpenAI API Key 搜索模拟、GPT-5.5 订阅退役处理、Responses created_at、协议与模型目录修复、EasyPay 上游确认、Stripe 快照恢复和退款 pending 保留扣减。目标 `060d0160417782cd962a0c7f460f6cd4565d4916`，48 commits / 153 files / +7,035 / -808。官方 VERSION 仍为 0.2.16，Fork 设为 0.2.17。无新增迁移或依赖，历史 SQL 不变；部署示例不改。保留全部 Fork 定制，包括代理故障拒绝直连、邮件样式及未活跃用户余额券群发。搜索保留 Kiro 缓存计量和 Anthropic OAuth 原生边界；退款 pending 审计保存来源卡，成功终结事务中作废，失败保留。验证见 `openspec/changes/sync-upstream-v0-2-17/`。
- **v0.2.16**：Claude Haiku 5.5、Gemini 显式上下文缓存、Vertex API Key / 模型同步、AI Studio `AQ.` Key、临时不可调度作用范围、分组多语言和运维账号列，以及 Responses 回退、图像限流、401 规则和支付倒计时等修复。目标 `01842576b3d5c1e2545b5a99ad555d36e28ed306`，79 commits / 213 files / +13,886 / -727。官方 VERSION 文件仍为 0.2.15，Fork 设为 0.2.16。新增 `242_group_i18n.sql` 与既有 `242_drop_platform_check_constraints.sql` 并存，并新增 `243`/`244`/`245`；历史 SQL 不变。Kiro 字段索引保留在分组 `i18n` 之后，缓存存储费继续使用冻结 `pricingAt`。Gemini 缓存路由继续经过 Access Ban。验证见 `openspec/changes/sync-upstream-v0-2-16/`。
- **v0.2.15**：统一平台目录、Command Code/Cline、多协议模型路由与发现、Anthropic 缓存/签名、Responses 工具历史、WS 每轮分组计费与图像用量、监控 TPS 和前端生命周期修复。目标 f2669c8cf62555cd92389b3f55920e9e6e7c6ff2，146 commits / 278 files / +19,277 / -1,241。Kiro 登记到共享目录，保留 XorPay、Access Ban、审计、永久重置卡、Ops、品牌和支付定制。新增 242 迁移仅交付，历史 SQL 不变；Go 1.27.2 与已有依赖安全升级同步到构建配置。验证见 openspec/changes/sync-upstream-v0-2-15/。
- **v0.2.12–v0.2.14**：TypeSafe Jev System One 原生平台、充值赠金/折扣阶梯、账号优先级快捷调整、API Key 分组排序；验证码原子限次/密码重置单次消费、Antigravity 错误脱敏、匿名订单限流、API Key 删除后的用量结算、EasyPay 回调防伪、新装管理员校验与 Codex 模型发现修复。目标 commit `0363b8cdba8cec3e2ba4b2dbd49c4481143fa55d`，47 commits / 176 files / +6,801 / -475；合入 168 个非部署路径，8 个部署路径保持原样。新增两个 `241` SQL，平台约束同时保留 Kiro 与 TypeSafe；301 个历史 SQL 不变。前端 Axios/Vue/source-map-js 跟随上游安全升级；全部 Fork 定制保留。验证见 `openspec/changes/sync-upstream-v0-2-14/`。
- **v0.2.11**：余额在途请求预留、API Key 创建数量/频率限制、Claude 原生额度重置操作、GPT-6.1 Sol、Astra Ultrafast、Codex 远程模型目录/订阅档位识别，以及 Claude Code only 分组兼容入口降级修复。目标 commit `96f4c115c9749078f90cbf210a01d39baf3f53b6`，23 commits / 90 files / +5,146 / -296；官方 VERSION 仍为 0.2.10，Fork 设为 0.2.11。合入 89 个非部署路径，部署示例保持原样；无新增迁移/依赖。保留全部 Fork 定制；原生 Claude 重置与订阅永久重置卡独立。验证见 `openspec/changes/sync-upstream-v0-2-11/`。
- **v0.2.10**：Sonnet 5.5 多协议支持、Claude 原生重置额度查询、风控用户白名单、仪表盘费用趋势，以及流式用量、聚合 WS 路由、工具重写、模型白名单映射和 Antigravity 保活修复。目标 commit `2f3fed2fdb0787141294cec81487a5df30426f7f`，33 commits / 118 files / +3,759 / -332；官方 VERSION 仍为 0.2.9，Fork 设为 0.2.10。无新增迁移/依赖。保留全部 Fork 定制、默认今天的仪表盘日期、Kiro credits 与历史模型兜底价格；Claude 原生额度查询与订阅重置卡独立。验证见 `openspec/changes/sync-upstream-v0-2-10/`。
- **v0.2.9**：模型白名单任意位置通配符、协议转换/推理/工具参数恢复、客户端断开 499、账号长上下文成本开关、渠道图片价格继承、Free Fast 零成本日志，以及 OpenAI 配额重置/WS、Antigravity、模型广场视频倍率和 CC Switch 修复。目标 commit `4c00df2e0183e2c70b7fa8ba45914205e36aad0c`，70 commits / 117 files / +3,133 / -378；官方 VERSION 仍为 0.2.8，Fork 设为 0.2.9。无新增迁移/依赖；三个部署配置未同步。保留全部定制及冻结 pricingAt、请求 context、单次推理倍率和 DeepSeek 峰谷计价。验证见 `openspec/changes/sync-upstream-v0-2-9/`。
- **v0.2.8**：OpenCode Go 用量窗口和管理端查询、Claude Code 版本自动同步、reasoning effort 分级计费、月度备份归档与保留、联盟线下提现幂等、Codex 推荐/积分、TypeSafe 内容审核、GPT-6 Sol/Luna、Claude Opus 5.5、图片余额与多协议兼容，以及网关/调度/代理/前端竞态修复。官方 commit `fd80b08c90b55edcad5b00171b53f08721d30da1`，473 个文件、+28,638/-2,096；新增迁移 `238b_content_moderation_engine_meta.sql`、`239_channel_reasoning_effort_multipliers.sql`、`240_affiliate_ledger_operation_id.sql`。本 Fork 保留 Kiro/XorPay/Access Ban/订阅重置卡/提示词审计/Ops/UI 定制，配额 CHECK 同时保留 Kiro、MiniMax、OpenCode Go。
- v0.2.7：Seedance Ark 原生视频任务、插件宿主 KV/账号目录与只读状态桥、兑换记录分页，以及 Antigravity/Gemini/DeepSeek/Anthropic、国产配额 403 暂停、OAuth 刷新、用量汇总和支付/交互修复。精确增量 71 commits / 131 files / +7,283 / -446，目标 commit `aea725f2ea644d5592d0bbb1d63b607efa7e200a`；官方 VERSION 为 0.2.5，本 Fork 设为 0.2.7。298 个历史 SQL 不变，无新增迁移。保留全部 Fork 定制；新 Seedance 根路由继续经过 Access Ban；插件账号目录在 Wire provider 中接线，支持重新生成。插件身份解析拒绝非 active 账号，避免禁用后通过旧账号 ID 获取凭据。严格从 v0.2.5 同步，不恢复官方 v0.2.7 已不包含的 Codex 票据模块。验证见 `openspec/changes/sync-upstream-v0-2-7/`。
- v0.2.5：OpenCode Zen/GO、多协议原生 Codex Images、API Key 服务商筛选与批量编辑、订阅批量操作、站点计费模式、注册确认密码及模型查询，并合入配额、认证、订阅并发、Antigravity/OpenAI/Grok/监控等修复。精确增量 197 commits / 447 files / +23,038 / -1,538，目标 commit `86f93c28ee34cc74b629dafb748bd5ac5ca8c5ea`；官方 VERSION 为 0.2.4，本 Fork 设为 0.2.5。新增两个 `238` 迁移，平台 CHECK 同时保留 Kiro 与 OpenCode，296 个历史 SQL 不变。批量删用户复用 Fork 接口；延期行锁保留月窗口对齐，固定期限重复分配仍发重置卡。全部定制保留；六个部署配置路径未同步，不执行真实迁移或部署。实施与验证见 `openspec/changes/sync-upstream-v0-2-5/`。
- v0.2.4：MiniMax 多协议平台、Grok 媒体资格控制、OpenAI 周成本估算与 Image 2.5、监控用户排行开关、支付帮助 Markdown，以及代理备用关系/部分更新、流式取消、HTTP/2 保活、渠道跨实例缓存和 Ops 日志保留修复。精确增量 70 commits / 266 files / +5,696 / -726，目标 commit `5de5e2bed035d43591a2e10e51f420ef6a84eb98`；官方 VERSION 为 0.2.3，本 Fork 设为 0.2.4。新增 `237_add_minimax_platform.sql` 的配额/Composite 约束同时保留 Kiro 与 MiniMax；295 个历史 SQL 不变。保留全部定制，管理端 Badge 继续显示 Kiro，用户端仍显示 Claude；不执行真实迁移或部署。实施与验证见 `openspec/changes/sync-upstream-v0-2-4/`。
- v0.2.3：Ollama Cloud DeepSeek 跨 Chat Completions/Responses/Messages 输出上限及 Anthropic Bearer 鉴权修复、账号测试模型显示名补全，以及 `236_group_model_allowlist_repair.sql` 白名单列收敛修复。精确增量 9 commits / 22 files / +1,517 / -24，目标 commit `8fa67d477d6651a744754392a8982ea589c26ae6`；官方 VERSION 仍为 0.2.2，本 Fork 设为 0.2.3。4 个重叠代码文件保留全部原有 Fork 差异，历史迁移不变；本次不执行数据库迁移或部署。实施与验证记录见 `openspec/changes/sync-upstream-v0-2-3/`。
- v0.2.2 官方能力：分组模型白名单、简易模式基础分组/账号绑定、推理拒绝映射、Codex 固定账号标准模型发现与 Astra Ultra/Pro/图像/instructions 兼容；支付与管理员发货隔离公共兑换失败计数，备份与迁移共享 advisory lock，以及 WS、Claude、Grok、Gemini/GLM/DeepSeek 计费和前端体验修复。新增 `235_group_model_allowlist.sql`，把 `models_list_config` 原值保留并更名为 `model_allowlist`，同时约束模型列表与实际准入，部署前审核既有开启配置，不能直接回退旧二进制。目标 commit `5485f368b29d05adb95a00f71801c7c23d8f48af`，tag object `86495464d82f78da55792db6aa5568ee1302716a`，102 commits / 261 files / +13,430 / -2,005。官方 VERSION 为 0.2.1，本 Fork 设为 0.2.2；保留全部定制，Access Ban 继续覆盖新的根路径路由 helper，支付/兑换继续透传重置卡 PurchaseSource，账号统计继续使用冻结 pricingAt 和单次 max 倍率。实施记录见 `openspec/changes/sync-upstream-v0-2-2/`。
- v0.2.1 官方能力：Codex 固定账号模型目录与回退、GPT-6 Astra/ultrafast、none 推理映射、精简账号列表、用量上游请求 ID、图片 URL 转 Base64（私网阻断和字节类型检测）、定价文件热重载、Claude CLI 版本覆盖/max 推理倍率，以及 WS 续聊、会话槽释放、渠道调度、Ops 代理归因/零值指标和 Alipay 待支付补偿。四个新增迁移为 `232_add_usage_log_upstream_request_id.sql`、`233_add_usage_log_upstream_request_id_index_notx.sql`、`234_channel_max_reasoning_effort_multiplier.sql`、`234_group_codex_models_manifest_config.sql`。目标 commit `578785ee7fb35030b094b69624efe25670a36f5f`，tag object `adc26f68f687685e847bfb997559f48e79cac475`；相对 v0.2.0 共 82 commits / 297 files / +12,622 / -1,033。官方 VERSION 仍为 0.2.0，本 Fork 设为 0.2.1；保留 87 个重叠文件中的定制，特别是 Kiro 精简列表状态、credits 与请求 ID 并存、冻结计价时刻和单次 max 倍率。实施与验证记录位于 `openspec/changes/sync-upstream-v0-2-1/`。
- v0.2.0 官方能力：分组级 OpenAI 强制/免费 Fast、按模型 reasoning effort 映射与 ceiling 超限拒绝/降级、Kimi 原生 Responses、Claude Fable 5.1、无 call ID 自动化启动；同时修复 WS terminal event 前关闭、模型冷却覆盖 404 `model_not_found`、Anthropic fallback beta、OpenAI API Key 缓存身份和 scheduler passthrough 快照，并为渠道增加可空 `cache_write_1h_price`。新增 `232_channel_cache_write_1h_pricing.sql`、`232_group_force_openai_fast.sql`、`232_group_reasoning_effort_over_limit.sql`、`233_group_free_openai_fast.sql`。精确增量 `2ac784c51a5d0925b324efef2ba6b3446c364781..aa236488351eb71e120fc2b6fb32e36b0374c918` 为 60 commits / 148 files / +4,926 / -543；tag object 为 `dd07c4d8d484878e617c945cc8bacc304a5a6560`，官方 tag 内 VERSION 仍为 0.1.185，本 Fork 设为 0.2.0。54 个重叠文件逐 hunk 合并并保留全部 Fork 定制。
- v0.1.185 官方能力：价格目录支持 `pricing.override_file` JSON 补丁覆盖，长上下文阶梯计价改由价格目录数据驱动，Codex 快速模型展示 priority service tier；账号统计成本统一应用模型定价策略与 DeepSeek 峰谷价格，并增强数据库启动瞬时错误重试、OpenAI WebSocket 空闲连接回收、Codex 图像输入能力保留和持续禁用账号过滤。同步还修复 API Key 请求错误合成 instructions、ctx_pool 容量降载错误码和 delegation bootstrap 缺少 call id。无新增迁移；官方 tag peeled commit 为 `2ac784c51a5d0925b324efef2ba6b3446c364781`，tag 内 VERSION 为 0.1.184，本 Fork 按 release tag 设为 0.1.185，并保留全部 Fork 定制。
- v0.1.184 官方能力：Codex 路由模型目录、实际路由能力同步和精确账号模型别名；管理员限制用户可访问公共分组；原生 compaction 与映射前推理强度用量记录；智谱团队 GLM Coding Plan 用量查询；国产三家平台账号挂载 Ollama Cloud 用量窗口；OpenAI 图像工具冷却和 TTFT 管理配置。同步还包含 OpenAI 多渠道 service tier、配额原子重置、WebSocket 隔离/大请求转发，以及 Anthropic/Grok/支付/SMTP/计费等修复。新增 `231_add_usage_log_native_compaction_v2.sql`、`231_add_usage_log_requested_reasoning_effort.sql`、`231_user_restrict_public_groups.sql`；官方 tag 内 VERSION 为 0.1.183，本 Fork 按 release tag 设为 0.1.184，并保留全部 Fork 定制。
- v0.1.183 官方能力：OpenAI OAuth 配额耗尽 429 按 5 小时/7 天窗口重置时间暂停，普通瞬时 429 继续重试；Codex `session-id` 粘性与容量溢出绑定保护；Responses custom tool/tool search 恢复后的类型化 item ID；Kimi 并发 403 临时冷却与故障转移；邮箱换绑别名占用检测和事务级并发保护；Antigravity 兼容模式 64000 token 上限；频道监控 V2 Composite 平台聚合 SQL 修复。无新增迁移；官方 tag 内 VERSION 为 0.1.182，本 Fork 按 release tag 设为 0.1.183，并保留全部 Fork 定制。
- v0.1.182 官方能力：OpenAI Responses Lite 统一 OAuth/API Key/HTTP/WS 处理、固定并行工具调用并保留数值精度；OAuth 图片生成原样保留用户提示词；OpenCode Go 正确解析用量重置时长；Anthropic 缓存创建明细去重计费；Antigravity 修正 Sonnet 4.5 兼容路由并保留显式 4.5；Composite 支持 Kimi Code K3；渠道监控 V2 将 Composite 错误归属到真实账号平台；余额充值完成后刷新用户余额。无新增迁移；官方 tag 内 VERSION 仍为 0.1.181，本 Fork 按 release tag 设为 0.1.182，并保留全部 Fork 定制。
- v0.1.181 官方能力：Gemini 工具 schema 递归移除 `deprecated`、规范化标量 enum 并丢弃非标量 enum；Responses Lite 识别 `additional_tools` 并保留必要的 `parallel_tool_calls=false`；Responses 拒绝字段重试一次清理同类 input item 的无效 `status`；Grok OAuth、模型探测与计费探针统一使用官方 CLI User-Agent，CLI 身份更新到 `0.2.120`。无新增迁移；官方 tag 内 VERSION 仍为 0.1.180，本 Fork 按 release tag 设为 0.1.181，并保留全部 Fork 定制。
- v0.1.180 官方能力：OpenAI 重置卡按用量阈值自动使用、Responses/Chat/WS Fast mode `service_tier` 传递与按上游实际档位计费、可配置模型列表响应读取上限、模型广场上下文阶梯/渠道分时定价、Ops 错误详情返回列表，以及实验性 OAuth 出站传输插件；新增迁移 `229_plugins.sql` / `230_plugin_artifacts.sql`。本 Fork 在新插件与自动重置服务 Wire 注入之外继续注入 Kiro、Kiro OAuth 与 IP Ban，并保留 Kiro 作为 Composite、渠道匹配和调度快照的具体平台。
- v0.1.179 官方能力：国产供应商自适应 API 协议、Composite Codex/Kimi/智谱/DeepSeek 路由、渠道 fast/flex 与上下文区间倍率、Anthropic Fast 计费、可配置代理探测、`/v1/responses/input_tokens` 和用量聚合优化；新增迁移 `226_add_usage_log_effective_model_indexes_notx.sql`、`227_composite_routes_add_cn_providers.sql`、`228_channel_pricing_multipliers.sql`。长上下文计费门控由“分组且账号开启”改为“任一开启”；本 Fork 将 Kiro 合并进共享平台目录、Composite 与调度快照并继续保留全部定制。
- 用户端用量分析与管理员端对齐（`UsageView` 重构、`request_type` 筛选等）
- 订阅支付金额显示修复
- `prefer_soonest_reset` 账号调度可选策略
- 订阅支付推广返佣（affiliate rebate）
- 注册/绑定邮箱后缀白名单（`registration_email_suffix_whitelist`）
- Grok 平台官方支持（OAuth、Responses 转发等）
- OpenAI Spark 链接型影子账号（`gpt-5.3-codex-spark` 独立配额调度）
- Grok media routing / 图像上传转换
- Anthropic OAuth dateline 指纹归一化与 Claude Sonnet 5 适配
- 订阅撤销软删除缓存失效修复、账号列表 Count clone 修复、Codex OAuth reasoning 保留修复
- v0.1.146 官方能力：账号请求头覆写、账号数据拖拽/批量导入增强、API key 并发统计、EasyPay 自定义支付方式、订阅 CNY 预览、Redis 索引清理加固、OpenAI Responses 入站端点归一化与新模型/定价更新
- v0.1.155 官方能力：Grok 渠道健康监控、Web SSO 批量导入、免费配额滚动 24 小时估算与 Free 计划徽标
- v0.1.155 官方能力：系统日志 Host 筛选、可选 Server Timing、调度器全量重建合并与到期事件修复
- v0.1.155 官方能力：OpenAI HTTP/2 keep-alive、reset credits、账号级长上下文计费，以及 Responses namespace / 图像生成链路修复
- v0.1.156 官方能力：OpenAI Agent Identity、Responses/ChatCompletions 兼容与流式边界修复、图片工具链和首包超时加固
- v0.1.156 官方能力：Grok OAuth 池健康与刷新容错、账号重复创建保护、调度缓存生命周期与批量重建优化
- v0.1.156 官方能力：账号/Key ID 列、OpenAI 认证模式提示、DataTable 缓存修复及 Server Timing 扩展
- v0.1.157 官方能力：操作审计日志、会话 IP/UA 绑定、敏感操作 step-up 2FA 与管理员角色提升保护
- v0.1.157 官方能力：异步图片生成任务、对象存储结果、图片输入 token/费用独立计费
- v0.1.157 官方能力：上游计费倍率探测与调度、API Key 倍率自省、Grok 自定义上游与渠道监控复制
- v0.1.158 官方能力：用户限额批量修改、分组一键复制、Grok 上游端点快捷切换，以及 Grok OAuth / WebSocket / 图片链路修复
- v0.1.159 官方能力：OpenAI 独立搜索 API Key 调度修复、真实客户端 IP 识别统一、Grok Free 函数工具缓存、Stripe 按需加载与 API Key 上游跳转
- v0.1.160 官方能力：OpenAI 兼容提示词安全审计、Grok 媒体账号隔离、被动图片 namespace 权限修复、S3 配置 step-up TOTP 与审计事件筛选删除
- v0.1.161 官方能力：敏感操作/会话绑定安全开关、Grok 受保护视频代理与媒体链路修复、OpenAI WS/Responses 流式修复、Docker 跨架构构建
- v0.1.162 官方能力：客户端 IP 解析可配置、异步生图对象存储后台配置、Grok 客户端工具缓存、更新检查 GitHub Token、Codex/计费/订阅展示修复
- v0.1.163 官方能力：分组级 OpenAI 推理策略、Grok `/responses/compact`、Redis ACL 用户名、优雅关停用量清理，以及调度、计费、SSE、移动端集中修复
- v0.1.164 官方能力：聚合分组、Ollama Cloud 用量同步、支付宝移动端深链，以及 OpenAI/Grok/计费/审计修复；本 Fork 将 Kiro 补入聚合分组合法目标、模型候选和渠道定价匹配
- v0.1.169 官方能力：上游 URL 路径片段闭集校验（GHSA-vrxq-qm4h-6hgg）、代理断流熔断 fail-open、glm-5.2 定价、订阅到期标签与套餐标题可读性、SMTP/Passkey/Qwen3Guard 辅助字段等修复；本 Fork 在 `gateway.go` 保留 Access Ban 的同时挂上 `guardResponsesSubpath`
- v0.1.170 官方能力：分组级利润控制、API Key 全平台上游倍率探测与可选自动同步、`profit-preview` 离线预演、内容审核代理与账号全结果批量选择；同时合入 Anthropic 流中断部分用量计费、OpenAI/Grok 转发、订阅窗口和 SMTP 等修复；本 Fork 继续保留 Kiro 粘性 TTL、Access Ban、XorPay、Kiro→Claude 别名和自定义确认弹窗
- v0.1.171 官方能力：腾讯天御/阿里云验证码 2.0 与 Turnstile 三选一门禁、Codex 出站身份统一和版本同步、composite 推理强度、OpenAI reset credit 缓存、退款 `require_force` 与 Stripe 幂等、订阅并发锁、token refresh 单飞、计费失败用量留痕、Messages 临时错误 failover、WS/调度取消和模型广场图片价格修复；本 Fork 在这些能力上继续叠加 Kiro/XorPay/Access Ban、续期定制、提示词审计与 UI 品牌定制
- v0.1.172 官方能力：安全加固、上游响应模型审计与迁移 `194`/`195`、订阅日额度午夜重置、金额按 `NUMERIC(20,8)` 量化，以及网关、OpenAI/Grok/Antigravity 模型兼容、腾讯验证码、Ops 日志、代理超时和模型广场等修复；官方 tag 内 VERSION 仍为 0.1.171，本 Fork 按 release tag 设为 0.1.172，并完整保留 Kiro/XorPay/Access Ban、Kiro→Claude、VersionBadge、支付与订阅定制
- v0.1.173 官方能力：Grok 邮箱密码 SSO、refresh token 重认证与 Redis 跨实例会话，媒体 Voice、视频/搜索计费和调度，渠道监控 V2、邮箱域名限量注册，以及 Gemini/OpenAI 集中修复；新增迁移 `194`–`206`、`217`–`220`，默认启用 Grok 免费额度软门禁、默认隐藏 V2 吞吐量并由迁移 `220` 备份后清理非 Grok/非 composite 视频价格，渠道监控则保持 V1、V2 显式启用；本 Fork 继续保留 Kiro/XorPay/Access Ban、Kiro→Claude、VersionBadge、GLM 与支付定制
- v0.1.175 官方能力：Codex OAuth 设备指纹收敛、按上游响应模型计费、大文件备份分卷上传与恢复，以及 OpenAI/Grok/Gemini/WS/审计/调度等集中修复；官方无 v0.1.174 tag，官方 tag 内 VERSION 仍为 0.1.173；本 Fork 按 release tag 设为 0.1.175，备份页保留 ConfirmDialog/密码弹窗并叠加分卷下载，同时完整保留 Kiro/XorPay/Access Ban、Kiro→Claude、VersionBadge、GLM 与支付定制
- v0.1.176 官方能力：Grok 4.6 / JWT 订阅档位识别、分组逐模型定价与长上下文阶梯开关、原生 `POST /x_search`，以及备份 leader 锁、渠道缓存、定价冲突、Responses 探测、Realtime 音频计费等修复；新增迁移 `221`。官方 tag 内 VERSION 仍为 0.1.175；本 Fork 按 release tag 设为 0.1.176。Gateway `/x_search` 保留 Access Ban，分组 `oneof` 仍含 `kiro`，Kiro 缓存模拟 UI 与逐模型定价 UI 并存，并完整保留 Kiro/XorPay/Access Ban、Kiro→Claude、VersionBadge、GLM 与支付定制
- v0.1.177 官方能力：迁移 `222`/`223` 建立分组用量日汇总并切换为服务端配置时区，分组页与仪表盘统计提供 `today` / `yesterday` / `total`；Codex remote compaction v2 使用会话级 `remote_compaction_v2` beta header，原生 v2 保留 `/responses` 路由，`x-codex-turn-state` 跨 HTTP/SSE/WS 回传并通过来源记录阻止跨账号回显，指纹收敛默认关闭且仅显式 opt-in；同时修复 Grok 长上下文计费、带版本媒体模型识别和账号页自动刷新偏好。官方 tag 内 VERSION 是 0.1.176；本 Fork 按 release tag 设为 0.1.177，并完整保留 Kiro/XorPay/Access Ban、提示词审计、套餐续期、Ops、UI 品牌及既有支付定制
- v0.1.178 官方能力：Kimi/智谱/DeepSeek 多协议供应商、渠道监控配额模式、渠道模型分时倍率、OpenAI Team 联动熔断、OpenAI 账号批量设置，以及 Codex/Gemini/Claude/WS/Ops 集中修复；官方迁移 `224_user_platform_quotas_add_cn_providers.sql`、两个 `225_*` 和 `226_channel_monitor_quota_mode.sql` 与 Fork 订阅重置卡 `224`/`225` 按完整文件名并存；本 Fork 完整保留全部既有定制。

v0.1.160 同步时额外补齐 `securityaudit.ProviderSet` 中 `*PromptService` 到 `PromptAdminService` 的 Wire 接口绑定，确保当前 Fork 可重新生成 `wire_gen.go`；其余提示词审计实现保持官方版本。

---

## v0.2.25 同步边界（2026-11-24）

- 请求头身份变量及错误详情、上游分组名称/备注手动与默认关闭的自动同步、billing include=group、OpenAI 透传错误规则/结构化 400 与 compact 过载重试修复。
- 36 个上游改动路径存在 Fork 差异，按 v0.2.24 三方整合；保留身份认证、访问封禁、Kiro credits/粘性、全期额度、重置卡与历史、XorPay、邮件券近期活跃范围、批量兑换及生图 NoContentPersistence。
- 身份变量只用于已认证 API Key 请求；共享模型列表与后台请求不上报用户身份，WS 按连接身份隔离，错误详情中的模板固定文本脱敏。Kiro 专属模型抓取同步适配新请求头方法签名。
- Kiro 保持排除通用倍率探测，创建/编辑/列表均不展示分组同步入口，创建 payload 不发送探测/自动同步开关；避免向上游添加不支持的调用或导致创建失败。
- 无新增或修改迁移；已有 gorilla/websocket 跟随目标标签更新到 `v1.5.4-0.20240701034025-d67f41855da4`。官方 VERSION 仍为 0.2.24，Fork 标为 0.2.25。未执行部署、生产迁移或真实外部业务请求。
- 规划、回归和审查见 [同步记录](../openspec/changes/sync-upstream-v0-2-25/verification.md)。

## v0.2.24 同步边界（2026-11-20）

- 以 v0.2.23 文件为共同基线整合发布标签，保留 Fork `67bcc8889` 的全部定制。190 个发布路径中 54 个与 Fork 差异重叠；不引入目标之后的 main。
- 公共与生图内部网关使用同一个 `RegisterGatewayRoutes`，新增重复字段拒绝在模型白名单之后、Composite 改写之前；Access Ban 仍在认证前。BOM 与伪 multipart JSON 同样接受模型准入检查，文本路由在读体前应用正确大小上限。
- 保留 Kiro 会话字段/credits、窗口用量、用户侧 Claude 品牌、全期订阅与永久卡、XorPay、邮件券及生图工作台。缓冲 Responses 补全工具/图片输出时继续提取 Kiro credits；Antigravity 429 每请求最多冷却一账号且日志保持 NoContentPersistence 脱敏。
- 审核复制优化继续尊重生图同步安全判定与禁止内容持久化，不将创作正文放入异步队列。无新增依赖或迁移，所有历史 SQL 不变；不执行生产迁移、部署或实际模型/支付请求。
- 官方 VERSION 仍为 0.2.23，Fork 按发布标签设为 0.2.24。重复 JSON/表单字段（包括相同值、大小写变体及嵌套字段）会被严格拒绝；旧有非标准客户端应删除重复字段。验证与审查见 [同步记录](../openspec/changes/sync-upstream-v0-2-24/verification.md)。

## v0.2.22 同步边界（2026-11-17）

- 以官方 v0.2.21 为三方基线整合发布标签 v0.2.22；初始 Fork HEAD `e9c30cb40`，无已有未提交改动。339 个发布路径同步，`deploy/config.example.yaml` 不覆盖；不引入标签后的 main。
- 新合并扣费保留订阅 USD/token（含零金额请求）累计、请求幂等和全期额度，不恢复 30 天自动清零；金额继续按 USD NUMERIC(20,8) 口径量化，复用既有事务更新与锁顺序。Kiro credits、冻结 pricingAt/context、30m/12h 缓存与账号统计定价并存。
- Vertex 多区域 1.1 倍仅作用于真实多区域默认模型价，自定义分组/渠道价不叠加；Kiro 直连不受影响。Kiro 直连分支、代理失败拒绝直连及粘性 TTL 与上游 OAuth 轮换和断连后限时绑定共存。
- 端点费用图表新增管理员账号成本，同时保留用户端标准成本/账号成本隐藏；近期客服、密钥模型、CC Switch 品牌、订单筛选及缓存命中率保持。
- 新 `251_group_force_openai_ultrafast.sql` 和 `251_group_usage_rollup_nonblocking_build.sql` 与 Fork 订阅历史迁移按完整文件名并存。上线前备份并验证新列、汇总状态/触发器与重建期间写入；不执行迁移，回退不得自动删列或索引，须验证汇总兼容性。
- Go 声明同步 1.27.4；jsonschema-go 0.4.2 从间接依赖改直接引用，无新依赖版本。保留镜像源、安装策略与部署配置；新增扣费通道和 Fork 邮件 Worker 均保留停机清理。验证结果见 [本次记录](../openspec/changes/sync-upstream-v0-2-22/verification.md)。

## v0.2.21 同步边界（2026-11-11）

- 精确同步官方 v0.2.20 → v0.2.21，不引入标签后的 main。基于初始 Fork HEAD 1e8e23216 三方合并，保留 3 个未提交的批量兑换源码/测试文件；不修改部署配置。
- 用量 SQL 的参数、类型、列及扫描顺序同时保留 Kiro credits、真实上游请求 ID 和新增 30m/12h token，静态插入共 65 个参数。Kiro cache_read 不重复计量、冻结 pricingAt、独立 429 冷却及原有计费规则不变。
- API Key 列表独立加载用量统计，同时保留用户端 Claude 品牌和倍率隐藏；账号刷新保留现有行，设置页吸附保存按钮与客服二维码/余额券邮件模板并存。
- EasyPay 每实例可显式选退款成功码 0 或 1，默认仍为 1；XorPay 和永久重置卡退款来源、幂等及事务不变。刷新令牌原子消费与既有 Access Ban/邮箱校验/登录失败封禁并存。
- 新增 249_add_usage_logs_api_key_cost_index_notx.sql 与 249_mcp_oauth.sql 按完整文件名共存；250_openai_cache_write_ttl.sql 新增 token 计数和四张价卡表的可空 12h 单价（NUMERIC(20,12)），既有通用缓存写入价格保持覆盖语义。迁移只交付不执行；部署前备份，测试大表 CONCURRENTLY 索引、失败索引重试及新列读写。降级前应核验新价卡兼容性，不自动删除表、列或索引。
- 无 go.mod、前端依赖或工具链版本变更；go.sum 仅同步上游校验和。已有退款 pending 降级约束继续适用。

## v0.2.20 同步边界（2026-11-07）

- 以 v0.2.19 为三方基线，保留 Kiro、XorPay、Access Ban、永久重置卡、余额券群发、客服二维码、订阅倒计时及同期新增的批量兑换码；不带入目标标签后的 main。
- MCP transport、OAuth 注册/令牌/撤销/发现以及用户授权接口接入 Access Ban；新精确邮箱黑名单与旧格式校验、邮箱后缀封禁并存，既有账号按上游规则登录。普通 IPBanGuard 命中后显式终止处理链，防止返回403后继续执行下游。
- 修复公共图片计费入口遗漏的按张开关判断：专用图片端点按组配置计费，混合请求仍遵循 token 定价，保留 Kiro opts。
- 新请求解析缓存保存/恢复 Kiro `BodySessionID` 并复制标量字符串；鉴权缓存同时保留 Kiro 配置、图片张数计费与隐私门；Kiro 被动用量查询遵守受限管理员密钥权限。
- 新 `249_mcp_oauth.sql` 只新增授权、一次性授权码与令牌摘要表；MCP 默认关闭，需显式配置 `mcp.public_url`。不执行迁移；部署前备份并验证 OAuth 撤销/轮换。回退二进制不删除表，已发授权应先撤销。
- MCP go-sdk 1.4.1 / moego 0.1.0-alpha.1 及安全升级按上游锁定；vue-i18n 11.4.16，CI Node 24。六个部署配置路径保持原样，新参数可用应用默认值；容器启用 MCP 时须由部署负责人显式传入配置。

## v0.2.18 同步边界（2026-10-23）

- 新 SQL `246_enable_gemini_image_generation_groups.sql` 与 Fork 邮件迁移按完整文件名共存；现有 SQL 保持字节不变。该迁移使旧 Gemini/Antigravity 分组生图开关开启，兼容旧版实际行为，不影响 composite/Kiro；部署前需备份并核对开关，降级二进制不会撤销数据更新。本次不执行迁移或部署。
- Vertex 配额使用独立 `SupportsAccountQuota`，不扩大 `IsAPIKeyOrBedrock` 的池模式适用范围；保留 Fork 的订阅 token 计数、冻结计价时刻、Kiro credits 和模型兜底。
- Kiro Responses 桥接保留真实上游响应头、缓存模拟及 credits 提取，同时采用新版流式断连后累计用量/错误收尾。账号 UI 同时保留 Kiro 直连/中转、区域 Select 和上游 Vertex 配额。
- 保留永久重置卡、XorPay、封禁、严格代理、邮件样式及零余额群发；PR #8139 已包含于目标标签，不重复追加。
- 构建配置同步 Go 1.27.3，前端按官方锁文件同步 DOMPurify 3.4.17 / follow-redirects 1.16.1 安全补丁。Fork 镜像源、禁用安装脚本、部署配置不变。

## v0.2.17 同步边界（2026-10-16）

- 无新增迁移或依赖。`deploy/config.example.yaml` 保持原样；新增 `gateway.min_output_tokens` 使用应用默认 128、校验下限 16。
- 退款 pending 保留实际已扣金额，终态成功不重复扣减、失败同事务归还；Fork 余额缓存仍隔离请求取消。审计持久化来源卡 ID/待作废标志，卡与成功退款同事务作废，不提前作废、不顺延订阅。旧 pending 审计按上游兼容处理；降级前须结清新格式 pending 退款。
- 保留 Anthropic OAuth/setup-token 原生搜索边界、Kiro 直连渠道开关和缓存计量，同时支持 OpenAI API Key 模拟搜索。共享搜索入口使用 `ResolveAccountProxyURL`，账号代理关系缺失时返回 failover 错误，禁止继续选择供应商或直连。
- 近期代理不可用时禁止直连、通知邮件样式和余额券群发保持；不执行数据库迁移、部署或真实邮件发送。

## 2.1 v0.2.15 同步增量（2026-10-09）

- 同步官方 v0.2.14→v0.2.15：平台清单与 Provider Profile 收敛、多协议路由统一，新增 Command Code / Cline 平台，OpenAI Web Search/Responses、OpenCode、Grok、Anthropic、监控与前端竞态修复。
- 新增迁移 `242_drop_platform_check_constraints.sql`，平台校验统一由应用层清单负责；Fork 保留 Kiro 并将其加入同一清单、配额和 Composite 路由能力。
- 保留 Kiro、XorPay、Access Ban、订阅重置卡、提示词审计、Ops、VersionBadge、支付与用户端 Claude 品牌定制。

## 3. Kiro 平台（最大定制模块）

### 3.1 功能范围

- 平台标识：`PlatformKiro`（`backend/internal/domain/constants.go`）
- OAuth / IDC 认证、令牌刷新、主动用量查询
- Anthropic Messages 请求/响应翻译（`backend/internal/pkg/kiro/translator.go`）
- Group 级：**缓存模拟**、**自动粘性会话**、**粘性 TTL**、**推理端点模式**（`q` / `krs`）
- 计费：`usage_logs.kiro_credits` 字段；cache_read **不重复计入** input_tokens
- 429：独立冷却（`kirocooldown`），**跳过**通用 `HandleUpstreamError`，与 DB `rate_limit_reset_at` 双向同步
- **count_tokens**：Kiro 直连账号（OAuth / 无 base_url 的 API Key）走本地 `estimateKiroInputTokens`，**不打上游**，避免 `jwt auth is not yet supported on count_tokens` → OAuth 401 临时不可调度
- **v0.2.1 接入**：精简账号列表继续返回 `kiro_quota_*` / `kiro_runtime_*` 六字段；Messages 直连和 Responses/Chat 桥接单独保留真实响应头快照，按账号配置提取上游请求 ID，避免流式包装生成的 Claude ID 混入用量记录。非流式 Web Search 记录最后一次实际上游响应头；合成多步搜索流没有单一真实响应头时保持 NULL。
- **已同步 nianzs/main（2026-10-17 `1cd5141b8`）的 Kiro 增量**：图片视觉 token、`KiroEndpointModeAuto`、API Key 直连/中转分流、`profile_arn` 解析回填、External IdP OAuth 后端、工具配对与截断、模型别名/输出上限、缓存创建/读取独立比例、可配置计量算法（默认 official）。迁移 `247` 增加分组比例列和 `gateway_kiro_tokenizer`。平台 CHECK 仍保持删除，渠道监控不加入 Kiro。直连与桥接继续保存真实上游响应头快照。前端 External IdP 两阶段引导 UI 仍待补齐

### 3.2 后端 — 新增文件（独立包）

```
backend/internal/pkg/kiro/
  oauth.go, oauth_test.go, oauth_invalid_grant_test.go
  signature.go, translator.go, translator_test.go
  models.go, models_test.go, fingerprint.go, fingerprint_test.go
  websearch.go, websearch_test.go, websearch_stream.go, websearch_stream_test.go

backend/internal/pkg/kirocooldown/
  store.go, store_test.go

backend/internal/pkg/anthropictokenizer/
  tokenizer.go, tokenizer_test.go, claude.json, NOTICE.md

backend/internal/handler/admin/
  kiro_oauth_handler.go
  group_handler_kiro_validation_test.go

backend/internal/service/
  kiro_oauth_service.go, kiro_oauth_service_test.go
  kiro_token_provider.go, kiro_token_provider_test.go
  kiro_token_refresher.go, kiro_usage_fetcher.go
  kiro_runtime.go, kiro_runtime_state.go, kiro_runtime_state_test.go
  kiro_runtime_state_integration_test.go
  kiro_http_helpers.go, kiro_http_helpers_test.go
  kiro_error_classifier.go, kiro_error_classifier_test.go
  kiro_cache_emulation.go, kiro_cache_emulation_test.go
  kiro_websearch.go, kiro_websearch_test.go
  kiro_alignment_test.go, kiro_alignment_unit_test.go
  kiro_credits_usage_log_test.go, kiro_mapping_fallback_test.go
  kiro_session_test.go, kiro_sticky_session_test.go
  account_test_service_kiro_test.go, account_test_service_kiro_apikey_fallback_test.go
  account_usage_service_kiro_apikey_test.go
  account_kiro_credit_unit_price_test.go
```

### 3.3 后端 — 修改的接线文件（节选）

| 文件                                                | 改动                                         |
| --------------------------------------------------- | -------------------------------------------- |
| `backend/cmd/server/wire.go` / `wire_gen.go`        | 注入 `kiroTokenProvider`（**在 grok 之前**） |
| `backend/internal/handler/admin/account_handler.go` | Kiro 可用模型分支                            |
| `backend/internal/service/account_test_service.go`  | 构造参数含 Kiro/Grok provider                |
| `backend/internal/service/account_usage_service.go` | `ProvideAccountUsageService` 注入 Kiro       |
| `backend/internal/handler/gateway_handler*.go`      | Kiro 路由与转发                              |
| `backend/internal/service/gateway_service.go` 等    | Kiro 运行时调度                              |
| `backend/ent/schema/group.go`                       | Kiro group 字段                              |
| `backend/internal/repository/usage_log_repo.go`     | `kiro_credits` 读写                          |

### 3.4 数据库迁移（Fork 新增）

| 文件                                        | 作用                                                                    |
| ------------------------------------------- | ----------------------------------------------------------------------- |
| `135_add_group_kiro_cache_emulation.sql`    | Group 缓存模拟字段                                                      |
| `145_allow_kiro_user_platform_quotas.sql`   | CHECK 约束加入 `kiro`                                                   |
| `151_add_group_kiro_auto_sticky.sql`        | 自动粘性开关                                                            |
| `152_add_group_kiro_sticky_session_ttl.sql` | 粘性 TTL                                                                |
| `153_add_group_kiro_endpoint_mode.sql`      | 端点模式 q/krs                                                          |
| `153_add_usage_log_kiro_credits.sql`        | 用量表 kiro_credits                                                     |
| `157_user_platform_quotas_add_grok.sql`     | **已改**：CHECK 含 `anthropic, openai, gemini, antigravity, kiro, grok` |

### 3.5 前端

| 路径                                                                        | 作用                                                           |
| --------------------------------------------------------------------------- | -------------------------------------------------------------- |
| `frontend/src/api/admin/kiro.ts`                                            | Kiro 管理 API                                                  |
| `frontend/src/composables/useKiroOAuth.ts`                                  | OAuth 流程                                                     |
| `frontend/src/utils/platformColors.ts`                                      | Kiro 紫色主题                                                  |
| `frontend/src/components/common/PlatformIcon.vue` / `PlatformTypeBadge.vue` | 平台图标                                                       |
| `frontend/src/types/index.ts`                                               | `GroupPlatform` / `AccountPlatform` 含 `'kiro'`；Kiro 相关字段 |
| 账号/分组/用量相关 View 与 Modal                                            | 创建编辑账号、用量统计、配额等                                 |

#### 用户端 Kiro → Claude 品牌别名（2026-07-14）

用户端采用展示别名，覆盖以下页面：

| 页面          | 路由                  | 改动记录                                                                                                                                                                                                                     |
| ------------- | --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 用户仪表盘    | `/dashboard`          | 平台用量统计中的 `kiro` 显示为 Claude                                                                                                                                                                                        |
| API Key 管理  | `/keys`               | 分组名称、描述、徽章、图标和下拉选项显示 Claude；搜索 `claude` 可命中内部 Kiro 分组                                                                                                                                          |
| 可用渠道      | `/available-channels` | 渠道、平台、分组及模型区域使用 Claude 名称、图标和配色；搜索 `claude` 可命中 Kiro 渠道                                                                                                                                       |
| 我的订阅      | `/subscriptions`      | 订阅分组名称、描述、平台徽章、卡片边框和操作按钮使用 Claude 展示                                                                                                                                                             |
| 购买 / 续费   | `/purchase`           | Kiro 与 Anthropic 套餐合并为 `Claude Max` 分类并统一徽章；OpenAI 协议下名称、分组名或描述含 `GLM` 的套餐独立显示为 `智普-GLM Plan MAX`，并从普通 OpenAI 分类排除；套餐卡片、确认页、续费弹窗和已有订阅使用 Claude 名称与配色 |
| 用量—错误请求 | `/usage`              | 错误列表和详情的平台字段显示 Claude                                                                                                                                                                                          |

实现边界：

- 后端、API 请求参数、数据库及前端内部筛选值继续使用 `kiro`，不得改为 `claude`。
- 管理端继续展示真实 Kiro 名称、官方 Kiro 图标和紫色主题。
- 用户侧动态分组、套餐名称和描述中的独立单词 `Kiro`（忽略大小写）替换为 Claude。
- 通用搜索选项使用隐藏搜索关键字，同时包含内部值 `kiro` 和展示值 `claude`；用户下拉中不展示 Kiro。
- 错误请求的平台字段使用 Claude 别名，但上游错误消息、响应体等诊断原文保持不变，避免影响排障。

主要实现文件：

- `frontend/src/utils/platformColors.ts`
- `frontend/src/components/common/{GroupBadge,GroupOptionItem,Select}.vue`
- `frontend/src/views/user/{KeysView,AvailableChannelsView,SubscriptionsView,PaymentView}.vue`
- `frontend/src/components/channels/AvailableChannelsTable.vue`
- `frontend/src/components/payment/SubscriptionPlanCard.vue`
- `frontend/src/components/user/dashboard/UserDashboardStats.vue`
- `frontend/src/components/user/{UserErrorRequestsTable,UserErrorDetailModal}.vue`

验证记录：Vue / TypeScript 类型检查通过；完整 Vitest 共 166 个测试文件、1094 项测试通过；品牌别名、下拉搜索和购买页定向回归共 5 个测试文件、25 项测试通过。

### 3.6 环境变量（部署）

`deploy/.env.example` / `deploy/config.example.yaml` 新增示例：

- `SUB2API_KIRO_TIME_CONTEXT` — 空=最稳定 cache 前缀；可选 `date` / `precise`

---

## 4. XorPay 支付（Fork 新增）

### 4.1 功能

- Provider key：`payment.TypeXorPay` = `"xorpay"`；原 `xorpay` 支付类型表示支付宝，新增支持 `wxpay` 微信扫码。
- 下单分别发送 `pay_type=alipay` / `native`，两位小数字符串金额参与签名；异步 notify、订单查询和退款共用既有 XorPay 协议。
- Webhook：`POST /api/v1/payment/webhook/xorpay`
- 微信支付复用现有通道筛选、限额、订单实例快照、金额一致性和履约幂等；不走官方微信 OpenID/JSAPI 流程。
- 管理端通过 XorPay 实例的微信支付开关启用，保留微信通道冲突检查。服务端和配置 helper 兼容 `xorpay_wxpay` 来源，当前设置页不新增来源下拉框。原实例不自动启用微信，无迁移。操作见 [XorPay 微信配置](XORPAY_WECHAT_SETUP.md)。

### 4.2 关键文件

**后端（新增）**

- `backend/internal/payment/provider/xorpay.go`

**后端（修改）**

- `backend/internal/payment/types.go` — `TypeXorPay`
- `backend/internal/payment/provider/factory.go` — 注册工厂
- `backend/internal/service/payment_config_providers.go`
- `backend/internal/service/payment_order_lifecycle.go`
- `backend/internal/handler/payment_webhook_handler.go`
- `backend/internal/server/routes/payment.go`

**前端**

- `frontend/src/components/payment/` — `ProviderCard`, `PaymentMethodSelector`, `PaymentQRDialog`, `PaymentStatusPanel`, `paymentFlow.ts`, `providerConfig.ts`
- `frontend/src/views/user/PaymentQRCodeView.vue`
- `frontend/src/views/admin/orders/AdminOrdersView.vue`
- `frontend/src/types/payment.ts` — `PaymentType` 含 `xorpay`
- `frontend/src/i18n/locales/{zh,en}.ts` — XorPay 文案

### 4.3 Fork 支付 UX 定制（相对官方 + XorPay 本身）

| 改动               | 文件/说明                                                                             |
| ------------------ | ------------------------------------------------------------------------------------- |
| 用户端隐藏倍率展示 | `PaymentView.vue`、`UsageView.vue` CSV 导出等                                         |
| XorPay 显示名      | 原 `xorpay` 仍为 **「XorPay（支付宝）」**；新增微信订单使用 `wxpay`，显示微信图标与扫码提示                                                   |
| 二维码支付金额     | `AmountInput.vue`、`PaymentStatusPanel.vue` — `payAmount` / `creditedAmount` 分离展示 |
| 订阅支付商品标题   | `payment_order.go` 的 `buildPaymentSubject`：优先使用去首尾空白的 `product_name`，为空时回退去首尾空白的套餐名称，再应用全局前后缀；移除默认硬编码的 `Sub2API Subscription`，保留管理员显式填写的商品名。XorPay、支付宝、微信、易支付和 Stripe 共用此逻辑；余额充值标题、金额和履约不变，无迁移，已提交给支付平台的旧订单标题不回写。回归见 `payment_order_result_test.go`。 |
| 充值订阅页优化     | 提交 `9a8786ff` 等近期支付流程调整                                                    |
| 套餐展示标签       | `subscriptionPlanLabels.ts` 统一购买页筛选 tab、套餐卡片及购买确认徽章：普通 OpenAI 显示 `OpenAI(GPT Pro 50x)`，Gemini 显示 `Gemini Ultra`；保留 GLM 特例、内部平台值和套餐配置 |
| 套餐卡片布局       | `SubscriptionPlanCard.vue` 与 `PaymentView.vue` 采用方案 A：平台/有效期独立首行，标题整行显示并自然换行；取消标题两行省略，窄屏标签与价格可换行。共用卡片的购买列表、续费弹窗同步生效，购买确认头部一致；保留配色、计价及选购行为。 |

---

### 4.4 余额消费最低为零（2026-11-25）

用户选择：允许已放行请求正常结束，消费最多扣除可用余额，超额由平台承担。例如余额 1 USD、真实消费 3 USD，结算后余额 0、实扣 1、平台承担 2；后续余额付费请求沿用耗尽拦截。已放行的并发请求也可能产生承担金额；本次不改为严格预付费。

- 单笔、按用户批量结算及兼容 `DeductBalance` 共用行锁下的 SQL；余额变更与结算明细、原有幂等/配额操作处于同一事务，充值不被旧快照覆盖。保持 `NUMERIC(20,8)` 金额精度及原请求指纹。
- 新增 `254_usage_balance_settlements.sql`：持久化请求/Key/用户、真实应扣、实际扣款、平台承担、扣前/扣后余额。`requested = charged + absorbed`，`before - after = charged`。原有用量、Key/账号/平台/订阅额度不缩减为实扣值。
- 已有负余额不自动免除，新消费不扩大旧欠款；历史处理需另行批准。Redis 不因消费从非负变负，耗尽时继续失效缓存；余额提醒使用实际结算值。批量生图按钱包精度冻结/结算，超出冻结金额拒绝结算，保留原有退款/重试流程。
- 新代码需要新增迁移；未执行生产迁移或部署。回滚代码会恢复原有透支策略，结算明细表应保留，切换期间暂停余额消费并核对在途结算。
- 实施与验证见 `openspec/changes/prevent-consumption-overdraft/`。

---

## 5. Grok 平台口径（Fork 与官方交叉）

官方 v0.1.142 已含 Grok OAuth / gateway / media routing；本 Fork **额外**保证：

- `constants.go` 中 `PlatformGrok` 与 Kiro 并存
- `account_handler.go` 同时有 Kiro、Grok 模型分支
- `wire_gen.go` 注入顺序：**kiro → grok**
- 迁移 `157` CHECK 约束**同时**保留 kiro + grok（见第 8 节红线）

---

## 6. UI / 品牌 / 交互

用量费用栏（2026-11-16）：`/admin/usage` 和 `/usage` 复用 `frontend/src/components/admin/usage/UsageTable.vue`，在费用下方增加灰色描边、绿色百分比的“缓存命中率”标签，支持中英文及深色模式。计算使用日志的互斥 Token 桶：`cache_read_tokens / (input_tokens + cache_creation_tokens + cache_read_tokens)`，不包含输出或重复累加缓存写入 TTL 明细。保留两位小数；有输入但无命中为 `0.00%`，无输入数据为 `—`。不改计费、接口或用户费用详情权限；验证覆盖管理员/用户显示模式、缓存写入与历史缺失字段。

数据来源为接口返回的 `usage_logs` 已保存记录，前端不生成模拟数值。该指标反映现有计量口径：Kiro 若启用了既有模拟缓存，缓存桶来自服务端模拟计量，不能一概解释为上游实测缓存命中率；本次不改变其记录或计费行为。

| 项               | 说明                                   | 主要文件                                                                 |
| ---------------- | -------------------------------------- | ------------------------------------------------------------------------ |
| 首页与认证页配色 | 靛蓝紫罗兰主题 + 动效                  | `frontend/src/views/HomeView.vue`, `components/layout/AuthLayout.vue`    |
| 登录页公开配置 | 复用 App Store 的服务端注入、缓存及在途请求，避免重复请求失败导致注册/找回密码入口消失；保留注册开关与后台模式限制 | `frontend/src/views/auth/LoginView.vue` |
| CC Switch 导入品牌 | Codex 的供应商标识、配置节及显示名使用站点名并安全转义；空名回退 `sub2api`，内置保留 ID 同名时加 `-custom`，端点、模型及鉴权保持不变 | `frontend/src/utils/ccswitchImport.ts` |
| 密钥可用模型弹窗 | `/keys` 在“使用密钥”前增加“查看可用模型”；所有平台使用当前密钥读取网关 `/v1/models`，保留账号映射（含 Antigravity 自定义别名）与分组授权，显示真实分组名及模型、支持点击复制、加载/无分组/空列表/失败重试；取消过期请求防止串组，移动端换行，保留 Kiro→Claude 用户端别名；不改后端权限或模型路由 | `frontend/src/components/keys/AvailableModelsModal.vue`、`frontend/src/api/keyModels.ts`、`frontend/src/views/user/KeysView.vue` |
| Logo             | Fork 自定义 SVG                        | `frontend/public/logo.svg`（Flymux 风格）                                |
| Select 组件      | 替换原生 `<select>`                    | 各 admin/account View；`components/common/Select.vue`                    |
| 导航与购买文案 | 用户自定义菜单名恰为“文档”时，侧栏、页头与浏览器标题统一显示“使用教程”；不改持久化配置、其他菜单或管理专用菜单；中文“充值/订阅”入口及关联页面标题、设置说明统一为“购买算力/订阅”，既有站点计费模式与链接保持不变 | `frontend/src/i18n/locales/zh/{common,misc}.ts`、`frontend/src/i18n/locales/zh/admin/settings.ts`、`frontend/src/utils/customMenu.ts`、`components/layout/{AppSidebar,AppHeader}.vue`、`router/title.ts` |
| 用户订单状态筛选 | `/orders` 以灰色圆角底、白色选中项展示“全部 / 待支付 / 已完成 / 失败 / 已退款”五个横向按钮；保留原状态值与服务端分页，切换回第 1 页，重复点击当前状态不重载，刷新保留页码，加载期间禁用切换；支持键盘焦点、深色模式及窄屏横向滚动，不改支付/退款流程 | `frontend/src/views/user/UserOrdersView.vue`、`frontend/src/views/user/__tests__/UserOrdersView.filters.spec.ts` |
| ConfirmDialog    | 替换原生 confirm/alert                 | 多处 admin/account 组件                                                  |
| 代理广告横幅     | 已移除                                 | `ProxyAdBanner` 相关引用清理                                             |
| 版本徽章         | 隐藏「立即更新」按钮                   | `frontend/src/components/common/VersionBadge.vue`                        |
| 客服入口         | 顶栏与登录/注册页共用微信弹窗；认证页表单上方分别提示“登录不上？右侧直接找客服”“注册遇到问题？右侧直接找客服”，右侧按钮支持点击开关、悬停、Esc/外部点击关闭；复用公开配置，未配置时隐藏，保留二维码 URL 校验 | `frontend/src/components/common/SupportContact.vue`、`components/layout/AppHeader.vue`、`components/layout/AuthLayout.vue`、`views/auth/LoginView.vue`、`views/auth/RegisterView.vue`；设置项 `contact_info` / `contact_qr_code` |
| Kiro 截图 / 社群 | 文档与资源                             | `assets/screenshots/kiro-*.png`, `assets/community/wechat-group.jpg`     |

---

## 7. Ops — 管理员删除错误日志（Fork 新增）

官方 v0.1.142 **无**此 API；本 Fork 新增：

| 层         | 文件                                                                |
| ---------- | ------------------------------------------------------------------- |
| 路由       | `backend/internal/server/routes/admin.go` — `DELETE /ops/errors`    |
| Handler    | `backend/internal/handler/admin/ops_handler.go` — `DeleteErrorLogs` |
| Service    | `backend/internal/service/ops_service.go`                           |
| Repository | `backend/internal/repository/ops_repo.go`                           |
| 前端       | `frontend/src/api/admin/ops.ts`, `OpsSystemLogTable.vue` 等         |

删除需带时间范围等过滤条件（见 repository 测试 `ops_repo_error_where_test.go`）。

---

## 8. 全局访问封禁 Access Ban（Fork 新增，官方 v0.1.146 无）

> **官方 v0.1.164 不存在** `ip_bans` 表、`IPBanService`、`IpBansView`。merge 上游时若出现同名模块，须逐行对比，**以本 Fork 实现为准**。

### 8.1 已提交能力（相对 v0.1.146）

- 数据表：`ip_bans`（迁移 `159_create_ip_bans.sql`）
- 规则类型（初版）：**IP / CIDR**（`pattern`）
- 中间件：`IPBanGuard`、`GatewayIPBanGuard` — 拦截注册与普通 API / Gateway 流量
- 管理端 CRUD：`GET/POST/PUT/DELETE /api/v1/admin/ip-bans`
- 前端：`IpBansView.vue`、`api/admin/ipBans.ts`
- Wire：`IPBanRepository` → `IPBanService` → `IPBanHandler`；`AuthService` 注入 `accessBanService`
- 客户端 IP：尊重 `X-Forwarded-For`（`ippkg.GetClientIP`）

**关键文件（新增）**

```
backend/ent/schema/ip_ban.go
backend/internal/handler/admin/ip_ban_handler.go
backend/internal/repository/ip_ban_repo.go
backend/internal/server/middleware/ip_ban_guard.go
backend/internal/service/ip_ban_service.go
backend/migrations/159_create_ip_bans.sql
frontend/src/api/admin/ipBans.ts
frontend/src/views/admin/IpBansView.vue
```

### 8.2 Fork 扩展（已提交，merge 时须保留）

在 IP 封禁基础上扩展为多类型 **Access Ban**：

| 规则类型        | `rule_type`    | 匹配逻辑                                  |
| --------------- | -------------- | ----------------------------------------- |
| IP / CIDR       | `ip`           | `ippkg.MatchesPattern(clientIP, pattern)` |
| User-Agent 子串 | `ua`           | UA contains pattern（不区分大小写）       |
| IP + UA 组合    | `ip_ua`        | 同时匹配 `pattern` + `ua_pattern`         |
| 邮箱后缀        | `email_suffix` | 注册/绑邮时匹配 `@domain` 或 `*.domain`   |

**新增/变更文件**

```
backend/internal/pkg/accessban/match.go          # 规则匹配核心（新增）
backend/internal/pkg/accessban/match_test.go
backend/internal/service/registration_email_format.go   # 注册邮箱格式强校验（新增）
backend/internal/service/registration_email_format_test.go
backend/internal/service/ip_ban_service_test.go
backend/migrations/160_extend_access_ban_rules.sql        # rule_type / ua_pattern / 索引（新增）
```

**联动修改（工作区）**

- `ip_ban_service.go` — 多规则类型 CRUD、`CheckClient` / `CheckEmail`
- `auth_service.go` — 注册/绑邮前 `ValidateRegistrationEmailFormat` + 邮箱后缀封禁
- `ip_ban_guard.go` — UA / 组合规则拦截
- `IpBansView.vue` — 规则类型筛选与创建表单
- `UsersView.vue` + `user_handler.go` — 管理端**批量删除用户**（`POST .../users/batch-delete`，跳过 admin 角色）

### 8.3 合并注意

- 迁移顺序：`159` 建表 → `160` 扩展列；**不可**被上游空迁移覆盖
- 中间件注册位置在 `routes` / `middleware` 链 — merge 后确认 `IPBanGuard` 仍在 Gateway 与 Auth 路由上
- 勿将 `IPBanService` 从 `AuthService` 构造中移除

#### 登录失败自动封禁（2026-09-05）

- `POST /api/v1/auth/login` 在验证码通过后，邮箱账号名（`@` 前，忽略大小写）包含 `admin` 且返回 `INVALID_CREDENTIALS` 时，自动写入现有 `ip_bans` 表，封禁单个公网 IP 24 小时。普通账号、验证码失败、账号已停用但密码正确、数据库/Token 服务异常不触发。
- 使用 `AuthService.LoginWithClientIP` → `IPBanService.BanFailedAdminLogin` → `IPBanRepository.UpsertAutomatic`；复用 `159`/`160` 表与唯一索引，无新增迁移。并发请求按规则唯一键去重，仍有效的自动规则不反复续期；到期后再次命中可更新到期时间。人工规则以及手动禁用的规则保持原样。
- 自动规则来源为 `admin_login_failure`，管理端 `/admin/ip-bans` 显示中文来源和原因，可编辑、禁用或删除。禁用保留人工解除意图，删除后新的失败请求可重新触发。自动写入成功立即失效本进程缓存，其余实例最迟在现有 5 秒缓存到期后生效；写入异常记录服务日志，登录仍返回原凭据错误。
- 自动创建和拦截统一使用 `GetTrustedClientIP`，不直接信任兼容模式的 `CF-Connecting-IP` / `X-Real-IP` 等任意头；现有人工封禁规则保留原有 IP 解析口径。反向代理部署必须正确配置 `server.trusted_proxies` 并由代理维护 `X-Forwarded-For`；未配置时只识别直连地址，内网、回环、链路本地等地址跳过自动封禁，避免封掉容器网关。不要将任意公网地址配置为可信代理。
- IP 封禁会影响同出口 IP 的其他用户；一次密码输错也会触发含 `admin` 账号的规则。可以在管理端禁用对应记录提前解除；该规则不会封禁账号本身。代码不会追溯扫描历史审计记录或自动封禁截图里的历史 IP。

### 8.4 订阅重置卡（Fork 新增，2026-08-16）

- 迁移 `224_create_user_subscription_reset_cards.sql` 新建永久明细表 `user_subscription_reset_cards`：记录订阅、`validity_days` 快照、来源类型/引用/序号、创建与消费时间；来源唯一键保证支付订单/兑换码幂等，同一订阅删除受 `ON DELETE RESTRICT` 保护。
- 迁移 `225_backfill_user_subscription_reset_cards.sql` 将历史 `manual_reset_credits > 0` 按所属 `group.default_validity_days`（限制在 1–36500 天）逐张回填为 `legacy_backfill` 卡；`ON CONFLICT DO NOTHING` 可重放，且不改旧计数。
- 对仍在有效固定期限内的支付购买、正数订阅兑换或**管理员分配**，不再顺延现有到期时间，而是发放一张以本次 `validity_days` 为快照的卡；支付/兑换用来源引用保证同一订单只发一次，管理员分配每次发一张。已过期订阅仍从当前时刻直接重开对应期限，不发卡。
- 消费接口为 `POST /api/v1/subscriptions/:id/reset-cards/consume`，按指定 `validity_days` 取一张未消费卡；同一事务内先锁订阅行，再新语句按幂等键重放或选卡消费，将日/周/月 USD 与 token 用量全部清零、窗口起点改为点击时刻，并放弃原剩余时间，从点击时重开该卡期限。旧 `POST .../:id/reset-daily` 保留并映射为消费 1 天卡。
- 用户 `/subscriptions` 接口返回按期限聚合的 `reset_cards.total/groups`；页面按期限展示按钮与数量，并标注“永久有效”。`manual_reset_credits` 继续作为兼容镜像：发卡加一、消费减一；新逻辑和 UI 不再以它作为卡期限的事实来源。
- 管理端分页及分组订阅列表同样批量加载 `reset_cards`，新增默认可见的重置卡数量/期限列；用户端顶部展示永久卡总数（含过期订阅持有的卡），消费后立即更新。没有卡时不展示提醒；重复加载重建汇总，避免重复累计，卡明细查询失败正常上抛。
- 订阅订单退款：未消费来源卡与订单 `REFUNDED` 同一事务作废，网关成功后才落盘；`REFUNDING` 重入根据 `REFUND_GATEWAY_SUCCEEDED` 审计或查询网关状态重试本地 finalize，不重复打退款网关。已消费来源卡仍需 force，且不缩短当前周期。

**高危同步文件（订阅重置卡）**：

```
backend/internal/service/subscription_service.go
backend/internal/service/user_subscription.go
backend/internal/service/user_subscription_port.go
backend/internal/repository/user_subscription_repo.go
backend/internal/service/payment_fulfillment.go
backend/internal/service/payment_refund.go
backend/internal/service/redeem_service.go
backend/internal/handler/subscription_handler.go
backend/internal/handler/dto/types.go
backend/internal/handler/dto/mappers.go
backend/internal/server/routes/user.go
backend/internal/server/middleware/audit_log.go
backend/migrations/224_create_user_subscription_reset_cards.sql
backend/migrations/225_backfill_user_subscription_reset_cards.sql
frontend/src/api/subscriptions.ts
frontend/src/types/index.ts
frontend/src/views/user/SubscriptionsView.vue
frontend/src/i18n/locales/zh/misc.ts
frontend/src/i18n/locales/en/misc.ts
```

#### 订阅周期额度（2026-11-14）

- 用户确认 90 天卡消费一小时后应剩余 89 天 23 小时，覆盖旧的 30 天月窗口定义。数据库/API 保留 `monthly_limit_usd`、`monthly_usage_*`、`monthly_window_start` 字段，业务语义改为当前订阅全期额度，结束时间统一为 `expires_at`；第 30/60 天不自动清零或补发。
- 模型重置时间、服务限额校验、列表归一化、自动维护、进度 API、管理端可用额度 SQL 与前端共享倒计时使用同一口径。用户端、管理端、购买/套餐配置及密钥用量页统一显示“周期额度”，到期倒计时显示额度结束，不承诺自动续期。
- 消费卡仍按购买时期限快照原子重开并清零，幂等重试不扣第二张卡、不清掉后续新增用量。管理员单条/批量重置不改到期时间；延期保持原用量和真实锚点，周期额度随新到期时间结束。旧 `align_monthly_reset` / `shift_monthly_reset` 参数继续接受（包含 0 天请求），不再写入未来锚点，界面移除该选项。
- 日额度午夜刷新、周额度滚动和钱包平台月额度保持原规则。已有订阅直接按真实到期时间显示；旧版本已自动清掉的历史用量不回填、不补扣。不新增迁移或依赖，不执行生产部署；回退版本会恢复 30 天规则，部署时应保持实例版本一致。实现及验证见 `openspec/changes/subscription-term-quota/`。

#### 订阅重置记录（2026-11-14）

- 管理端 `/admin/subscriptions` 与用户端 `/subscriptions` 增加“重置记录”入口，共享分页弹窗，显示重置时间、方式、日/周/月范围和卡期限；包含历史已消费卡及上线后的管理员单条/批量重置。自动周期刷新不计入；历史管理员操作不伪造回填，界面注明覆盖范围。
- 新增 `GET /api/v1/subscriptions/:id/reset-history` 与 `GET /api/v1/admin/subscriptions/:id/reset-history`。用户查询先校验订阅归属，其他用户与不存在订阅统一返回 404；管理端可查看已撤销订阅，受限管理 Key 需要 `subscriptions.read`。结果不包含卡来源引用、幂等键或管理员身份。
- 新迁移 `251_subscription_quota_reset_history.sql` 新建管理员记录表和查询索引。`ResetUsageWindows` 使用同一个 SQL CTE 清零指定窗口并写入记录，插入失败或批量外层事务失败时一同回滚；日窗口保持午夜锚点、周/月保持点击锚点，不改变金额/token/卡数量或期限规则。
- 重置卡历史直接读取现有 `consumed_at`，已消费后退款仍保留历史；同一幂等键重放不会多出记录。查询按时间、类型、ID 稳定倒序，分页上限 100。
- 只交付 SQL；真实数据库验证使用临时测试容器，未执行生产迁移或部署。上线前备份并先应用新增迁移；回退应用可保留表和既有记录，旧程序运行期间的管理员操作无法补记。设计与验证见 `openspec/changes/subscription-reset-history/`。

### 8.5 管理端订阅可用额度筛选（Fork 新增，2026-08-31）

- `/admin/subscriptions` 状态下拉增加 `active_available`（中文“生效中+没用完”）并设为默认筛选。
- 该组合筛选只返回 `status=active` 且未到期的订阅；日、周、周期任一已配置额度达到上限即排除，无限额订阅保留。
- 查询按日/周规则识别已进入新窗口的额度，周期额度不按 30 天恢复可用；在数据库查询阶段过滤，确保分页条目、总数与页数一致；无数据库迁移。
- 关键文件：`backend/internal/repository/user_subscription_repo.go`、`backend/internal/service/domain_constants.go`、`frontend/src/views/admin/SubscriptionsView.vue`、`frontend/src/api/admin/subscriptions.ts` 与管理端中英文 i18n。

### 8.6 余额券邮件群发（Fork 新增，2026-10-15；近期活跃扩展 2026-11-21）

参考 Dowsure LLM Router 的「兑换码 + 群发邮件」链路，但改为服务端按收件人自动生成余额券；规格见 `openspec/changes/add-inactive-user-redeem-email-campaigns/`。

- 入口：`/admin/redeem` 新增「群发余额券」向导（券设置 → 选择收件人 → 邮件内容 → 确认发送）与「发送记录」（批次进度轮询、明细、失败原因、重试失败项、删除已结束批次）。
- 筛选：以 `usage_logs` 最后一次 API 调用时间划分互斥分段 `inactive_1_2m` / `inactive_2_3m` / `inactive_3m_plus` / `never_used`，边界为服务端 UTC `now.AddDate(0, -N, 0)`；向导多选 Tab 取并集，默认选前三段、不选「从未使用」。另有互斥选项 `zero_balance`：不限最后使用时间，只包含 `users.balance` 恰好为 0 的用户（负余额不包含），不能与未活跃分段同时选择。只包含未删除、`status=active`、`role=user`、邮箱有效的用户；单批最多 1000 人。搜索结果标记已退订用户并默认不勾选。
- 近期活跃：新增 `active_7d` / `active_30d`，以最后 API 调用时间对比 UTC 基准时刻向前 7×24 / 30×24 小时（含下界）。30 天包含 7 天；可与其他时间范围多选取并集，同一用户仅返回一次。近 7 天优先标记 `active_7d`；旧未活跃范围仍按自然月，默认选项不变。`zero_balance` 仍只能单选。搜索与建批复核共享谓词，标题改为通用「群发余额券」，历史记录支持双语新范围标签。详见 `openspec/changes/add-redeem-email-active-audiences/`。
- HTTP 校验补丁（2026-11-21）：搜索与建批 DTO 的 `oneof` 白名单同步增加 `active_7d` / `active_30d`，`max=4` 调整为 `max=6`。通过真实 Gin JSON binding 测试每个公布选项、5/6 项组合及非法输入，修复此前 UI/服务已支持但 HTTP 层仍拒绝的问题。
- 迁移 `253_redeem_email_active_segments.sql` 扩展批次 CHECK 接受近期活跃范围，保留旧记录和零余额互斥；只在隔离测试库验证，不执行生产迁移。
- 迁移 `248_redeem_email_zero_balance_segment.sql` 放宽 `redeem_email_batches.segments` 检查，允许单独的 `zero_balance`；不修改 `246`。未执行迁移。
- 模板：注册为通知邮件事件 `redeem.inactive_coupon`（`category=marketing`、`Optional=true`，占位符 `coupon_value` / `redeem_code` / `expires_at` / `unsubscribe_url` + 通用变量），在 `/admin/settings` → 邮件 → 邮件模板中按 zh/en 编辑和恢复官方模板；批次不再保存模板，发送时按收件人语言取当前保存的版本。事件定义在独立文件 `internal/service/redeem_email_notification.go` 中由 `init` 追加，上游 `notification_email_service.go` 只在 `sampleVariables` 中加了一行 `notificationEmailForkSampleVariables` 钩子。设置页支持 `?tab=email&email_event=<event>` 深链（向导的「去设置中编辑模板」使用）。
- 发券：建批事务内锁定用户并按同一分段谓词复核，为每位收件人生成独立 `balance` 码（`NUMERIC(20,8)`，最多 8 位小数，备注 `redeem_email_batch:<批次ID>`），有效期 1–3650 天或长期有效（默认 30 天）。`Idempotency-Key` + 请求哈希（分段、收件人、面值、有效期）按创建人隔离：同 key 同请求回放原批次，不同请求返回 409。
- 投递：`RedeemEmailWorkerRuntime` 通过 leader 锁（Redis，失败回退 PG advisory lock）全局约 1 封/秒，经 `NotificationEmailService.Send` 发送（不写 per-recipient delivery key）；发送前复核兑换码可用性和退订状态，已退订明细记为 `skipped` / `RECIPIENT_UNSUBSCRIBED`（不计失败、不参与重试，兑换码仍保留）；租约超时标记 `DELIVERY_OUTCOME_UNKNOWN` 且不自动重发，由管理员确认后重试。
- 数据：迁移 `246_redeem_email_campaigns.sql` 只新增 `redeem_email_batches`（`segments TEXT[]`、`reference_time`、`skipped_count`）/ `redeem_email_items`（状态含 `skipped`）两张表；新表使用原生 SQL 仓储，**未新增 Ent schema**，不修改上游 `user` / `redeem_code` schema。删除批次只删发送记录，已生成的兑换码保留。
- 审计：创建与预览路由请求体整体不入审计库；动作名 `admin.redeem_email_batches.{create,retry_failed,delete}`。
- 已知取舍：不做跨批次去重，同一用户可在多个批次中各收到一张券。
- 关键文件：`backend/internal/service/redeem_email_{models,renderer,service,worker}.go`、`backend/internal/repository/redeem_email_repo.go`、`backend/internal/handler/admin/redeem_email_handler.go`、`backend/internal/server/routes/admin.go`、`backend/internal/server/middleware/audit_log.go`、Wire 接线（`cmd/server/wire.go`、`wire_gen.go`、`internal/{handler,service,repository}/wire.go`、`internal/handler/handler.go`）、`frontend/src/components/admin/redeem/`、`frontend/src/views/admin/RedeemView.vue`、`frontend/src/api/admin/redeem.ts`、`frontend/src/types/index.ts`、`frontend/src/i18n/locales/{zh,en}/fork.ts`。

### 8.7 用户端批量兑换（Fork 新增，2026-11-07）

用户端 `/redeem` 支持把多条兑换码直接粘贴进同一个输入框后一次兑换。群发邮件里的 32 位十六进制码、管理端 `XXXX-XXXX-XXXX-XXXX` 码都可以按行粘贴。

- 解析：按换行、空格、逗号和分号拆分，去掉首尾引号和行首序号/项目符号，去重并保留原大小写。单次最多 50 条。
- 提交：仍逐条调用既有 `POST /api/v1/redeem`，不新增接口，每条兑换保持原来的事务、锁和失败计数。单条「不存在 / 已使用 / 已过期」继续处理后面的码；遇到限流（429 / `REDEEM_RATE_LIMITED`）、未登录、网络错误或 5xx 时停止，剩余码标记为未提交。
- 结果：全部成功则清空输入框；有失败或未提交时，只把这些码留在输入框里，避免重试时再次提交已成功的码。离开页面后停止后续请求。
- 关键文件：`frontend/src/views/user/RedeemView.vue`、`frontend/src/views/user/redeemCodes.ts`、`frontend/src/i18n/locales/{zh,en}/dashboard.ts`。

---

### 8.8 在线 AI 生图工作台（Fork 新增，2026-11-17）

- 图片/视频统一入口（2026-11-26）：用户与管理员保留原 `/image-studio` 路由，菜单改为“AI图片/视频工作台”，Tab 切换。视频按 basketikun/infinite-canvas 新版统一创作页移植，独立 `video-studio` 模块及用户隔离 IndexedDB，包含首尾帧/全能参考、参数、会话、提示词、资产、下载与任务恢复。管理员视频配置独立保存，实际模型只开放既有 Grok/聚合视频网关；继承分组媒体开关、授权、所有者校验、审核与计费。图片业务模块不变；新增尾帧审核输入，创作内容仍不落服务端。无依赖或迁移；固定上游基线、源码映射、MIT 许可、配置和验证见 `docs/VIDEO_STUDIO_UPSTREAM.md`。

- 工作区全屏与预览分栏（2026-11-25）：工具栏提供无背景、无边框的全屏图标及右侧可见文字“全屏工作台／退出全屏”（2026-11-26 补齐），按钮自适应宽度并随工具栏换行，青绿展开／紫色收起；全屏隐藏站点 header 与侧栏并释放内容宽度，保留草稿及组件状态，Esc 优先关闭弹窗后再退出全屏。预览桌面端左侧图片与缩放，右侧导航、下载、元信息、缩略图与默认展开的提示词；信息区独立滚动，窄屏上下排列，短横屏保留切图和关闭入口。复用本地 Blob 交互，无新增依赖或请求。验证入口：`frontend/e2e/image-studio-preview.cjs`。

- 参数布局：分辨率/宽高比、尺寸/质量共享两列网格，仅剩一项填满整行；参数容器不足270px时回落单列，保留模型能力和原请求参数。

- 局部编辑修复：支持 `supports_mask` 的模型发送原图及同尺寸独立PNG遮罩，透明区域表示编辑，选区导出二值化避免擦除抗锯齿残边误提交。其他支持至少两张参考图的模型明确使用“标记引导编辑”，原图+标记图分离，不发送不支持的mask。笔刷范围与屏幕尺寸一致，擦除只清除选区；空选区/重置不请求，编辑固定一张且不覆盖普通草稿。mask保留于本地任务/历史/备份，明确重试恢复，普通生成不串用。工作台纯图标补中英文hover/focus气泡，支持鼠标移入阅读、Escape及卸载清理；BaseDialog关闭提示默认不启用，仅工作台opt-in。验证见 `openspec/changes/add-image-studio/mask-tooltip-fixes.md`。

- 交互修复（2026-11-18）：同次多图提交使用本地批次元数据，扇形叠图可展开逐张操作；单张原图下载、多选本地 ZIP，不新增依赖或服务器存储。旧历史仍逐条展示，删除记录不补位、不复活；每张仍独立同步请求及计费。已有运行/排队任务时所有新增生成入口须确认，收藏使用实色白勾；当前会话整批结束通知与普通提示五秒关闭，持久存储错误保留。工作流桌面左右独立滚动、手机单列；管理端支持当前筛选全选/全不选，保留筛选外及失效组原配置，点击保存才生效。验证见 `openspec/changes/add-image-studio/interaction-fixes.md`。

- 审查修复：收藏及分类元数据在单个 IndexedDB 事务内读取更新，不复活其他标签页已删除的记录/分类；备份导入先完成旧草稿写入，再恢复导入草稿，空草稿备份保留当前编辑。管理员可原样保留已保存的失效分组覆盖，仍拒绝新增或修改失效项。

- AI 工作流：`/image-studio/workflow-models` 自动发现已授权、可调度的文本模型，`/image-studio/workflow` 经进程内 Chat Completions 网关使用原分组账务；管理员可隐藏/命名文本模型，保留带斜杠的合法文本别名，Composite 按实际 Chat 路由目标判定能力；原生文本分组使用各自候选目录。支持需求与可选参考图创建模板、1–20条系列完整提示词，严格验证输入/JSON输出，用户先编辑审核再本地保存或明确生图。文本分组不需要生图权限；参考图能力来自已知模型/账号能力，未知不开放。取消/关闭/账号切换丢弃旧响应，无前端收费重放；OpenAI协议fallback、Anthropic SSE错误、Kiro等文本网关的回显日志及账号错误字段受隐私标记保护；审核密钥健康状态只保留稳定错误信息，原错误/冻结状态不变。

- 图片预览：上传/粘贴/拖入的参考图、已保存/未保存结果、历史参考图和素材统一使用本地 Blob 预览器；当前结果列表可连续切图，提供缩略图、适应窗口/原始尺寸、滚轮/按钮/双指缩放、拖动与键盘、实际尺寸/格式及当前图片下载。图片加载失败可重新读取；切图复位，关闭归还焦点，素材预览不触发加入参考图。短横屏隐藏缩略图与缩放滑杆，保留缩放按钮和前后导航。无新增上传、收费请求、依赖或存储格式变化。

- 低高度交互：原生成按钮被裁切时自动显示固定底部操作条，复用禁用/提交逻辑，支持回到参数、侧栏宽度、安全区与可视视口；弹窗/移动导航优先。等待卡片以取景框、扫描与片层聚合表达创作，区分真实排队/运行状态，低动态使用静态替代，离屏/后台暂停；仅当前会话完成的作品播放一次显影。配置流程见 [`IMAGE_STUDIO_SETUP.md`](IMAGE_STUDIO_SETUP.md)。

- 入口：用户 `/image-studio`、管理员 `/admin/image-studio`。参考 `tigerowo/infinite-canvas` 提交 `dc1f1d7e741eaae24c7c11a8de8892f36ed01b49` 的生图工作台，提供侧边/底部输入布局、参考图（支持 Ctrl/Cmd+V 直接粘贴、聚焦参考区及拖入高亮；普通文字/弹窗粘贴不受影响，连续添加仍受模型数量限制）、独立并发任务、预览下载、收藏/多分类/搜索/批量操作、素材、本地模板变量与系列生成、蓝色标记局部编辑。
- 零用户配置：目录复用现有分组授权、模型 allowlist、账号映射和 composite 路由。用户选模型即可；同名模型按分组区分。管理员可隐藏、改名或声明既有别名的图片协议，覆盖项存入既有 settings；不改变公共模型 API。
- 计费：服务器每次提交重验权限，进程内调用原同步网关，继承并发、余额/订阅配额、图片/token 定价和用户倍率。前端展示有效单张估价，实际用量仍以原网关结算为准。客户端不自动重放收费请求，刷新将未完成任务标记中断；主动重试可能再次计费。
- 内部身份：按用户/分组使用 `internal:images:` 前缀的真实 API Key 账务行，由 JWT secret HMAC 派生，唯一约束处理首次并发。只允许带私有上下文的进程内认证；不占用户密钥计数，不在列表/搜索/DTO 中暴露密钥，禁止通过普通 API 修改或删除。无需新增表或迁移。
- 本地内容：数据库 `sub2api-image-studio-{userID}` 使用原生 IndexedDB，历史仅来自可解码的图片 Blob；同一事务写入历史并移除任务，成功后才显示为已保存。聚焦/刷新/跨标签变化重读数据库，外部删除不从内存复活。存储失败提供未保存提示、下载及重存。支持本地 JSON 备份与原子导入；切换账号中止旧请求并隔离结果。Web Locks 标识活跃任务所属标签页，避免误判中断；不支持时提示未追踪并保留手工确认重试，不改写共享运行状态。
- 服务端隐私：仅既有身份、配置和账务/安全计数元数据可持久化，不接异步生图/对象存储。`NoContentPersistence` 保留审核判定和阻断，关闭提示词审核内容入队、Ops 内容捕获、面板请求体审计及上游回显日志/账号错误持久化。Grok/Images URL 结果经既有受限抓取转内联，响应 `Cache-Control: no-store`。
- 主要文件：`backend/internal/{handler,service}/image_studio*.go`、`backend/internal/pkg/ctxkey/privacy.go`、网关/审核/日志接线；`frontend/src/features/image-studio/`、用户与管理页面、路由导航及中英文文案。规划和验收证据：`openspec/changes/add-image-studio/`。浏览器回归：`frontend/e2e/image-studio-{storage,ui,tabs}.cjs`，使用本地 HTTP fixtures，不请求生产模型。

---

## 9. 迁移红线（必守）

`user_platform_quotas.platform` 的 CHECK 约束会被多次 DROP + 重建：

```
142（上游）→ anthropic, openai, gemini, antigravity
145（Fork）→ + kiro
157（上游原版）→ 仅 + grok → ⚠️ 会丢掉 kiro
157（本 Fork 已修复）→ anthropic, openai, gemini, antigravity, kiro, grok
```

**今后 merge 上游若覆盖 `157`，必须重新确认约束含 `kiro`。**

v0.2.4 新增 `237_add_minimax_platform.sql` 再次重建配额和 Composite 路由约束：最终平台集合为 anthropic/openai/gemini/antigravity/grok/kiro/kimi/zhipu/deepseek/minimax。237 已补回 Kiro；监控 provider 只按官方范围增加 MiniMax，历史 295 个 SQL 不改。部署前备份并验证实际 Kiro 行及约束；本次不执行真实迁移。

---

v0.2.5 新增 `238_opencode_go_platform.sql`：配额/Composite CHECK 增加 OpenCode 并保留 Kiro 和 MiniMax，monitor 不新增 Kiro。另一迁移 `238_purge_unlimited_user_platform_quotas.sql` 删除三档限额全 NULL 的行；任何非 NULL（含零）限额仍保留。执行前备份历史行，验证清理耗时与数据兼容性；删除后不能仅靠二进制降级恢复。本次仅交付 SQL，不执行迁移。

v0.2.12–v0.2.14 新增两个 `241` 迁移：订单 `bonus_amount DECIMAL(20,2) NOT NULL DEFAULT 0`，以及 TypeSafe 平台配额/Composite 约束（保留 Kiro）。旧订单赠送额默认 0，订阅不参加优惠，充值返利剔除免费 USD 部分。密码重置旧链接升级后需重新申请；新装管理员凭据校验不影响已有管理员。未执行迁移/部署；上线前备份并核验新平台数据及优惠订单，不能假设直接降级旧程序即可回退。

## 10. 合并冲突高危文件（上游常改 + Fork 已改）

改动或 merge 前**必须先读现有实现**：

| 文件                                                                                                                 | Fork 改动性质                                                     |
| -------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| `backend/cmd/server/wire.go` / `wire_gen.go`                                                                         | Kiro + Grok TokenProvider / UsageService 注入                     |
| `backend/internal/domain/constants.go`                                                                               | `PlatformKiro`, `PlatformGrok`                                    |
| `backend/internal/handler/admin/account_handler.go`                                                                  | Kiro + Grok 可用模型分支                                          |
| `backend/internal/service/account_test_service.go`                                                                   | struct + 构造含 kiro/grok                                         |
| `backend/internal/service/account_usage_service.go`                                                                  | Provide 版注入 Kiro                                               |
| `backend/internal/payment/types.go` / `provider/factory.go`                                                          | `TypeXorPay`                                                      |
| `backend/migrations/157_user_platform_quotas_add_grok.sql`                                                           | CHECK 含 kiro+grok                                                |
| `backend/internal/i18n/*` + `frontend/src/i18n/locales/{zh,en}.ts`                                                   | Kiro / XorPay 文案                                                |
| `frontend/src/types/index.ts` / `types/payment.ts`                                                                   | Kiro / XorPay 类型                                                |
| `frontend/src/views/user/PaymentView.vue`                                                                            | 倍率隐藏 + 支付金额 UX                                            |
| `frontend/src/components/common/VersionBadge.vue`                                                                    | 隐藏在线更新按钮                                                  |
| `backend/internal/service/ip_ban_service.go`                                                                         | Access Ban 多规则类型                                             |
| `backend/internal/server/middleware/ip_ban_guard.go`                                                                 | Gateway/Auth 封禁中间件                                           |
| `backend/internal/service/auth_service.go`                                                                           | 注册邮箱校验 + 邮箱后缀封禁                                       |
| `backend/migrations/159_create_ip_bans.sql` / `160_extend_access_ban_rules.sql`                                      | Access Ban 表结构                                                 |
| `frontend/src/views/admin/IpBansView.vue`                                                                            | 封禁规则管理 UI                                                   |
| `backend/internal/server/routes/gateway.go`                                                                          | Access Ban 中间件；官方新增 `/x_search` 须同步挂 `ipBanAnthropic` |
| `backend/internal/handler/admin/group_handler.go`                                                                    | 分组 `oneof` 含 `kiro`；合入官方字段时不可丢掉                    |
| `backend/internal/service/subscription_service.go` / `user_subscription_port.go`                                     | 固定期限重复购买发卡、过期重开、按期限消费与兼容镜像              |
| `backend/internal/repository/user_subscription_repo.go`                                                              | 重置卡幂等发放、行锁消费、用量/窗口原子重开与聚合查询             |
| `backend/internal/service/payment_fulfillment.go` / `redeem_service.go`                                              | 支付订单与正数订阅兑换的来源引用、`validity_days` 快照            |
| `backend/internal/service/payment_refund.go`                                                                         | 未消费来源卡与 REFUNDED 同事务作废；REFUNDING 重入不重复打网关    |
| `backend/internal/handler/subscription_handler.go` / `server/routes/user.go`                                         | 新消费接口及旧 `reset-daily` → 1 天卡兼容映射                     |
| `backend/migrations/224_create_user_subscription_reset_cards.sql` / `225_backfill_user_subscription_reset_cards.sql` | 永久卡明细、历史兼容计数按分组默认期限回填                        |
| `frontend/src/views/user/SubscriptionsView.vue` / `api/subscriptions.ts` / `types/index.ts`                          | `/subscriptions` 按期限展示永久重置卡与消费动作                   |

---

## 11. 构建与部署差异

| 文件                         | Fork 改动                                                             |
| ---------------------------- | --------------------------------------------------------------------- |
| `Dockerfile`                 | `NPM_CONFIG_REGISTRY` 默认 npmmirror；`pnpm install --ignore-scripts` |
| `.dockerignore`              | 保留 `docs/`、`docs/legal/` 供前端合规页构建                          |
| `deploy/docker-compose*.yml` | 与 Fork 环境相关的挂载/标签（含 SELinux `:Z` 等历史修复）             |
| `deploy/.env.example`        | Kiro 相关 env 示例                                                    |

---

## 12. 完整新增文件清单（85，已提交）

<details>
<summary>backend（点击展开）</summary>

```
backend/internal/handler/admin/group_handler_kiro_validation_test.go
backend/internal/handler/admin/kiro_oauth_handler.go
backend/internal/payment/provider/xorpay.go
backend/internal/pkg/anthropictokenizer/NOTICE.md
backend/internal/pkg/anthropictokenizer/claude.json
backend/internal/pkg/anthropictokenizer/tokenizer.go
backend/internal/pkg/anthropictokenizer/tokenizer_test.go
backend/internal/pkg/kiro/fingerprint.go
backend/internal/pkg/kiro/fingerprint_test.go
backend/internal/pkg/kiro/models.go
backend/internal/pkg/kiro/models_test.go
backend/internal/pkg/kiro/oauth.go
backend/internal/pkg/kiro/oauth_invalid_grant_test.go
backend/internal/pkg/kiro/oauth_test.go
backend/internal/pkg/kiro/signature.go
backend/internal/pkg/kiro/translator.go
backend/internal/pkg/kiro/translator_test.go
backend/internal/pkg/kiro/websearch.go
backend/internal/pkg/kiro/websearch_stream.go
backend/internal/pkg/kiro/websearch_stream_test.go
backend/internal/pkg/kiro/websearch_test.go
backend/internal/pkg/kirocooldown/store.go
backend/internal/pkg/kirocooldown/store_test.go
backend/internal/service/account_kiro_credit_unit_price_test.go
backend/internal/service/account_test_service_kiro_apikey_fallback_test.go
backend/internal/service/account_test_service_kiro_test.go
backend/internal/service/account_usage_service_kiro_apikey_test.go
backend/internal/service/kiro_alignment_test.go
backend/internal/service/kiro_alignment_unit_test.go
backend/internal/service/kiro_cache_emulation.go
backend/internal/service/kiro_cache_emulation_test.go
backend/internal/service/kiro_credits_usage_log_test.go
backend/internal/service/kiro_error_classifier.go
backend/internal/service/kiro_error_classifier_test.go
backend/internal/service/kiro_http_helpers.go
backend/internal/service/kiro_http_helpers_test.go
backend/internal/service/kiro_mapping_fallback_test.go
backend/internal/service/kiro_oauth_service.go
backend/internal/service/kiro_oauth_service_test.go
backend/internal/service/kiro_runtime.go
backend/internal/service/kiro_runtime_state.go
backend/internal/service/kiro_runtime_state_integration_test.go
backend/internal/service/kiro_runtime_state_test.go
backend/internal/service/kiro_session_test.go
backend/internal/service/kiro_sticky_session_test.go
backend/internal/service/kiro_token_provider.go
backend/internal/service/kiro_token_provider_test.go
backend/internal/service/kiro_token_refresher.go
backend/internal/service/kiro_usage_fetcher.go
backend/internal/service/kiro_websearch.go
backend/internal/service/kiro_websearch_test.go
backend/migrations/135_add_group_kiro_cache_emulation.sql
backend/migrations/145_allow_kiro_user_platform_quotas.sql
backend/migrations/151_add_group_kiro_auto_sticky.sql
backend/migrations/152_add_group_kiro_sticky_session_ttl.sql
backend/migrations/153_add_group_kiro_endpoint_mode.sql
backend/migrations/153_add_usage_log_kiro_credits.sql
backend/migrations/159_create_ip_bans.sql
backend/ent/schema/ip_ban.go
backend/internal/handler/admin/ip_ban_handler.go
backend/internal/repository/ip_ban_repo.go
backend/internal/server/middleware/ip_ban_guard.go
backend/internal/server/middleware/ip_ban_guard_test.go
backend/internal/service/ip_ban_service.go
```

</details>

<details>
<summary>frontend（点击展开）</summary>

```
frontend/public/logo.svg
frontend/src/api/admin/ipBans.ts
frontend/src/api/admin/kiro.ts
frontend/src/components/account/__tests__/AccountTodayStatsCell.spec.ts
frontend/src/components/account/__tests__/OAuthAuthorizationFlow.spec.ts
frontend/src/components/common/__tests__/GroupBadge.spec.ts
frontend/src/components/common/__tests__/PlatformIcon.spec.ts
frontend/src/components/common/__tests__/PlatformTypeBadge.spec.ts
frontend/src/components/payment/__tests__/AmountInput.spec.ts
frontend/src/composables/useKiroOAuth.ts
frontend/src/utils/__tests__/platformColors.spec.ts
frontend/src/utils/__tests__/subscriptionPlanValidity.spec.ts
frontend/src/utils/subscriptionPlanValidity.ts
frontend/src/views/admin/IpBansView.vue
frontend/src/views/admin/__tests__/BackupView.spec.ts
frontend/src/views/admin/settings/__tests__/EmailTemplateEditor.spec.ts
```

</details>

<details>
<summary>deploy / 根目录（点击展开）</summary>

```
FORK_CUSTOMIZATIONS.md
assets/community/wechat-group.jpg
assets/screenshots/kiro-account-management.png
assets/screenshots/kiro-add-account.png
assets/screenshots/kiro-cache-emulation.png
```

</details>

<details>
<summary>Fork 扩展新增（7）</summary>

```
backend/internal/pkg/accessban/match.go
backend/internal/pkg/accessban/match_test.go
backend/internal/service/ip_ban_service_test.go
backend/internal/service/registration_email_format.go
backend/internal/service/registration_email_format_test.go
backend/migrations/160_extend_access_ban_rules.sql
frontend/src/components/common/__tests__/Select.spec.ts
```

</details>

---

## 13. 完整修改文件清单（215，已提交）

按目录分组；生成命令：`git diff v0.1.146..HEAD --name-only --diff-filter=M`

<details>
<summary>backend（点击展开）</summary>

```
backend/cmd/server/VERSION
backend/cmd/server/main.go
backend/cmd/server/wire.go
backend/cmd/server/wire_gen.go
backend/cmd/server/wire_gen_test.go
backend/ent/group.go
backend/ent/group/group.go
backend/ent/group/where.go
backend/ent/group_create.go
backend/ent/group_update.go
backend/ent/migrate/schema.go
backend/ent/mutation.go
backend/ent/runtime/runtime.go
backend/ent/schema/group.go
backend/ent/schema/user_platform_quota.go
backend/go.mod
backend/go.sum
backend/internal/config/config.go
backend/internal/config/config_test.go
backend/internal/domain/constants.go
backend/internal/domain/constants_test.go
backend/internal/handler/admin/account_codex_import_test.go
backend/internal/handler/admin/account_data.go
backend/internal/handler/admin/account_data_handler_test.go
backend/internal/handler/admin/account_handler.go
backend/internal/handler/admin/account_handler_available_models_test.go
backend/internal/handler/admin/account_handler_list_test.go
backend/internal/handler/admin/account_handler_mixed_channel_test.go
backend/internal/handler/admin/account_handler_passthrough_test.go
backend/internal/handler/admin/batch_update_credentials_test.go
backend/internal/handler/admin/group_handler.go
backend/internal/handler/admin/ops_handler.go
backend/internal/handler/admin/user_platform_quota_admin_test.go
backend/internal/handler/dto/mappers.go
backend/internal/handler/dto/types.go
backend/internal/handler/gateway_handler.go
backend/internal/handler/gateway_handler_chat_completions.go
backend/internal/handler/gateway_handler_responses.go
backend/internal/handler/gateway_handler_warmup_intercept_unit_test.go
backend/internal/handler/gateway_models_test.go
backend/internal/handler/handler.go
backend/internal/handler/payment_webhook_handler.go
backend/internal/handler/wire.go
backend/internal/model/error_passthrough_rule.go
backend/internal/payment/provider/factory.go
backend/internal/payment/types.go
backend/internal/pkg/usagestats/account_stats.go
backend/internal/repository/account_repo.go
backend/internal/repository/account_repo_temp_unsched_test.go
backend/internal/repository/api_key_repo.go
backend/internal/repository/group_repo.go
backend/internal/repository/ops_repo.go
backend/internal/repository/ops_repo_error_where_test.go
backend/internal/repository/simple_mode_default_groups.go
backend/internal/repository/usage_log_repo.go
backend/internal/repository/usage_log_repo_integration_test.go
backend/internal/repository/usage_log_repo_request_type_test.go
backend/internal/server/api_contract_test.go
backend/internal/server/routes/admin.go
backend/internal/server/routes/payment.go
backend/internal/service/account.go
backend/internal/service/account_service.go
backend/internal/service/account_test_service.go
backend/internal/service/account_usage_service.go
backend/internal/service/account_usage_service_test.go
backend/internal/service/account_wildcard_test.go
backend/internal/service/admin_service.go
backend/internal/service/api_key_auth_cache.go
backend/internal/service/api_key_auth_cache_impl.go
backend/internal/service/billing_service.go
backend/internal/service/domain_constants.go
backend/internal/service/gateway_forward_as_chat_completions.go
backend/internal/service/gateway_forward_as_responses.go
backend/internal/service/gateway_record_usage_test.go
backend/internal/service/gateway_request.go
backend/internal/service/gateway_service.go
backend/internal/service/gateway_streaming_test.go
backend/internal/service/gateway_websearch_emulation.go
backend/internal/service/gateway_websearch_emulation_test.go
backend/internal/service/group.go
backend/internal/service/group_test.go
backend/internal/service/openai_privacy_retry_test.go
backend/internal/service/ops_port.go
backend/internal/service/ops_repo_mock_test.go
backend/internal/service/ops_service.go
backend/internal/service/ops_upstream_context.go
backend/internal/service/payment_config_providers.go
backend/internal/service/payment_order_lifecycle.go
backend/internal/service/pricing_service.go
backend/internal/service/ratelimit_service.go
backend/internal/service/ratelimit_service_401_test.go
backend/internal/service/scheduler_snapshot_service.go
backend/internal/service/token_cache_invalidator.go
backend/internal/service/token_refresh_service.go
backend/internal/service/token_refresh_service_test.go
backend/internal/service/token_refresher_test.go
backend/internal/service/usage_log.go
backend/internal/service/usage_log_helpers.go
backend/internal/service/wire.go
backend/migrations/157_user_platform_quotas_add_grok.sql
backend/scripts/resolve-version.sh
```

</details>

<details>
<summary>frontend（点击展开）</summary>

```
frontend/src/api/__tests__/settings.authSourceDefaults.spec.ts
frontend/src/api/admin/accounts.ts
frontend/src/api/admin/index.ts
frontend/src/api/admin/ops.ts
frontend/src/api/admin/settings.ts
frontend/src/api/admin/users.ts
frontend/src/components/account/AccountStatusIndicator.vue
frontend/src/components/account/AccountTodayStatsCell.vue
frontend/src/components/account/AccountUsageCell.vue
frontend/src/components/account/BulkEditAccountModal.vue
frontend/src/components/account/CreateAccountModal.vue
frontend/src/components/account/EditAccountModal.vue
frontend/src/components/account/OAuthAuthorizationFlow.vue
frontend/src/components/account/QuotaDimensionRow.vue
frontend/src/components/account/QuotaNotifyToggle.vue
frontend/src/components/account/__tests__/AccountStatusIndicator.spec.ts
frontend/src/components/account/__tests__/AccountUsageCell.spec.ts
frontend/src/components/account/__tests__/EditAccountModal.spec.ts
frontend/src/components/admin/ErrorPassthroughRulesModal.vue
frontend/src/components/admin/account/AccountTableFilters.vue
frontend/src/components/admin/account/ReAuthAccountModal.vue
frontend/src/components/admin/payment/AdminOrderTable.vue
frontend/src/components/admin/payment/__tests__/orderCurrencyDisplay.spec.ts
frontend/src/components/admin/usage/UsageStatsCards.vue
frontend/src/components/admin/usage/UsageTable.vue
frontend/src/components/admin/usage/__tests__/UsageStatsCards.spec.ts
frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts
frontend/src/components/admin/user/UserPlatformQuotaModal.vue
frontend/src/components/admin/user/__tests__/UserPlatformQuotaModal.spec.ts
frontend/src/components/channels/AvailableChannelsTable.vue
frontend/src/components/charts/EndpointDistributionChart.vue
frontend/src/components/charts/GroupDistributionChart.vue
frontend/src/components/charts/ModelDistributionChart.vue
frontend/src/components/charts/UserBreakdownSubTable.vue
frontend/src/components/charts/__tests__/GroupDistributionChart.spec.ts
frontend/src/components/charts/__tests__/ModelDistributionChart.spec.ts
frontend/src/components/common/GroupBadge.vue
frontend/src/components/common/GroupOptionItem.vue
frontend/src/components/common/PlatformIcon.vue
frontend/src/components/common/PlatformTypeBadge.vue
frontend/src/components/common/Select.vue
frontend/src/components/common/VersionBadge.vue
frontend/src/components/layout/AuthLayout.vue
frontend/src/components/payment/AmountInput.vue
frontend/src/components/payment/OrderTable.vue
frontend/src/components/payment/PaymentMethodSelector.vue
frontend/src/components/payment/PaymentQRDialog.vue
frontend/src/components/payment/PaymentStatusPanel.vue
frontend/src/components/payment/ProviderCard.vue
frontend/src/components/payment/SubscriptionPlanCard.vue
frontend/src/components/payment/__tests__/PaymentStatusPanel.spec.ts
frontend/src/components/payment/paymentFlow.ts
frontend/src/components/payment/providerConfig.ts
frontend/src/components/user/PlatformUsageBreakdown.vue
frontend/src/components/user/UserPlatformQuotaCell.vue
frontend/src/components/user/dashboard/UserDashboardCharts.vue
frontend/src/components/user/dashboard/UserDashboardRecentUsage.vue
frontend/src/components/user/dashboard/UserDashboardStats.vue
frontend/src/composables/__tests__/useModelWhitelist.spec.ts
frontend/src/composables/useModelWhitelist.ts
frontend/src/constants/account.ts
frontend/src/i18n/locales/en.ts
frontend/src/i18n/locales/zh.ts
frontend/src/types/index.ts
frontend/src/types/payment.ts
frontend/src/utils/platformColors.ts
frontend/src/views/HomeView.vue
frontend/src/views/admin/AccountsView.vue
frontend/src/views/admin/BackupView.vue
frontend/src/views/admin/ChannelsView.vue
frontend/src/views/admin/GroupsView.vue
frontend/src/views/admin/ProxiesView.vue
frontend/src/views/admin/RiskControlView.vue
frontend/src/views/admin/SettingsView.vue
frontend/src/views/admin/SubscriptionsView.vue
frontend/src/views/admin/UsageView.vue
frontend/src/views/admin/UsersView.vue
frontend/src/views/admin/__tests__/SettingsView.spec.ts
frontend/src/views/admin/__tests__/UsageView.spec.ts
frontend/src/views/admin/__tests__/UsersView.spec.ts
frontend/src/views/admin/__tests__/groupsModelsListLayout.spec.ts
frontend/src/views/admin/ops/components/OpsDashboardHeader.vue
frontend/src/views/admin/ops/components/OpsSystemLogTable.vue
frontend/src/views/admin/orders/AdminOrdersView.vue
frontend/src/views/admin/settings/EmailTemplateEditor.vue
frontend/src/views/user/AvailableChannelsView.vue
frontend/src/views/user/KeysView.vue
frontend/src/views/user/PaymentQRCodeView.vue
frontend/src/views/user/PaymentView.vue
frontend/src/views/user/SubscriptionsView.vue
frontend/src/views/user/UsageView.vue
frontend/src/views/user/UserOrdersView.vue
frontend/src/views/user/__tests__/PaymentView.spec.ts
frontend/src/views/user/__tests__/UsageView.spec.ts
frontend/src/views/user/__tests__/paymentUx.spec.ts
frontend/src/views/user/paymentUx.ts
```

</details>

<details>
<summary>deploy / 根目录（点击展开）</summary>

```
.dockerignore
.gitignore
Dockerfile
deploy/.env.example
deploy/README.md
deploy/config.example.yaml
deploy/docker-compose.local.yml
deploy/docker-compose.yml
```

</details>

---

## 14. 与官方同步新版本（标准流程）

1. `git fetch upstream tag vX.Y.Z`
2. `git merge vX.Y.Z`（建议 `--no-commit` 先解决冲突）
3. **冲突原则**：Kiro / XorPay / Access Ban / Grok 定制**全部保留**；wire 注入顺序 **kiro 在前、grok 在后**
4. **必查**：`157` 迁移 CHECK、`159`/`160` Access Ban、`224`/`225` 重置卡迁移、`wire_gen.go`、`account_handler.go`、支付/兑换来源传递、订阅服务/仓储/路由、`SubscriptionsView.vue`、`VersionBadge.vue`、`auth_service.go`
5. 验证：
   - 后端：`go build ./... && go vet ./... && go test -tags=unit ./internal/...`
   - 前端：`pnpm install --frozen-lockfile && pnpm vitest run && pnpm typecheck`
6. 更新 `backend/cmd/server/VERSION`、本文档统计与 `AGENTS.md` 摘要
7. UTF-8 乱码复核

---

## 15. AI / 开发者禁区

- **不要**删除或弱化 Kiro / XorPay / **Access Ban** 模块以「与上游对齐」
- **不要**恢复 `VersionBadge` 的「立即更新」按钮（Fork 禁用在线升级）
- **不要**把迁移 `157` 改回仅含 `grok`
- **不要**删除 `159`/`160` 迁移或回退为仅 IP 封禁
- **不要**删除 `224`/`225` 重置卡迁移、恢复有效固定期限重复购买顺延，或把 `manual_reset_credits` 当作卡期限事实来源
- **不要**在 merge 时用上游版本覆盖 `ProvideAccountUsageService` 而丢掉 Kiro 注入
- **不要**从 `AuthService` 移除 `accessBanService` 或注册邮箱格式校验
- **不要**猜测 Kiro credits 或 cache_read 计费口径；见第 3 节
- 修改支付/配额/迁移/SQL 时，同步检查 `dowalet_db_log` 或本项目 SQL 变更约定（若适用）

---

## 工作台界面整改补充（2026-11-26）

用户端 `/image-studio` 合并顶部工具栏，修复视频提交/查询交接闪动、轮询强制滚动和参数浮层外部点击；结果卡片与底部输入区分开布局。图片/视频共用精选提示词弹窗，沿用 infinite-canvas 七个公开来源、分类/标签/搜索、一键填入，原有自定义和工作流执行保留。公开缓存独立 IndexedDB，不改变私有创作数据隔离、图片生成、后端授权和计费。来源与维护说明见 `VIDEO_STUDIO_UPSTREAM.md`。

## 16. 相关文档

| 文档                                                                   | 说明                                             |
| ---------------------------------------------------------------------- | ------------------------------------------------ |
| [PAYMENT_CN.md](./PAYMENT_CN.md)                                       | 支付配置（官方文档；XorPay 配置见代码与第 4 节） |
| [ADMIN_PAYMENT_INTEGRATION_API.md](./ADMIN_PAYMENT_INTEGRATION_API.md) | 外部支付对接 Admin API                           |
| [legal/admin-compliance.zh.md](./legal/admin-compliance.zh.md)         | 管理端合规文案                                   |

---

_最后更新：2026-11-24 · 基准：官方 v0.2.25 · 增量清单见 `FORK_CUSTOMIZATIONS.md` · 工作区：`git status --short`_
