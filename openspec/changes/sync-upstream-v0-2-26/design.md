# 同步设计

- 2026-11-29 核实 GitHub releases/latest HTTP 302 指向 v0.2.26（API 请求返回 403）；目标 commit 9494105c7e350937a61bd78b3f913a5a10867f58。
- 初始 main/origin/main 为 ac387298c，工作区干净。在当前 checkout 的 sync-upstream-v0.2.26 分支实施，通过验证后快进 main 并正常推送，不改写历史。
- 历史采用 squash 同步，以官方 v0.2.25 为共同文件基线做三方整合。152 commits / 201 paths / +10030 -1110；保留两份部署示例，199 个源码路径中 59 个存在 Fork 差异，11 个文本冲突。逐文件保留定制，不覆盖为纯官方版本。
- 无新增依赖或 SQL；历史迁移与部署配置保持字节一致。官方 VERSION 为 0.2.25，Fork 按标签设为 0.2.26。
- 关键交叉点：Kiro credits 与新增上游错误统计并存；粘性 CAS 保留 Kiro 分组 TTL；字节流改造继续采集 Kiro credits；审核惰性队列与图片安全事件保持 NoContentPersistence，不持久化创作内容。
- 本机无 OpenSpec CLI，沿用已有 openspec/changes 结构记录并直接检查文档；不安装工具。未配置可调用 CodeGraph，以 Git/rg 和项目测试分析。
- 验收：官方非重叠文件一致，全部冲突消除；前端 Vitest/typecheck/lint/build、后端 unit/build/vet 和相关隔离集成检查；独立审查无阻塞问题；UTF-8/BOM/乱码/差异检查通过；远端 main 与最终提交一致。
