# 验收记录（2026-11-26）

## 最终结果

| 检查 | 实际结果 |
| --- | --- |
| 前端 `pnpm typecheck` | 通过 |
| 前端 `pnpm lint:check` | 通过，无错误 |
| 前端 `pnpm test:run` | 441 个文件、3602 项测试通过（追加审查后的新鲜结果） |
| 前端 `pnpm build` | 通过；保留现有 AccountsView 大 chunk 提示，无新增构建阻断 |
| 后端 `go build ./...` | 通过 |
| 后端 `go vet ./...` | 通过 |
| 后端聚焦测试（命令见下） | handler、service、middleware 三包通过 |
| `node e2e/video-studio.cjs` | 通过；实际发起 1 次模拟创建、2 次查询；刷新没有再次创建 |
| `node e2e/video-studio-storage.cjs` | 通过；真实 IndexedDB Blob 往返、用户隔离、过期 generation 和删除竞态 |
| `node e2e/video-studio-acceptance.cjs` | 通过；36 张截图，3 次模拟创建、4 次查询，覆盖正常/失败重试/离线/停用、播放下载、后台批量配置与弹层实际点击区域 |
| 图片 `image-studio-storage.cjs` | 通过：12 项存储/恢复/账户隔离检查 |
| 图片 `image-studio-tabs.cjs` | 通过：8 项跨标签页保护与无自动重发检查 |
| 图片 `image-studio-preview.cjs` | 通过：预览、全屏、键鼠/触摸、窄屏及历史/素材来源，pageErrors 为空 |
| 图片 `image-studio-mask.cjs` | 通过：遮罩、擦除、重试、备份、普通生成不串用遮罩和非遮罩模型引导模式 |
| Impeccable `detect --json --changed` | 无发现（`[]`） |
| UTF-8 / BOM / 替换字符 / 个人绝对路径扫描 | 通过 |
| `git diff --check` | 通过 |

后端实际命令：

```sh
go test ./internal/handler ./internal/service ./internal/server/middleware -run 'Test(VideoStudio|ImageStudio|.*Grok.*Video|.*Grok.*Media)' -count=1
```

浏览器脚本使用宿主已有 Playwright/Chrome、本地 Vite 5179 和本地 MP4 fixture。视频脚本覆盖 Tab 草稿保持、全屏 Escape、参考图片上传、参数映射、实际媒体元信息、生成/恢复、资产、历史删除、管理端独立保存、深浅色和移动端无横向溢出。图片业务模块 `frontend/src/features/image-studio/` 没有差异。

## 独立审查与修复

按 executing-plans 技能派发一次只读独立审查；未发现 Critical。以下为首轮审查修复，先观察回归失败，再运行修复后检查。用户随后要求的追加审查见下节；没有再次派发独立审查代理。

1. Final: fixed 同账户刷新丢失未保存 task ID / Blob — `keeps unsaved task/blob recovery data through a same-user refresh` RED→GREEN；另验证切换账户清空私有内存结果。
2. Final: fixed 截断/部分响应误记成功 — `TestVideoStudioContentRejectsIncompleteSuccess` 与 `TestVideoStudioChunkedContentSignalsReadFailure` RED→GREEN，覆盖声明长度、意外 206、空响应及 chunked 断流。
3. Final: fixed 共享尾帧解析改变图片行为 — `TestVideoStudioDoesNotChangeImageReferenceParsing` RED→GREEN；新增视频专用解析，`TestVideoStudioLastFrameIncludedInSafetyInputs` 证明尾帧仍进入视频审核。
4. Final: fixed 维护文档及 MIT 被忽略 — `git check-ignore` 先确认命中 `docs/*`；增加单文件白名单后 `git status` 显示文档为待纳入文件，无暂存操作。
5. 审查最初把 WebM 扩展名定为 Minor；按下载产物会被外部编辑器错误识别的实际影响提升为 Important。Final: fixed 使用实际容器扩展名 — `uses the actual container extension for downloaded assets` RED→GREEN，卡片与资产入口共用函数。

首轮修复后整套前端测试 3601/3601、相关后端三包与构建检查通过。

## 用户追加验收、截图与推送授权

用户要求逐项截图、代码审查、功能验收，并授权检查通过后直接 push。本轮重新阅读私有接口、内部凭据、任务归属、响应下载、前端状态控制、IndexedDB、管理端及图片页接入代码，以多用户、多标签页和账户切换为验收场景。

追加审查发现并修复：

1. 390px 手机下参数弹层越过左侧边界。浏览器边界断言先失败；改用视口内定位后通过。
2. 桌面空态的参数弹层被工作区 overflow 裁剪。`elementFromPoint` 实际命中检查先失败；与现有 Select 一样通过 Teleport 离开裁剪容器，按视口边界定位，监听窗口尺寸/滚动并清理监听器。再验证 390px 手机及 1280×600 短窗口，顶部可见且可交互，Escape 和外部点击保持可用。
3. 旧账户的“重新保存”完成后会清空新账户的存储错误。新增挂起写入、切换账户、再完成写入的测试，观察 RED 后增加 epoch 校验，GREEN。旧账户的回调不能再刷新、恢复查询或更改新账户状态。
4. 后台视频 Tab 的跳转按钮误用图片文案，改为独立中英文本“打开视频工作台”。

本轮最终验证重新执行前端 typecheck、lint、441 文件/3602 测试及 build，全部通过。后端聚焦三包、build、vet 通过。图片 storage/tabs/preview/mask 四套浏览器回归再次通过；视频 native IndexedDB 检查通过。图片业务模块仍无差异。Impeccable 检测为 `[]`。没有遗留的已确认阻断问题。

截图脚本 `frontend/e2e/video-studio-acceptance.cjs` 生成 `index.html`、`manifest.json` 和 36 张 PNG；通过 `STUDIO_SCREENSHOT_DIR` 选择输出目录，截图与测试媒体不提交仓库。相册中的视频是本地 320×180 MP4 验收素材，API 使用隔离测试替身，没有冒充真实生成结果。

| 截图 | 覆盖 |
| --- | --- |
| 01–08 | 图片原页、视频空态、模型、首尾帧、@ 引用、参考图预览、参数及全能参考模式 |
| 09–13 | 提示词空态/保存复用、资产空态/图片与视频上传/预览 |
| 14–20 | 生成中、暂停、重试确认、结果/真实元信息/播放/下载、生成结果加入资产、历史检索与删除确认 |
| 21–24 | 全屏、真实深色主题切换、手机结果、手机参数 |
| 25–29 | 上游失败、确认后新建任务、离线、后台停用、无可用模型，保留本地结果 |
| 30–35 | 后台图片/视频独立 Tab、显示名保存、添加别名、启停、筛选后批量显隐/恢复自动 |
| 36 | 1280×600 窗口的参数面板边界及实际点击区域 |

首次模拟生成暂停后刷新只查询原任务，不重复创建。取消重试不创建任务；明确确认重试后恰好增加一条任务。后台视频保存没有调用图片配置写入接口。前端没有 pageerror。

## 限制与范围

- 用户已确认选用新版统一创作页的视频模式；固定上游提交及维护映射见 `docs/VIDEO_STUDIO_UPSTREAM.md`。
- 本次实际接通现有 Grok/聚合视频路由；其他供应商协议、任意脚本、Agent/音频/画布不在此次范围内，不声称完整供应商一致性。
- 没有执行真实付费上游请求，生产渠道的模型能力、参数组合和价格需按管理员实际配置验证。
- 本机未发现 OpenSpec CLI，因此沿用仓库 spec-driven 目录手工维护 proposal/spec/design/tasks，没有安装新工具。
- 初次验收没有提交或推送；本轮用户已授权验收通过后提交并推送。没有创建 PR、部署、执行迁移或新增依赖。
