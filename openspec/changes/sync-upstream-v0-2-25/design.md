# 同步设计

- GitHub latest/release API 于 2026-11-24 核实目标 v0.2.25，发布时间 11:21:08 Asia/Shanghai。tag object 8fab111f92447a3edbcdf67a517e45b6a3cc4d31，commit 348997267d756eab5729bb6b78a275a15d18796a。
- 初始 main 与 origin/main 同为 7431026d1，工作区干净。在当前 checkout 的 sync-upstream-v0.2.25 分支执行，验收后快进 main 并正常 push；保留原始提交，不改写历史。
- 历史为 squash 同步，采用官方 v0.2.24 文件作三方共同基线。15 commits / 91 paths / +3486 -319；36 路径含 Fork 差异，预合并无文本冲突。逐项复核身份变量、Kiro 排除探针、隐私标记、账务和管理员 API 权限。
- 官方 VERSION 仍为 0.2.24，本 Fork 按发布标签设为 0.2.25。无 SQL 变更；所有历史迁移保持字节一致。不修改部署配置。
- 仅跟随官方升级现有 gorilla/websocket 到 v1.5.4-0.20240701034025-d67f41855da4；保留 Fork go.mod/go.sum 其他差异，通过编译、测试、go mod verify 验证。
- OpenSpec CLI 本机不可用，沿用仓库既有文档结构并直接校验。用户已明确要求实施与 push，不停留在规划阶段。
