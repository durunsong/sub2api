# 同步设计

## 输入与方法
输入见 proposal.md。工作区初始干净，直接在用户指定项目中实施；不操作暂存区。逐文件比较官方基线、Fork HEAD、官方目标，保留非重叠 Fork 文件，对 80 个重叠路径逐项审查；禁止官方整树覆盖。CodeGraph 与 OpenSpec CLI 当前不可用，使用 Git/rg 和项目既有 spec-driven 文档结构。

## 定制兼容
Kiro 的 provider 注入、credits/cache_read、真实上游请求 ID、429 冷却、Composite/配额/调度继续保留；平台枚举同时增加 TypeSafe。新 /v1/systemone 继承 Gateway Access Ban、鉴权、内容审核和提示词审计。新平台不加入官方未支持的监控列表。
XorPay、支付实付/到账分离、隐藏用户倍率、人民币快捷金额保留；充值促销嵌入 Fork 现有页面布局。登录失败自动封禁、邮箱格式/后缀规则、永久重置卡、active_available、退款来源幂等、Ops、VersionBadge、GLM 分类和品牌别名保持现有行为。

## 金额、安全和历史数据
充值优惠严格采用目标 release 的口径：余额订单按输入金额匹配最高适用档位；赠金增加到账 USD，折扣降低实付基数；订阅不参加。计算复用 decimal 和现有币种精度，bonus_amount 为 DECIMAL(20,2)，默认 0；订单创建时冻结优惠，后续配置变更不影响旧单；推广返利剔除赠金。沿用现有支付幂等、事务发货和退款补偿，不修改永久卡口径。
两个新 241 SQL 仅交付：payment_orders 新增 bonus_amount；平台配额/Composite CHECK 在上游集合上保留 kiro。历史 301 SQL 不更改。迁移按既有 runner 事务执行，上线前备份并在测试数据库核验；不自动执行。若上线后回退，须评估新平台行、优惠订单和新 schema，不能仅降级二进制。
密码重置改哈希存储/单次原子消费，旧未使用链接升级后需重新申请。已有管理员不受新装校验影响。EasyPay 只接受标准通知参数，return_url 去除客户端查询参数；XorPay 回调独立保留。

## 验证与回退
先只引入上游 EasyPay 安全回归，在旧实现上观察预期失败，再合入修复。合并后运行官方新增测试与现有 Fork 回归，针对充值页面和平台兼容补测试。运行后端 go test ./...、go test -tags=unit ./...、go build ./...、go vet ./...；前端 pnpm run test:run/typecheck/lint:check/build。构建不启动真实服务。代码以未暂存差异交付，可逐文件审查回退。
