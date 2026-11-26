# 验收记录（2026-11-26）

## 已完成

1. 合并图片/视频 Tab、本地说明及右侧操作，移除“生图工作台”重复标题；桌面单行、手机自动换行；管理员 Tab 保留。
2. 修复视频交接 active 空档（回归先观察到 `[1, 0, 1]`，修复后连续）；轮询不滚动、不重挂载输入区；新记录只滚动结果面板；顶部操作固定。异步读取按序防覆盖、交接中暂停不会重启。
3. 重做结果卡的提示词/参数/状态排版，缩小运行/暂停/失败状态区，统一布局。
4. 参数面板点击内部保留；点击输入框、页面其他位置、切换工作台关闭；Escape 返回触发器焦点。
5. 提示词默认精选，上游七个来源、分类/多标签/搜索、分页、封面状态、详情、精确一键回填。原图片工作流和视频自定义存储保留在独立 Tab；不自动生成，不改模型/参数。

## 新鲜验证结果

- `pnpm --dir frontend test:run`：444 文件、3609 测试通过。
- 最终封面加载提示后，`pnpm --dir frontend exec vitest run src/features/video-studio src/features/prompt-library src/features/image-studio/components/__tests__/WorkflowLibrary.spec.ts`：11 文件、30 测试通过。
- `pnpm --dir frontend typecheck`、`pnpm --dir frontend lint:check`、`pnpm --dir frontend build`：通过。构建仍有既有 AccountsView 大 chunk 提示；未增加依赖。
- `node frontend/e2e/studio-polish.cjs`：20 张截图、2 次明确模拟生成、7 次状态查询、无 pageerror；检查轮询滚动位置、输入组件身份、图库精确回填、外部点击、手机无横向溢出、深色、离线 IndexedDB 缓存、公开源不带 Authorization/cookie。
- `node frontend/e2e/video-studio-acceptance.cjs`：36 场景通过，含暂停/刷新恢复、显式重试确认、播放下载、资产、管理配置、手机/短视口。
- `node frontend/e2e/video-studio-storage.cjs`：真实 IndexedDB Blob、用户隔离、旧 generation 和删除竞态通过。
- `node frontend/e2e/image-studio-tabs.cjs`：多标签页所有权、完成同步、无自动重发通过。
- `node frontend/e2e/image-studio-mask.cjs`：原图不变、原生遮罩像素、引导双参考、备份与重试通过。
- `node frontend/e2e/image-studio-preview.cjs`：全屏、原图预览/缩放拖动/触摸、素材、历史参考、手机及无多余生成通过。脚本等待两条结果保存完成后再展开，避免在临时任务切换为历史时点击已卸载元素。
- `node frontend/e2e/image-studio-workflow.cjs`：原有 AI 模板/系列审核、变量、显式生成、取消及手动重试通过。
- `node frontend/e2e/image-studio-ui.cjs`、`node frontend/e2e/image-studio-motion.cjs`：原图片布局/模板/备份/手机、等待动画/暂停/成功失败/无自动重发回归通过。
- 上述浏览器脚本使用宿主已有 Playwright/Chrome，通过 `NODE_PATH` 指定运行库；未安装新依赖。
- `git diff --check`、所有本任务文件 UTF-8/BOM/U+FFFD/个人路径检查通过。

## 审查结论与限制

已审查调用链、图片控制器不变、任务状态交接/暂停/账户隔离、缓存失败回退、公开 URL 协议/凭据边界、Vue 文本输出、移动端排版。未发现本次变更的阻断问题。没有后端、计费、鉴权、迁移或生产配置变更，因此未重复运行 Go 测试。

真实公开源快照已读取 Banana Prompt Quicker、DavidWu、Freestylefly、Awesome GPT Image、YouMind GPT Image 2。Awesome GPT-4o 和 YouMind Nano Banana Pro 在本机多次连接超时；相应网络失败有明确提示，未以模拟词条冒充。封面 CDN 在当前网络也存在延迟，展示加载/缺图占位；网络恢复可刷新。离线首次打开没有缓存的来源无法显示内容，已有缓存和自定义仍可使用。

生成接口全部为本地 fixture，未调用真实付费视频供应商。截图及相册保存在任务产物目录，不进入 Git。原有完整验收与本次 20 张截图分别保留。
