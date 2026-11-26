# 验证记录（2026-11-26）

## 实现结果

- 原 `xorpay` 下单仍为支付宝；`wxpay` 下单发送 XorPay `native`，返回二维码，沿用 CNY 两位小数字符串及签名。
- XorPay 实例可启用微信，现有管理页开关及冲突检查接通；后端兼容 `xorpay_wxpay` 来源与限额选择。
- 微信二维码流程复用原有展示与轮询。固定实例回调验签、金额拒绝、成功到账和重复通知幂等均由自动化检查覆盖。
- 官方微信 OAuth 不应用于 XorPay；历史支付宝及未绑定实例的官方微信订单保持原语义。

## 失败到通过的证据

1. 新增 provider/service 测试首先暴露缺少 wxpay 能力、来源归一化及路由；前端 14 项测试中 3 项失败（缺少来源及支持类型）。补齐实现后通过。
2. SettingsView 微信开关测试先复现与其他微信实例冲突时仍发送更新请求；补齐 XorPay 可见方法判断后通过。
3. 独立审查发现全局 registry 的 wxpay 映射可能被 XorPay 覆盖。`TestXorPayRegistrationPreservesLegacyWechatProvider` 复现预期 wxpay、实际 xorpay；限定历史别名注册后，两种加载顺序均通过，包括旧订单 provider 查询和回调归属。
4. 测试编写期间修正了限额 fixture 的 JSON 字段名及 PaymentTradeNo 字段类型；未因此修改业务金额/限额逻辑。

## 实际命令

在 `backend/`：

```sh
go test -tags=unit ./internal/payment/... ./internal/service ./internal/handler ./internal/handler/admin -run 'Payment|XorPay|VisibleMethod|Webhook|Provider|Registry|PcAggregate|PcGroup' -count=1
go test -tags=unit ./internal/service -run 'TestXorPay' -count=1
go build ./...
go vet ./...
```

支付回归覆盖 5 个包，通过；构建及 vet 退出码均为 0。最后补充的单实例无来源/遗留来源配置回归另行运行。

在 `frontend/`：

```sh
pnpm vitest run src/components/payment/__tests__ src/components/admin/payment/__tests__ src/api/__tests__/settings.paymentVisibleMethods.spec.ts src/views/admin/__tests__/SettingsView.spec.ts src/views/user/__tests__/PaymentQRCodeView.spec.ts src/views/user/__tests__/PaymentView.spec.ts src/views/user/__tests__/PaymentResultView.spec.ts src/i18n/__tests__/forkOverlayKeys.spec.ts
pnpm typecheck
pnpm build
pnpm exec eslint src/api/admin/settings.ts src/api/__tests__/settings.paymentVisibleMethods.spec.ts src/components/payment/providerConfig.ts src/components/payment/__tests__/providerConfig.spec.ts src/components/payment/__tests__/paymentFlow.spec.ts src/components/payment/__tests__/PaymentQRDialog.spec.ts src/views/admin/SettingsView.vue src/views/admin/__tests__/SettingsView.spec.ts
```

Vitest 实际匹配 19 个测试文件，246 项全部通过；构建中的 i18n 完整性检查额外 3 项通过。typecheck、build 与修改文件 ESLint 退出码均为 0。命令中的 PaymentQRCodeView.spec.ts 路径没有独立匹配文件，不作为额外覆盖证据；支付流程、二维码弹窗、购买页及结果页由实际匹配测试覆盖。

已有 SettingsView 测试输出 Vue locale mock 和 JSDOM 本地请求警告，但无失败；构建输出 Browserslist 数据过期及静态/动态 import 混用提示。未为这些非本次业务问题升级依赖或改动公共设置页测试设施。新增二维码用例显式 mock Canvas context。

仓库根目录：`git diff --check`，及针对全部修改/新增文件的 Python UTF-8 严格解码、BOM 和替换字符检查。确认文档路径存在，并在 `.gitignore` 中仅增加新配置文档的白名单。

## 交付边界

- 未读取真实商户凭据，未发起真实支付或退款，未部署、迁移、提交或推送。
- HTTP 协议使用本地测试服务器，数据库测试使用内存 SQLite；真实微信商户权限、外网回调和付款到账仍需部署后的小额验收。
- 本次为 Native 扫码，不包含公众号 JSAPI/小程序/手机浏览器直接唤起。
- 本机无 openspec CLI，未执行其验证命令；规划文件按既有目录手工维护。
- 未运行全仓所有测试；运行范围为相关支付测试与前后端构建/静态检查。

操作步骤见 `docs/XORPAY_WECHAT_SETUP.md`。
