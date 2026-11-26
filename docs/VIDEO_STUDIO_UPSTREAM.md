# AI 图片/视频工作台：新版视频移植与维护

## 入口与功能

- 用户入口仍为 `/image-studio`，管理员仍为 `/admin/image-studio`。菜单均为“AI图片/视频工作台”，用“图片创作 / 视频创作”Tab 切换；`?tab=video` 可直达视频页。
- 视频采用用户确认的 **新版 v1.0.0 统一创作页的视频模式**：空白会话居中输入、点阵背景；生成后对话结果与底部输入框；模型选择、参数浮层、首尾帧/全能参考、引用图片、创作记录、提示词、资产、预览及下载。
- 参数包含 480p/720p/1080p/自定义、宽高、七种比例（含自动）、4–30 秒。首尾帧最多 2 图，全能参考最多 9 图；单图 10 MiB、总参考图 30 MiB。这些是工作台输入范围，最终支持的组合以所选上游为准。
- 图片与视频组件分别保留草稿。原图片 API、生成控制器、存储、遮罩、工作流执行逻辑均不变；图片视图使用合并工具栏，提示词弹窗新增精选/自定义 Tab。

## 配置与实际支持边界

1. 先在既有 Grok 或聚合分组中配置可调度的视频渠道、模型映射、授权和视频价格。现有网关要求分组开启 `allow_image_generation`，该开关同时控制图片/视频生成。
2. 在管理页的“视频创作”Tab 中启用工作台，按分组/关键词筛选，设置显示名称、隐藏模型或添加已映射的视频别名。全选只影响当前筛选中的有效分组；失效分组的旧覆盖可保留或删除，不能修改。
3. 工作台配置保存在独立 `video_studio_config` 设置项中，图片 Tab 不会被覆盖。此处只管理展示，不授予分组权限、不创建上游模型、不更改价格。
4. 用户按本人分组权限发现可用模型。刷新或“继续获取结果”只查询已有 task ID；暂停获取不取消供应商任务，也不承诺退款。重新生成会明确确认，并按新请求计费。

本次接通 **已有 Grok 视频网关及其聚合路由**。上游仓库的其他供应商、任意运行脚本、画布、音频创作、Agent 和插件系统不在本次接入范围；不能将外观和交互移植理解为已接通上游所有供应商。升级这些能力时，需要单独实现授权、协议、安全审查和计费契约。没有使用浏览器供应商密钥，也没有执行真实付费调用。

## 上游基线与对应关系

参考仓库：<https://github.com/basketikun/infinite-canvas>。

固定提交：`4c67f1ece4f4edaef81e58e964eb78abb1e9809d`（2026-11-25，v1.0.0 系列）。不使用旧独立“视频创作台”页面，也不与旧 tigerowo Fork 混用。移植目标是现有 Vue 应用，复用 sub2api 的主题、通用选择器和弹窗，不引入 React 运行时。

下表的上游路径均位于 `packages/`；前端目标位于 `frontend/src/features/video-studio/`。

| 上游文件 | 本地对应 | 保留内容 |
| --- | --- | --- |
| `workbench/src/pages/create/index.tsx` | `components/VideoStudio.vue` | 会话、空态、结果区与输入区切换 |
| `workbench/src/pages/create/components/create-composer.tsx` | `components/VideoComposer.vue` | 参考图、提示词、模型、参数和工具入口 |
| `workbench/src/components/video-settings-panel.tsx` | `components/VideoSettings.vue` | 清晰度、尺寸、比例、秒数、模式 |
| `workbench/src/lib/media-size.ts` | `parameters.ts` | 比例/分辨率尺寸转换及默认值 |
| `workbench/src/pages/create/components/create-entry.tsx`、`video-result.ts` | `components/VideoEntry.vue`、`media.ts` | 视频结果、真实文件元信息、下载和加入资产 |
| `workbench/src/pages/create/components/create-history.tsx` | `components/VideoStudio.vue` | 会话搜索、打开和删除 |
| `workbench/src/pages/create/use-create-generation.ts` | `useVideoStudio.ts`、`tasks.ts` | 提交、状态查询、恢复和失败处理 |
| `workbench/src/services/create/logs.ts` | `types.ts`、`storage.ts` | 会话与记录模型（改为用户隔离 IndexedDB） |
| `model-runtime/src/call-scripts/xai.ts` | `backend/internal/handler/video_studio.go` | `duration/resolution/aspect_ratio`、`image/last_frame/reference_images` 映射 |

后续更新流程：先记录新的上游提交，按表逐项查看 diff；把 UI/参数变更映射到独立模块；新增供应商必须先补网关契约；运行下列回归后再更新本文件的基线。不要把上游配置、远程文件存储或凭据方案直接复制进用户端，也不要改动已有图片模块来迁就视频。

## 本地数据、恢复与隐私

- IndexedDB：`sub2api-video-studio-{userID}`，版本 1，`records/drafts/prompts/assets` 四个 store。视频/参考图片使用 Blob；无服务端创作历史、提示词或媒体库。可读取本人原有图片资产作为参考图，不迁移或修改图片数据库。
- 先保存不可变请求记录，再发起生成；task ID 到达后立即保存。已知 ID 的 pending 任务在刷新后继续查询；没有 ID 的记录保留“提交结果未知”提示，绝不自动再次生成。
- 记录更新按 `id + generationId` 原子校验；跨标签页刷新使用 BroadcastChannel，删除后的晚到回调不能复活记录。账户切换中止旧请求、关闭旧连接并清空视图状态。
- 下载或保存失败可继续获取原任务。结果已下载但 IndexedDB 配额不足时保留内存结果，允许下载/重存；关闭页面前应先下载，浏览器清理站点数据会删除本地创作内容。
- JWT 私有接口通过进程内同步网关，复用内部账务身份、分组权限、任务所有者绑定、审核和既有视频结算；不会把面板 JWT 转发上游。
- `NoContentPersistence` 贯穿生成/查询/内容代理，面板审计排除生成请求正文及查询串；仅保留既有身份、任务归属、账务和安全元数据。尾帧仅在视频路径中纳入审核输入，图片解析行为不变。响应禁止缓存，不向浏览器暴露上游内容 URL。
- 单个下载上限 256 MiB；新代理共享输出内存预算，超限返回可重试的读取错误，不重发生成。校验响应长度、拒绝意外部分响应及上游断流；浏览器解码验证后才保存成功，并记录真实尺寸/时长。无数据库迁移和新增依赖。

## 本地验证

前端：`pnpm typecheck`、`pnpm lint:check`、`pnpm test:run`、`pnpm build`。

后端：`go build ./...`、`go vet ./...`，以及
`go test ./internal/handler ./internal/service ./internal/server/middleware -run 'Test(VideoStudio|ImageStudio|.*Grok.*Video|.*Grok.*Media)' -count=1`。

浏览器在本地 Vite `127.0.0.1:5179` 上运行。使用已有 Playwright/Chrome，通过 `NODE_PATH` 指向宿主提供的运行库，不安装依赖：

- `node e2e/video-studio.cjs`：Tab 草稿、参数、Blob 落库、刷新恢复不重复提交、资产、删除、管理配置、深浅色及移动端。提供 `STUDIO_VIDEO_FIXTURE` 指向本地有效 MP4，默认使用临时目录中的测试视频；所有 API 都由本地 fixture 拦截。
- `node e2e/video-studio-storage.cjs`：真实 IndexedDB 用户隔离、Blob 往返、删除竞态与过期 generation 拦截。
- `node e2e/video-studio-acceptance.cjs`：完整交互验收及 36 张截图，自动输出可逐张打开的 `index.html` 相册和 `manifest.json`；`STUDIO_SCREENSHOT_DIR` 指定输出目录。覆盖前后台、付费重试确认、离线/停用、播放下载、移动端和短窗口弹层可操作性。
- 原 `image-studio-*.cjs` 验证图片行为不变。

这些测试不证明某个生产渠道已开通，也不替代使用真实授权渠道的付费联调。

如本机已有 ffmpeg，可生成可重复的测试视频：
`ffmpeg -f lavfi -i color=c=teal:s=320x180:d=1 -c:v libx264 -pix_fmt yuv420p /tmp/sub2api-video-fixture.mp4`。

本次验收结果与审查修复记录见 `openspec/changes/add-video-studio/verification.md`。

## 2026-11-26 交互整改

- 图片/视频 Tab、本地存储说明、原有图片操作合并为紧凑工具栏，删除重复标题；移动端换行。
- 视频提交与进度查询连续交接，旧 IndexedDB 读取不能覆盖新状态；暂停交接不会自动重启。只在新增记录/切换会话时滚动结果面板，不随进度更新滚动，不移动外层页面。结果区标题和操作固定，单一输入组件保留焦点和输入状态。
- 统一提示词、元信息、进度/失败卡片。参数弹层仅排除触发器和面板本身，点击输入区/页面其他位置、Escape、切换 Tab 均关闭。
- `frontend/src/features/prompt-library` 对应上游 `workbench/src/services/api/prompt-source-presets.ts`、`prompt-source-runtime.ts`、`prompts.ts` 和 `components/prompts/prompt-select-dialog.tsx`。沿用同一固定上游提交和本文 MIT 许可。
- 默认精选提示词包含上游七类来源，直接读取 `yukkcat/image-prompts` 的公开 JSON。分类、搜索、标签 OR 筛选、缩略图、详情和一键回填；一键使用只填提示词，不自动生成、改变模型或附加参考图片。自定义 Tab 沿用用户原有库；“创作工作流”按钮直接进入自定义。
- 公开提示词使用独立 `sub2api-public-prompt-library` IndexedDB 缓存，一小时过期，可手动刷新；失败保留缓存并显示失败来源。公开数据可以跨本机账户复用，私有提示词、草稿和媒体仍在原有按用户隔离的数据库里。请求不带面板 JWT/cookie；只呈现 HTTPS 图片和来源链接，不执行远程脚本/HTML。封面仍依赖原来源网络，失败显示占位卡，不影响使用文本。
- 测试 `node frontend/e2e/studio-polish.cjs`（使用宿主已有 Playwright）覆盖 20 个截图场景、轮询不跳动、精确文本回填、原有自定义、参数点击关闭、暗色/手机和缓存离线。可设置 `STUDIO_PROMPT_FIXTURE_DIR` 为公开 JSON 快照目录、`STUDIO_SCREENSHOT_DIR` 为截图目录；提供快照目录时缺失来源按网络失败模拟，绝不以自造提示词冒充上游数据。未提供时仅使用明确的测试 fixture。

## MIT 许可

移植与参考上述源文件，保留原许可全文：

```text
MIT License

Copyright (c) 2026 basketikun

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
