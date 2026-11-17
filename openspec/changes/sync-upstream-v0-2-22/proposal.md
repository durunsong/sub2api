# 同步上游 v0.2.22 并保留 Fork 定制

## Why
本地基线为 v0.2.21，需同步官方 v0.2.22 的功能与修复，继续保留既有定制。

## What Changes
- 精确整合官方 10ba8f9b4dc2e6023d4a9879ab77b21841babd10 → b0fa9f3e6eeab1968a53e1289b69daa4bb40ca2b（187 commits，340 路径，+20708/-1652）。
- Ultra Fast、Vertex 多区域/Gemini Carbon、透传模型过滤、并发扣费/调度/JSON 优化及发布内全部修复。
- 保留 Kiro/XorPay/Access Ban、全期订阅额度/永久重置卡/重置历史、邮件券、品牌与所有用户端定制。

## Capabilities
### Modified Capabilities
- upstream-sync: 发布标签三方整合与定制兼容。

## Impact
后端、前端、ORM/DI、构建工具链和两份新增 SQL。版本为 0.2.22。生产配置、历史 SQL、真实数据、提交/推送/部署不在范围内。
