# 在线 AI 生图工作台验收记录

## 2026-11-17 AI 工作流补齐与最终验收

此节记录用户再次授权“全面完善之后 push”后的最终补齐工作；下方各节保留此前阶段的真实结果，其“未实现/未推送”描述属于当时状态。

已补齐原参考工作台的 AI 创建工作流和 AI 系列提示词。用户自动获得已授权文本模型，无需配置 URL 或密钥；可选择当前参考图（仅已确认视觉能力的模型）。AI 输出先进入可编辑草稿，明确保存才写入 IndexedDB，明确生图才进入原图片队列。取消、关闭、账号切换会中止并丢弃旧响应；提交前提示文本分组计费与取消可能已计费。管理员可隐藏/重命名文本模型。配置全流程见 `docs/IMAGE_STUDIO_SETUP.md`。

后端复用进程内 `/v1/chat/completions`、原分组账务身份与授权，逐次重验模型。输入和 JSON 输出有结构、数量、长度限制；拒绝远程参考图和未知字段。文本模型错误日志、审核与审计同样遵守 `NoContentPersistence`，仅保留账务与安全状态元数据。最终审查发现工作流请求体在 handler 执行前被审计读取，已补路由级正文及查询参数排除，并验证成功和提前拒绝两条路径。

最终审查还用真实失败测试复现并修复：审核供应商回显内容进入密钥健康状态、原生文本分组误用 Claude 默认目录、带斜杠文本别名被排除、Composite 显式路由别名未解析实际目标。审核失败改为稳定非空诊断，保留原 error/frozen 状态；目录修复使用已有平台模型与 Chat 路由 resolver，继续按公开名字授权和隐藏，图片路径字符限制保持有效。

本轮验证（分别在 frontend/backend 根目录）：

| 命令 | 结果 |
| --- | --- |
| `pnpm exec vitest run` | **428 文件 / 3483 项通过**。 |
| `pnpm exec vitest run src/features/image-studio src/i18n/__tests__/localeKeyCompleteness.spec.ts` | **16 文件 / 98 项通过**，包含 AI 交互、目录配置和本地存储回归。 |
| `pnpm typecheck`、`pnpm run lint:check` | 均通过。 |
| `pnpm run build` | 通过；最终计费提示文案修改后重跑成功。保留既有 Browserslist、混合导入及大 chunk 警告。 |
| `go test -tags=unit ./internal/...` | 通过，**57 个有测试的包**；最终审查补丁另跑受影响包回归。 |
| `go vet ./...` | 通过；最终后端补丁另做编译与 vet。 |
| `go test -tags=unit ./internal/handler ./internal/service ./internal/server/middleware -count=1` | 最终修复波完整回归通过：handler 66.473s、service 248.427s、middleware 1.142s，退出 0。 |
| `go build ./...`、`go vet ./internal/handler ./internal/service ./internal/server/routes ./internal/server/middleware` | 最后代码编辑后均退出 0。 |
| `go test -tags=unit ./internal/server/middleware -count=1` | 审计修复后完整中间件包通过。独立审查另跑 generate/workflow × 200/403 四场景通过。 |
| `node e2e/image-studio-{storage,ui,tabs,preview,motion,workflow}.cjs`（逐一执行） | **六个脚本全部通过**，使用宿主已有 Playwright 与系统 Chrome。 |
| `impeccable detect --json --changed` | `[]`。 |

浏览器新增验收包含：AI 模板审核与本地保存刷新、参考图、系列提示词编辑后明确提交两次生图、取消/失败不重放、空目录、手机深色布局；管理端文本模型隐藏、重命名、刷新和无图片协议字段。原存储、跨标签、参考图粘贴拖拽、统一预览、低高度悬浮入口与等待动效回归通过。视觉复核桌面创建、系列编辑、手机深色与等待界面；截图禁用过渡动画后清晰完整。所有上游均为 HTTP fixtures，无真实收费调用。

最终独立复核结论：规格 PASS、代码质量 PASS，五项发现全部 ADDRESSED，无遗留阻塞。全部 112 个新增/修改文件已检查 UTF-8、BOM、U+FFFD 与秘密模式；58 个 Go 文件 gofmt 无遗漏，`git diff --check` 通过。配置及规划 Markdown 本地链接有效。完整受影响包测试通过后，按用户授权提交并推送功能分支，远端 SHA 核对由最终交付输出记录。

验证边界：未连接生产数据库或真实模型供应商，未实测真实资金结算；未验证实体手机软键盘、Safari 或真实后台页面暂停。OpenSpec CLI 缺失，按现有结构手工维护规划，不安装依赖。没有新增依赖或数据库迁移，不执行部署。

流程决定：按用户最新明确授权完成后推送，覆盖早期“保留未提交”的阶段要求；只推送功能分支，不合并或部署。OpenSpec 使用人工结构核对，代价是缺少 CLI 校验证据。审查快照包含全部未跟踪文件，避免仅审查 HEAD diff 漏掉新增实现；最终以待提交文件清单再次核对，降低快照遗漏风险。

## 2026-11-17 推送前审查与修复

审查基线：`b8b45c8fe520e24d3ae26fee29d2e1486137391c`，包含本功能所有未跟踪文件。独立后端与需求审查完成；用户授权仅在审查无问题后提交推送，当前仍有下述功能缺口，未提交或推送。

已修复三个问题，并保留先红后绿的回归：

- `storage.ts` / `useImageStudio.ts`：收藏与多分类改为单个 IndexedDB 事务内读取和更新，只修改仍存在的记录，并过滤已删除分类；不再把先前读取的完整历史写回。真实 Chrome 双连接验证删除后不复活、并发两次收藏原子反转和分类删除一致性。
- `useImageStudio.ts`：导入前取消待发自动保存并等待已开始的草稿写入，恢复本次导入的草稿/参考/参数/布局；不含草稿的备份保留当前编辑，导入失败恢复自动保存，账号切换后忽略旧结果与旧写入错误。新增 6 项回归。
- `image_studio.go`：允许原样保留已保存的失效分组覆盖或删除它们，仍拒绝新增/修改失效项；旧设置读取失败不放行。四种保存情形已回归；配置指南和 Fork 摘要已同步。

未解决的 P1 需求缺口：参考项目生图页面嵌入 `CreativeWorkflowWorkspace`，支持 `runWorkflowAgent` 创建工作流和 `generateSeriesPromptDrafts` 调用文本模型生成系列提示词；当前 `WorkflowLibrary.vue:54` 仅拼接手工模板。用户“完全按照参考工作台”的要求没有豁免此能力，不能认定完整交付或据此推送。此处没有用改写原需求消除缺口。

本轮新鲜检查（分别在 frontend/backend 根目录）：

| 命令 | 结果 |
| --- | --- |
| `pnpm exec vitest run` | 426 文件、3471 项全部通过。 |
| `pnpm typecheck` | 通过（导入草稿修复后）。 |
| `pnpm run lint:check` | 全量通过。 |
| `pnpm run build` | i18n、vue-tsc 和 Vite 通过；保留既有 Browserslist、混用静态/动态导入和大 chunk 警告。 |
| `go test -tags=unit ./internal/handler ./internal/service ./internal/handler/dto ./internal/server/middleware ./internal/securityaudit -run 'Test.*(ImageStudio\|Ephemeral\|ImagesB64\|ImageURL\|Backfill\|RequestLogger)' -count=1` | 五包通过。 |
| `go build ./...` / `go vet ./...` | 均通过。 |
| `node e2e/image-studio-storage.cjs` | 12 项全部 true，含三项新增跨连接元数据检查。 |
| `node e2e/image-studio-ui.cjs` | 7 次 fixture 生成；参考输入、局部编辑、模板系列、分类、刷新不重放等通过，无页面错误。 |
| `node e2e/image-studio-tabs.cjs` | 8 项全部 true；Web Locks 活跃页保护、完成同步、关闭恢复、失败删除同步和无 Locks 降级均通过。悬浮入口新增后旧选择器会匹配两个按钮，脚本已按现有入口顺序选择可交互按钮，复跑通过。 |

浏览器脚本使用已有宿主 Playwright（`NODE_PATH`）与系统 Chrome，供应商为 HTTP fixtures。尚未完成真实供应商扣费、真实数据库接入完整网关的生成结算联调，不把 handler fake relay 测试算作端到端结算证明。99 个新增/修改文本文件 UTF-8、无 BOM/替换字符；`git diff --check` 通过。


## 2026-11-17 统一图片预览补充

新增 `ImagePreview.vue`，将工作台旧的简单图片弹窗替换为统一的本地 Blob 查看器。覆盖上传/粘贴/拖入参考图、已保存/未保存生成结果、历史参考缩略图和素材。素材图片点击只预览，明确按钮才加入参考；生成历史按当前筛选集合连续浏览。支持完整比例、缩略图/前后切换、适应窗口/原始尺寸、按钮/滚轮/双指缩放、拖动、快捷键、下载当前图片、加载错误本地重读、实际尺寸/格式/文件大小及可折叠说明。关闭归还焦点、切图复位、账户切换关闭；短横屏合并工具栏并隐藏缩略图但保留导航，焦点循环跳过隐藏控件。

先观察素材预览和历史参考图入口用例失败，再接线通过。新组件验证尺寸/导航复位、URL 生命周期、错误重读、焦点归还与触摸拖动；审查发现双指中心计算方向错误及短横屏隐藏缩略图仍进入焦点循环，两项新增断言均先失败后修复通过。

最终命令（frontend 根目录）：

- `pnpm exec vitest run`：**426 文件 / 3465 项通过**。
- `pnpm typecheck`、本轮页面/组件/单测/locale/浏览器脚本的 `pnpm exec eslint`：通过。
- `pnpm run build`：i18n、vue-tsc 和 Vite 构建通过；保留已有 Browserslist 数据过期及大 chunk 提示。
- `node e2e/image-studio-preview.cjs`（宿主 Playwright，经 NODE_PATH）：真实本地文件上传、参考图/结果集合/历史参考/素材预览、完整比例、适应/原始尺寸、鼠标拖动、CDP 双指触摸输入、下载、嵌套弹窗 Escape 与焦点、手机横竖屏均通过。仅 2 次明确点击生成的 HTTP fixture 请求，所有预览/下载无额外 POST、无内容上传，页面错误为空。
- `node e2e/image-studio-ui.cjs`（同上）：原有 7 次生成、粘贴/拖入、局部编辑、模板系列、素材、分类、移动布局、刷新/删除回归通过。首次遇到并发请求到达顺序不同导致旧脚本断言失败；将测试修正为两条系列提示词各一次的无序集合断言，未改变队列实现，复跑通过。
- `impeccable detect --json --changed`：`[]`。UTF-8、无 BOM/替换字符、范围内尾空格、`git diff --check` 通过。

视觉已检查 `/tmp/sub2api-image-preview/` 中桌面浅色、手机深色横竖屏截图。使用真实 Vue/Blob/IndexedDB，供应商响应为 fixtures；触摸来自 Chrome CDP，尚未验证实体手机、Safari 或真实供应商扣费。本轮不涉及后端，不重跑后端检查。未新增依赖、迁移、提交、推送或部署；此前三项审查遗留仍未处理。

## 2026-11-17 配置流程、悬浮生成与显影动效补充

本轮只改工作台交互、中英文文案、测试和文档，不改后端网关/价格/用户权益，不调用真实供应商。管理员完整流程见 `docs/IMAGE_STUDIO_SETUP.md`：平台协议 → 分组与账号 → 定价 → 用户授权 → 工作台展示 → 普通用户最小出图验账。新增文档已加入 `.gitignore` 例外，避免丢失。

- 原按钮被视口或参数内部滚动裁切时显示底部生成操作条，原位置可见时收起；共用禁用及自定义尺寸约束。支持编辑参数并恢复输入焦点、桌面侧栏展开/折叠、移动横竖屏、安全区和 visualViewport 键盘偏移。弹窗/移动导航打开时隐藏底栏。
- 取景框、片层聚合和扫描光表达生成；静态排队，真实等待秒数和活动任务汇总，无伪造进度/ETA。当前会话完成的作品短暂显影，重读历史不重播。循环离屏暂停；隐藏事件暂停；减少动态效果保留静态等待画面及 150 ms 无位移完成确认。
- TDD：悬浮入口、相同一次提交、无效尺寸/弹窗、成功显影用例先失败后通过；等待组件 3 项用例先失败后通过。实际浏览器发现深色卡片仍白底，确认 Vue scoped 对 `:global(.dark) .generation-pending` 编译成仅 `.dark` 的根因；加入计算背景色断言观察失败，再修改为正确祖先选择器并通过。

新鲜验证：

| 命令 | 结果 |
| --- | --- |
| `cd frontend && pnpm exec vitest run` | **425 文件 / 3458 项通过**。含新增 visualViewport 键盘遮挡回归。 |
| `pnpm typecheck` | 通过。 |
| `pnpm exec eslint`（本轮页面、组件、测试、locale 和两个 UI 脚本） | 通过。 |
| `pnpm run build` | i18n、vue-tsc 与 Vite 构建通过。保留已有 Browserslist 旧数据、部分 chunk 超过 500 kB 提示。 |
| `node e2e/image-studio-ui.cjs`（宿主 Playwright，通过 NODE_PATH 解析） | 7 次 HTTP fixture 生成，参考图粘贴/拖入、模板系列、局部编辑、分类、移动端、刷新不重放等通过；页面错误为空。首次复跑因本地 Vite 已停止连接拒绝，重启后通过。 |
| `node e2e/image-studio-motion.cjs`（同上） | 5 次 HTTP fixture 请求；低高度/高视口切换、两种布局、侧栏宽度、弹窗与移动导航优先、手机横竖屏无溢出、编辑焦点、排队取消、独立失败/成功、普通/减少动画、离屏暂停/恢复、模拟隐藏事件暂停/恢复、新成功显影和刷新不重播通过。 |
| `impeccable detect --json --changed` | 空 findings（`[]`），未修无关样式或工具配置。 |
| UTF-8/BOM/替换字符/新增行尾空格、`git diff --check` | 通过；文档本地链接有效。 |

视觉证据：`/tmp/sub2api-image-studio-motion/` 的 normal/reduced 和 success-normal/success-reduced 各有固定时刻 PNG、filmstrip 与 JSON 采样；另有布局截图、排队前后截图和 pause-report。使用已登录 HTTP fixtures 的 Playwright 捕帧替代无鉴权上下文的通用 motion-capture 命令。独立 motion reviewer 第一轮要求完成序列无遮挡重采；第二轮结论 **ship（仅当前证据覆盖范围）**，无必须修复项。

边界：真实 Vue、IndexedDB 与 Chrome 渲染，供应商响应为测试数据。未实测真实供应商/真实扣费、物理手机/软键盘、帧率或实测对比度。headless Chrome 切页始终报告 visible，因此后台暂停只验证注入 visibilitychange 和单测行为，不能宣称真实切后台验收。键盘位置使用 visualViewport 自动化断言，不能替代物理设备验收。

此前三项审查遗留（收藏/分类跨标签删除竞态、失效分组覆盖阻塞、参考 AI 工作流能力缺失）本轮未处理；配置文档给出旧分组覆盖的临时移除方法。修改仍在 `feat/image-studio`，未提交、推送或部署。

## 2026-11-17 参考图直接输入补充

新增原生 paste 事件处理，支持工作台 Ctrl/Cmd+V 图片粘贴（无需点击剪贴板按钮）；普通文字和弹窗内粘贴不受影响。参考图区可键盘聚焦，拖入文件高亮并提示松开添加，嵌套元素不反复闪烁。连续粘贴/拖入在读取完成后再次核对当前模型数量限制。

先验证组件直接粘贴/拖入和并发超限用例失败，再实现并通过。浏览器用真实 ClipboardEvent/DataTransfer 文件载荷验证图片添加、文字事件不被阻止、拖入高亮、多图添加、超限拒绝和移除；不操作宿主系统剪贴板、不触发收费请求。修改范围为工作台页面、addReferences 校验、中英文提示、组件/状态测试及既有 UI 浏览器脚本。

本次验证：`pnpm vitest run src/features/image-studio src/i18n/__tests__/localeKeyCompleteness.spec.ts`（12 文件、66 项通过）、`pnpm typecheck`、目标文件 ESLint 和 `node e2e/image-studio-ui.cjs` 均通过；浏览器断言 `directImagePaste`、`plainTextPastePreserved`、`referenceDropAndLimit` 为 true，页面错误为空。原生快捷键通过 paste 事件接入；自动化派发剪贴板事件，不宣称操作过宿主系统剪贴板。

本次补充仅处理参考图输入。此前独立审查的三项问题（收藏/分类跨标签删除竞态、失效分组配置阻塞、参考 AI 工作流能力缺失）仍待处理，不以本次通过替代它们的验收。

日期：2026-11-17。分支：`feat/image-studio`。修改保留在工作区，未提交、推送、迁移或部署。

## 交付范围

| 用户要求 | 实现与证据 |
| --- | --- |
| 参照 infinite-canvas 生图工作台 | 固定参考提交 `dc1f1d7e741eaae24c7c11a8de8892f36ed01b49`；侧边/底部布局、参考图上传/粘贴/拖入、独立并发任务、取消/显式重试、预览/下载/复用、收藏、多分类、搜索、批量操作、素材、模板变量/系列、蓝色标记局部编辑；桌面明暗与移动端截图已检查。 |
| 用户无需配置 | `/image-studio` 使用现有登录身份，自动读取授权分组模型；前端没有 URL/API Key 输入。支持 GPT、Grok、Gemini 和配置了有效映射的 nano-banana 等图片别名；每次生成复核权限。 |
| 分组定价 | 同一进程调用既有同步网关，复用分组/用户倍率、模型与尺寸价格、余额/订阅额度和结算。别名报价采用真实计费模型回退；报价已含倍率，界面标明估算。 |
| 可控制模型展示 | `/admin/image-studio` 提供全局启停、按分组隐藏/展示、名称覆盖和既有模型协议声明。隐藏后手工提交旧工作台模型 ID 也拒绝；恢复自动名称和保存后刷新有回归。 |
| 只在 IndexedDB 保存内容 | 按用户分库保存图片 Blob、参考图、历史、草稿、素材、分类和模板。历史保存与任务移除在同一事务；损坏/缺图记录不显示。外部清理、失败事务、账号切换、刷新、跨标签同步、备份和原子导入均有验证。 |
| 服务端不留创作内容 | 不接异步图片任务或对象存储；关闭该路径的请求体审计、提示词入队、Ops 内容捕获和上游回显日志/账号错误持久化。审核判定、封禁及账务元数据保留。 |
| 避免意外重复收费 | 前端禁止 401 自动重放；排队任务先本地保存，刷新不提交；成功任务原子退场；局部编辑固定一张。Web Locks 识别另一标签页仍运行的任务，主动重试提示可能再次计费。 |

## 自动化检查

在各项目根目录运行：

| 命令 | 结果 |
| --- | --- |
| `cd backend && go build ./...` | 通过。 |
| `cd backend && go vet ./...` | 通过。 |
| `cd backend && go test -tags=unit ./internal/...` | 全部通过；最后 Gemini 尺寸和请求体上限调整另做受影响包的新鲜回归。 |
| `go test -tags=unit ./internal/handler ./internal/service ./internal/handler/dto ./internal/server/middleware ./internal/securityaudit -run 'Test.*(ImageStudio\|Ephemeral\|ImagesB64\|ImageURL\|Backfill\|RequestLogger)' -count=1` | 通过。最终后端定向回归；包括权限、隐藏、真实别名转发、计费回退、私有身份、请求体限制、协议、隐私及 URL 回填。 |
| `go test -race ./internal/service -run TestImageStudioIdentityConcurrent -count=1` | 通过，并发首次生成只创建一个内部身份。 |
| `go test ./internal/repository -run 'TestHTTPUpstream(PublicHostsOnly\|.*Private.*)' -count=1` | 通过，包含私网和重定向边界。 |
| `cd frontend && pnpm vitest run` | **424 个测试文件、3447 项测试全部通过**。 |
| `pnpm vitest run src/features/image-studio src/i18n/__tests__/localeKeyCompleteness.spec.ts` | 12 文件、62 项通过。 |
| `pnpm typecheck` | 通过。 |
| `pnpm lint:check` | 全量通过。 |
| `pnpm build` | 通过，含中英文 key 完整性检查。 |
| `git diff --check` | 通过。 |

对新增及修改文本检查 UTF-8 解码、BOM、U+FFFD 替换字符；检查均通过。中文界面已在真实截图中复核。

## 真实浏览器验证

本地 Vite：`pnpm exec vite --host 127.0.0.1 --port 5179 --strictPort`。使用已有 Playwright 与系统 Chrome；未安装依赖。可通过 `STUDIO_TEST_URL` 覆盖地址，`NODE_PATH` 指向本机已有 Playwright 运行时。

运行 `node e2e/image-studio-storage.cjs`、`node e2e/image-studio-ui.cjs`、`node e2e/image-studio-tabs.cjs`，三个脚本均通过。

- 存储 9 项：账号隔离、可解码 Blob 回读、成功任务原子移除、重开草稿布局、删分类保图、备份恢复、非法导入不部分写入、损坏图片过滤、外部清空可见。
- UI：7 次独立模拟生成、混合成败、失败任务删除、素材复用、局部编辑单张、变量系列、自定义尺寸、多分类、布局恢复、刷新不重复生成、外部删图、移动无横向溢出；管理页启停/隐藏/改名/自定义别名保存回读；无 JavaScript 页面错误。
- 多标签页 8 项：活跃 owner 保护、owner 持久化、完成同步、终态删除同步、关闭原页后转中断、不自动重放、不支持锁时提示明确且不改写共享运行状态。
- 网络断言确认除显式生成请求外，历史/素材/模板操作不上传创作内容。浏览器 HTTP fixtures 与真实 IndexedDB 组合，未调用收费模型。

## 独立审查与修复

交叉审查覆盖权限、分组定价、内部身份、SSRF、内容持久化、IndexedDB 和 UI。已修复：私有身份 DTO 回显/管理员改组入口、Antigravity/RequestLogger 内容日志遗漏、别名二次校验和报价回退、历史/任务非原子交接、旧读取失败覆盖新结果、管理名称重置、多标签页误判中断、局部编辑沿用批量张数及移动输入区遮挡。

## 验证边界

- 没有读取生产配置或使用真实上游凭据，因此没有声称真实供应商已成功出图扣费。上游实网、部署拓扑和真实资金结算仍需在部署环境验收。
- 本地记录随浏览器/站点数据清理而删除，不提供服务端恢复。单图存储上限 32 MiB，参考图单张 10 MiB，备份上限 128 MiB；超限明确报错。
- 构建保留既有的 Browserslist 数据较旧和部分产物大于 500 kB 提示；未因此更新依赖或重构无关页面。
- OpenSpec CLI 不在 PATH，规划文件按既有结构手工创建；未运行 CLI 校验。Impeccable skill 的 `scripts/detect.mjs` 在本机缺失，未运行该检测器，使用类型检查、ESLint、组件测试和真实截图检查替代。
- 参数依据：[OpenAI 图片生成](https://developers.openai.com/api/docs/guides/image-generation)、[Google 图片尺寸](https://firebase.google.com/docs/ai-logic/generate-images-gemini#supported-image-sizes)、[xAI 图片生成](https://docs.x.ai/developers/model-capabilities/images/generation)。Gemini 2.5/3.1 Flash Lite 限 1K，默认请求省略 `imageSize`；已知高分辨率型号开放 2K/4K。Grok 尺寸复用既有比例/分辨率转换。
