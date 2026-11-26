# Design

## Baseline and approach
用户确认采用新版。固定 `basketikun/infinite-canvas` 提交 `4c67f1ece4f4edaef81e58e964eb78abb1e9809d`（v1.0.0 系列）。按视频模式移植 Vue，保留 MIT 声明。采用现有组件和主题，避免第二套 React 运行时。原图片组件保持挂载，仅在视频 Tab 激活时隐藏并停用全局粘贴/浮动按钮；图片生成、存储、遮罩与工作流不改。

## Data and requests
独立 `sub2api-video-studio-{userID}` IndexedDB：records、drafts、prompts、assets。记录带 sessionId/generationId，原子条件更新防止删除后复活。刷新仅恢复已知 task ID 的查询，不自动重新收费；没有 ID 的遗留任务显示提交结果未知。用户切换中止旧请求。存储失败保留结果供下载和重存，同账户刷新不丢弃内存恢复信息。
视频设置独立保存于现有 settings 的 video_studio_config，不修改图片 schema。模型目录只提供已授权的现有 Grok/聚合视频路由。请求使用私有内部身份与 NoContentPersistence，经原网关验证任务所有者和现有视频计费。参数按上游 xai.ts 适配，未接通供应商不伪装可用。不增加浏览器密钥或任意脚本执行。

## Source mapping
| Upstream | Target |
| --- | --- |
| pages/create/index.tsx | VideoStudio.vue |
| pages/create/components/create-composer.tsx | VideoComposer.vue |
| components/video-settings-panel.tsx, lib/media-size.ts | VideoSettings.vue, parameters.ts |
| pages/create/components/create-entry.tsx, video-result.ts | VideoEntry.vue, useVideoStudio.ts |
| services/create/logs.ts, services/generation-records.ts | types.ts, storage.ts |
| model-runtime/src/call-scripts/xai.ts | handler/video_studio.go |

## Acceptance
前后端自动化覆盖严格入参/权限/隐私、参数转换、防收费重放、恢复、账号隔离、删除竞态；浏览器实际验证 IndexedDB、Tab 草稿保持、响应式与主题。运行图片回归、typecheck、lint、构建、编码/diff 检查。当前未发现 OpenSpec CLI，按仓库既有 spec-driven 结构持久化，不安装工具。
