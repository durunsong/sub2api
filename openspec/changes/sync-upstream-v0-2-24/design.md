# 同步设计

目标 tag object `7515db6dc56c2a6120dfb7abed70ef089ebc1457`，commit `5c2982a81b46e2142d0143d2f83c1ce9c4da1d24`。官方 VERSION 仍为 0.2.23，Fork 按发布标为 0.2.24。

已有同步采用压缩提交，Git merge-base 不是 v0.2.23，不能直接 merge 标签。以 v0.2.23 文件为共同基线逐文件三方整合 190 个路径；54 个有 Fork 差异的路径保留定制并人工复核冲突。原工作区干净，在当前 checkout 完成；不改其他仓库。

无新增依赖及 SQL；所有历史迁移必须逐字节不变。Kiro credits、缓存 TTL、全期订阅额度、永久重置卡、XorPay、访问封禁、邮件券、用户品牌与生图 NoContentPersistence 不回退。重复 JSON/multipart 字段与 WebSocket 帧采用官方严格拒绝规则；安全校验须在所有公共路由和生图内部网关路径生效。

先在旧代码运行上游模型候选回归证明 BOM/multipart 白名单问题，再整合实现和测试。兼容性问题遵循最小失败测试再修复。后端全量 unit/build/vet/embed，前端全量 Vitest/typecheck/lint/build，必要的数据库与浏览器测试使用本地临时容器/fixtures。

提交前独立审查整体增量与共享路径。推送前复核 origin/main 未前进，使用普通 push；保留原提交可回溯，不改写历史。本次不改变金额精度、事务、幂等和补偿口径。OpenSpec CLI 本机未安装，沿用项目既有文档结构记录并人工验证引用及编码。
