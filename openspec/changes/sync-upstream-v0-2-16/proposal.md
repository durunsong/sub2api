# 同步官方 v0.2.16

## 目标与范围
用户已授权将 v0.2.15 Fork 升级至官方 release v0.2.16，保留全部定制。基线 `f2669c8cf62555cd92389b3f55920e9e6e7c6ff2`，目标 `01842576b3d5c1e2545b5a99ad555d36e28ed306`；Fork 起点 `0652f07ff`。官方增量 79 commits / 213 files / +13,886 / -727。

引入 Claude Haiku 5.5、Gemini 显式上下文缓存、Vertex API Key 与模型同步、AI Studio `AQ.` Key、临时不可调度作用范围、分组多语言、运维账号列，以及 Responses 回退、图像限流、401 规则和支付倒计时修复。官方 tag 内 VERSION 文件仍为 0.2.15，Fork 设为 0.2.16。

## 非目标
不部署、启动服务、执行迁移或真实支付，不读取秘密配置，不提交、推送或创建 PR，不合入 release 之后的 main。

## 验收
上游增量完整并存于 Fork；Kiro/XorPay/Access Ban/提示词审计/永久重置卡/active_available/Ops/UI 定制回归通过；历史 SQL 字节不变。后端 build/vet/unit 与相关前端测试通过。最终差异、UTF-8、BOM、冲突标记检查。
