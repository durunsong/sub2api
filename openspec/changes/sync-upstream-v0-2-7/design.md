# 设计

## 固定输入
起点 HEAD `0c99c78fb268f1f5c34be222457bdeccd6f3058e`，初始工作区干净，文件树与 v0.2.5 Fork 提交相同。旧官方基线 `86f93c28ee34cc74b629dafb748bd5ac5ca8c5ea`；目标 tag object `7484192016807acf55c6ef4f2827d371eb61ec1c`，commit `aea725f2ea644d5592d0bbb1d63b607efa7e200a`。不带入 main 或已不在 v0.2.7 的 Codex 票据模块。

## 合并
临时目录逐文件三方合并，审查 27 个重叠路径。处理 Wire 时同时注入 PluginKVStore/AccountDirectory 与原 Kiro provider；保留 Kiro 在 Grok 前。独立审查发现上游仅在生成文件手写 SetAccountDirectory，会被 Wire 重新生成丢弃，因此增加 ProvidePluginManager 注册到源 ProviderSet 并由生成器产出调用。回归通过真实宿主 RPC 验证声明能力可用、未声明能力仍拒绝。金额输入保留人民币/快捷金额/max 限制并恢复最近合法文本。监控测试沿用实际平台数量。
OpenSpec CLI 和 CodeGraph 不可用，沿用仓库已有 spec-driven 文档结构与 Git/rg，不安装工具。已有用户授权覆盖本地可逆整合，规划后继续实现。

## 定制与数据
保留 Kiro credits/cache_read、429 冷却、真实上游 ID/状态 DTO、XorPay、Access Ban/admin 登录失败封禁、提示词审计、重置卡快照/幂等/退款事务/旧接口、active_available、Ops 删除、Claude 别名、GLM 分类、支付金额分离/隐藏倍率、品牌/Select/ConfirmDialog、VersionBadge 禁止在线更新。
Seedance 新根路由通过现有 rootRoute，必须保留 Access Ban、API Key、白名单和 Composite 校验。兑换分页保留当前用户隔离与重置卡来源传递。
298 个历史 SQL 字节不变，不执行数据库操作，不改变 Fork 金额精度、币种或重置卡事务边界。交付保持未提交可审查差异。

## 验证
先运行金额输入合并回归观察旧行为失败；合入后复跑。执行前端全量 Vitest/typecheck/lint/build，后端普通及 unit 测试、vet/build。按 Git blob 验证非重叠 Fork 文件、历史 SQL、版本、编码和 whitespace。外部供应商/数据库集成不由单元测试代替。

## 提交前审查修复
新增插件账号目录在解析旧 ID 时也校验 StatusActive，与列表边界一致；disabled/error 状态不得获得凭据。单纯暂停调度但仍 active 的账号继续可用，保持本次官方暂停后刷新能力。仅增加入口校验和真实目录回归测试。另为 docs/seedance-api.md 增加 Git ignore 例外，确保上游文档随代码提交。
