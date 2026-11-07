# 同步设计

## 精确三方合并
Fork HEAD c53b90760 为工作区快照。以官方 v0.2.19 为 base、Fork HEAD 为 ours、官方 v0.2.20 为 theirs，在临时目录逐文件三方预合并，不改 Git 索引。逐段解决冲突并审查所有重叠文件；独有 Fork 文件和历史 SQL 做字节保留校验。VERSION 设为 0.2.20（上游文件仍为 0.2.19）。

## 兼容与权限
保留 Kiro credits、cache_read 去重、429 独立冷却、真实上游请求头、冻结计价时刻、严格代理与单次推理倍率。新增 MCP 路由必须继承现有 Access Ban 边界；精确邮箱黑名单与已有邮箱后缀封禁/格式校验并存。保留 Kiro/Grok provider 顺序、邮件 Worker 清理、永久卡期限/幂等/退款事务及余额券定制。
前端保留 Claude 品牌、GLM 分类、支付金额、隐藏倍率/在线更新、客服二维码和订阅倒计时。新管理员全权密钥按上游显式确认流程，不能默认为全权。

## 数据与依赖
仅新增 249_mcp_oauth.sql：OAuth grants/codes/credentials 表，凭据仅摘要存储；不改历史 SQL、不执行迁移。上线前备份、测试迁移及 OAuth 撤销/轮换；回退二进制不删除新增表，已发令牌需撤销。资金仍沿用现有 USD/token/credits 口径、精度与幂等事务；本次不更改业务定价或重置卡规则。
按官方锁文件引入 MCP go-sdk v1.4.1 与 moego v0.1.0-alpha.1，分别提供协议实现和 JSON 批处理；自写替代会偏离 release 且扩大风险。已有 brotli/compress/vue-i18n 等升级包含安全修复。moego 为 alpha，保留上游回退与预算测试；vue-i18n v11 用类型检查和全量前端测试验证兼容。仅锁定依赖同步，禁用前端安装脚本，不另加依赖。

## 验收
先引入上游 HTTP 408→Google DEADLINE_EXCEEDED 测试观察旧版失败，再合入实现；冲突适配按需补回归。运行后端 build/vet/unit、前端 Vitest/lint/typecheck/build，逐文件核对 upstream 增量与 Fork 差异、SQL、UTF-8/BOM/冲突标记。CLI 不可用，按已有 OpenSpec 格式维护文档并做本地结构检查。

## 实施中发现的兼容遗漏
原 v0.2.19 Fork 图片计费公共入口缺少上游按张开关 guard，既有两个回归分别观察到期望 1/0.8 USD、实际 0.018/3.15 USD。按 v0.2.20 官方条件恢复，仅作用于专用图片端点，混合请求和 Kiro opts 不变。新 OAuth suffix-only helper 会绕过原有格式/封禁校验，显式补回两项检查，保留既有用户精确黑名单登录规则。

## 并发工作保护
实施期间另一个任务提交 4996405cc（用户端批量兑换码），相对初始 c53b90760 新增8个文件改动；已将两份语言文件增量合回上游更新。该任务其余源码与测试不覆盖，后续三份兑换文件格式化改动也保留。AGENTS.md 与差异文档的并发格式化在当前内容上增量更新，不回滚。

普通 API 的旧 IPBanGuard 命中后仅写响应而未 Abort，新增 MCP 路由测试暴露下游仍执行。新增 IP/UA 两个副作用测试观察红灯后，在共享 blocked 分支统一 c.Abort；不改封禁匹配、可信代理或错误码。
