# 设计

## 固定输入与方法
Fork 起点 `ee11c1aa6d3e15e9382506b8d99d17d594da0f09`，工作树干净；目标 tag object `8f35716ba69976fc0ec950894114d2604bb5454c`，commit `49a39b6dc1abed30fd227611e8af1108bc427610`。官方 VERSION 为 0.2.5，Fork 按发布 tag 设为 0.2.6。
使用旧官方 tag / 当前 Fork / 目标 tag 在临时目录逐文件三方预演，再写入原项目；不移动 HEAD、不暂存。39 个重叠路径，7 个文本冲突。CodeGraph 与 OpenSpec CLI 不可用，沿用已有 spec-driven 文件结构并使用 Git/rg，不安装额外工具。

## 保留与兼容
保留 Kiro OAuth/credits/cache_read/429 冷却/真实上游 ID/六个精简 DTO 字段，以及 Wire 中 Kiro 在 Grok 之前的依赖；保留 XorPay、Access Ban/登录失败自动封禁、提示词审计、订阅重置卡的期限明细/来源幂等/退款事务/旧接口兼容/提示、active_available、Ops 删除、Claude 别名、GLM 分类、VersionBadge 禁止在线更新与品牌定制。
支付金额输入同时保留 Fork 的人民币符号、快捷金额及 max 上限，并纳入官方非法输入恢复；必须在所有校验通过后更新最近合法值，不能提前覆盖导致超额恢复失效。
Wire 追加官方 SettingService 依赖而保留 Kiro/IPBan。导出测试同时保留 Kiro 格式保护与票据私密信息过滤。依赖采用上游安全升级，保留 Fork tokenizer 和更高 x/image 版本，不降级现有依赖；通过 Go 编译/测试验证。

## 数据与安全
所有历史 SQL 字节不变，不改变 Fork 金额精度、币种、事务与重置卡口径。兑换分页只查询当前用户，保持兼容旧不带分页参数的响应。Codex 票据保持上游开关/代理验证/状态脱敏/调度只读设计，不主动配置或调用真实采集端点。
受保护的 `deploy/config.example.yaml` 不读不写，新增配置通过应用默认值与后台设置提供；部署配置另行核对。无真实迁移/部署，回滚仅针对本次未提交代码差异，不重置其他工作。

## 验证
先合并金额输入测试并在旧实现观察非法输入恢复失败，再应用最小兼容合并；保留现有上限测试。执行 Go 普通/unit、vet/build，前端 Vitest/typecheck/lint/build。逐字节检查未涉及的 Fork 文件和全部历史 SQL，复核关键调用路径、UTF-8/BOM/冲突标记及最终 diff。无 Docker/PostgreSQL 时明确记录 integration 限制，不将单测当作真实支付/OAuth 联调。
