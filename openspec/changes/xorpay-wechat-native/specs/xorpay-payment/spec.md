## ADDED Requirements

### Requirement: XorPay 微信扫码
系统 SHALL 将 XorPay 的 wxpay 订单以 native 类型、两位小数字符串金额及固定顺序 MD5 签名提交，并保留 xorpay 支付宝行为。

#### Scenario: 微信扫码下单
- **WHEN** 用户通过启用微信的 XorPay 实例创建 wxpay 订单
- **THEN** 发送 native，返回微信二维码，不触发官方微信 OAuth。

#### Scenario: 支付宝兼容
- **WHEN** 原 xorpay 订单下单
- **THEN** 仍发送 alipay，历史订单含义不变。

### Requirement: 配置与安全闭环
系统 SHALL 提供 XorPay 微信来源，保持实例筛选、限额、回调验签、金额一致性与履约幂等。

#### Scenario: 多通道选择
- **WHEN** 微信官方、易支付与 XorPay 同时启用且来源为 xorpay_wxpay
- **THEN** 下单及展示限额使用 XorPay；未启用微信的 XorPay 不参与。

#### Scenario: 回调重复或金额不符
- **WHEN** XorPay 微信收到重复回调或错误金额
- **THEN** 重复回调不重复到账，错误金额不完成订单；签名被修改时拒绝。
