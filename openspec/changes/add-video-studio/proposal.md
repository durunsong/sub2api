# AI 图片/视频工作台

## Why
用户已选择新版 infinite-canvas 视频创作体验；现有图片工作台保持原样，通过 Tab 增加视频。

## What Changes
- 用户和管理员保留 `/image-studio` 与 `/admin/image-studio`，菜单改为 AI图片/视频工作台。
- 视频会话、居中/底部输入框、参数浮层、参考图、历史、素材、提示词、预览下载及任务恢复。
- JWT 请求通过既有内部网关授权与计费；创作内容仅保存到用户隔离 IndexedDB。

## Impact
新增独立 video-studio 模块，图片只增加页面集成控制。无依赖、迁移、生产操作或 Git 提交。
