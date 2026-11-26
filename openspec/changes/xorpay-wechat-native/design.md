# 设计与边界

依据：https://xorpay.com/htdocs/xorpay.md（2026-11-26 读取）。

选择扩展现有 XorPay provider，而非复制独立微信 provider：下单、签名、回调、查询和退款协议相同，现有订单已经保存 provider_key 与 provider_instance_id，可区分微信官方与 XorPay。

- xorpay 支付类型继续发送 alipay，wxpay 发送 native。未知类型在请求上游前拒绝。二维码内容按 info.qr 原样返回。
- 金额延续 decimal 格式化为两位小数字符串，以同一字符串参与签名和发送；币种 CNY，平台到账金额与支付金额仍由现有服务计算。
- 管理端 XorPay 可选 xorpay、wxpay；已有实例不自动增加 wxpay。当前设置页用实例支付类型开关，沿用微信通道冲突检查，不新增来源下拉框。后端及配置 helper 的微信来源增加 xorpay_wxpay（兼容别名 xorpay），兼容通过 API 配置多 provider 按来源选择；全局开关、限额与现有行为一致。
- 用户端微信订单继续使用 wxpay，复用微信图标、扫码提示、订单恢复和轮询；无需新增展示类型。微信内同样提供二维码，不走官方 OpenID/JSAPI 引导。
- 统一回调地址 /api/v1/payment/webhook/xorpay。复用验签、金额一致性校验、订单原子状态更新与既有履约幂等；查询无金额，不能绕过回调金额校验。退款复用原 aoid 与金额签名。
- XorPay 的 SupportedTypes 声明微信能力，但全局 registry 只注册历史 xorpay 别名；新增微信按实例路由，避免覆盖没有实例快照的旧官方微信订单查询/验签归属。独立审查定位此风险，使用两种服务商加载顺序复现后修复。
- 不读取生产配置或密钥；测试仅用虚构凭据、本地 HTTP stub 与内存 SQLite。appSecret 不进入用户响应或订单快照。
- 不做 JSAPI/公众号 OAuth、小程序、线下付款码、自动启用、生产发布或历史订单迁移。回退时先停用 XorPay 微信新下单并处理完在途订单，再回退应用。

本机未安装 openspec CLI，按仓库现有目录结构手工记录；不安装依赖。用户已授权完成对接，按项目规则直接实施技术上明确的扫码扩展。
