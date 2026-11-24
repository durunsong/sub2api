# v0.2.25 同步要求

## ADDED Requirements

### Requirement: 指定发布版本完整整合
系统 SHALL 支持 API Key 请求头的 user_id/api_key_id/group_id/request_id 变量、实际发送身份头错误详情、上游分组名称与备注手动及默认关闭的自动同步、billing include=group、多值 include、OpenAI 透传规则及结构化 400、compact 过载可重试错误与上游文字校验。

#### Scenario: 身份变量编辑与发送
- WHEN 管理员配置受支持变量且有效用户请求到达
- THEN 编辑器显示预览和行内错误，网关只以当前可信身份展开；后台请求不伪造用户身份，日志不保存模板秘密。

### Requirement: 保留 Fork 定制
系统 SHALL 保留 Kiro credits/缓存/粘性及用户端 Claude 别名、XorPay、Access Ban、全期订阅额度与永久卡、重置历史、邮件券近期活跃与零余额受众、批量兑换、生图本地存储与 NoContentPersistence、品牌与支付 UI。

#### Scenario: 合并后回归
- WHEN 完成代码整合
- THEN 原有专属模块与历史 SQL 不被覆盖；前后端回归、build/vet/lint/typecheck 和相关集成测试通过，UTF-8/BOM/冲突标记检查通过，无未处理的严重审查问题。
