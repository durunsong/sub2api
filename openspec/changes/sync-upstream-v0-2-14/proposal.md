# 同步官方 v0.2.12–v0.2.14

## 目标
从现有 v0.2.11 Fork 同步到最新稳定发布 v0.2.14，保留全部已有定制。用户已要求实施升级。

## 范围
官方基线 96f4c115c9749078f90cbf210a01d39baf3f53b6，目标 0363b8cdba8cec3e2ba4b2dbd49c4481143fa55d（tag object 1400a7b482974d98db5b284a8b2afbe3eaf9aaef），Fork 起点 2fd1d9ae7。47 commits / 176 files / +6,801 / -475；排除 8 个 deploy 路径，逐文件三方同步其余 168 个路径。

新增 TypeSafe /v1/systemone、充值优惠阶梯、账号优先级快捷调整和 Key 分组排序；同步验证码与密码重置安全、Antigravity 错误脱敏、匿名订单限流、结算容错、EasyPay 回调防伪和新装管理员校验。前端现有依赖跟随官方安全升级。

## 非目标
不部署、不启动服务、不执行数据库迁移或真实支付，不读取秘密/本地配置，不提交或推送，不合入 release 之后的 main。既有 API Key 数量并发竞态如上游未修复则保留为已知风险，不额外扩大升级范围。

## 验收
保留全部定制；历史 SQL 字节不变；新平台约束含 Kiro 与 TypeSafe；后端 build/vet/default 和 unit 测试、前端 test/typecheck/lint/build 通过；差异、UTF-8、BOM、冲突标记检查通过。
