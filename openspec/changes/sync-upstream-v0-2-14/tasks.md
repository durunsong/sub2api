# 实施任务

依据 proposal.md、design.md、specs/upstream-sync/spec.md；按已授权范围直接执行，禁止自动提交/部署。

- [x] 固定 release SHA，清点上游 176 路径、80 重叠路径和 27 冲突文件，排除 8 个部署路径。
- [x] 导入 easypay_notify_security_test.go，运行 go test ./internal/payment/provider -run TestEasyPayNotify -count=1 观察旧实现拒绝伪造失败。
- [x] 三方合并 168 个路径，解决 Kiro/TypeSafe 枚举、Access Ban 路由、充值 UI、依赖冲突；新增 241 平台迁移保留 Kiro。
- [x] 运行支付安全/促销、平台/路由/订阅定制回归，修复合并兼容问题并保留必要测试。
- [x] 后端 default/unit 测试、build、vet、lint；前端 test、typecheck、lint、build；核验新增依赖锁文件。
- [x] 审查全部定制保留与 SQL 不变，更新 VERSION、AGENTS、Fork 两份说明及 verification.md。
- [x] 最终 diff/UTF-8/BOM/冲突标记/路径检查，报告实际结果与上线限制。
