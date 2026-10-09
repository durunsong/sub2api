# 同步官方 v0.2.15

## 目标与范围
用户已授权将 v0.2.14 Fork 升级至官方 release v0.2.15，保留全部定制。基线 0363b8cdba8cec3e2ba4b2dbd49c4481143fa55d，目标 f2669c8cf62555cd92389b3f55920e9e6e7c6ff2；Fork 起点 a089395d22be496472f20ba295085d5381813c57。官方增量 146 commits / 278 files / +19,277 / -1,241。

引入平台目录、Command Code/Cline、多协议路由与模型发现、Anthropic/Responses/WS 修复、监控 TPS 和前端竞态修复。已有 Go 依赖与构建工具链跟随安全升级，保留 Fork 较新的 x/image。

## 非目标
不部署、启动服务、执行迁移或真实支付，不读取秘密配置，不提交、推送或创建 PR，不合入 release 之后的 main。

## 验收
上游增量完整并存于 Fork；Kiro/XorPay/Access Ban/提示词审计/永久重置卡/active_available/Ops/UI 定制回归通过；历史 SQL 字节不变。后端 default/unit、build/vet 和前端 test/typecheck/lint/build 验证；最终差异、UTF-8、BOM、冲突标记检查。
