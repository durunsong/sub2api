## Why

将定制 Fork 从 v0.2.2 同步至官方 v0.2.3，保留全部现有定制。

## What Changes

- 固定 release tag，合入 9 commits / 22 files / +1517 / -24。
- Ollama Cloud DeepSeek 跨 Chat Completions、Responses、Messages 输出上限处理及 Anthropic Bearer 鉴权。
- 账号测试模型显示名补全，且不污染共享模型目录缓存。
- 纳入 236 白名单列修复迁移，显示版本设为 0.2.3，维护定制清单。

## Capabilities

### Modified Capabilities

- `upstream-release-sync`: release 增量完整性及 Fork 兼容。

## Impact

仅 Go 后端、一个新增 SQL 迁移及文档；没有前端和依赖版本变化。

## Non-Goals

不引入 main 后续提交，不改变定制业务口径，不读取秘密配置，不运行真实迁移或部署，不暂存、提交或推送。
