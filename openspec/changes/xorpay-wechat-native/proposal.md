# XorPay 微信扫码支付

## Why

已有 XorPay 支付宝支付可用，但 provider 把 pay_type 固定为 alipay，管理端不能为 XorPay 启用微信。按现有网页扫码场景补齐微信支付。

## What Changes

- 同一 XorPay provider 支持原有 xorpay（支付宝）及 wxpay（微信 native）。
- 微信来源增加 xorpay_wxpay；沿用实例配置、路由、限额及订单快照。
- 保留原支付宝订单、验签、到账幂等、退款和用户端展示。

## Impact

Go provider、支付来源与实例筛选、Vue 管理端选项和回归测试。无新依赖、迁移、真实支付或部署操作。
