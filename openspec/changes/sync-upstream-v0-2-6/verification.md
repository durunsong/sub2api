# v0.2.6 同步验证记录

## 固定范围

- Fork 起点：`ee11c1aa6d3e15e9382506b8d99d17d594da0f09`，工作区起初干净。
- 官方旧基线：v0.2.5 `86f93c28ee34cc74b629dafb748bd5ac5ca8c5ea`。
- 官方目标：v0.2.6 tag object `8f35716ba69976fc0ec950894114d2604bb5454c`，commit `49a39b6dc1abed30fd227611e8af1108bc427610`。
- 精确增量 60 commits / 132 files / +4,836 / -304，没有带入 tag 后的 main。
- 官方 VERSION 仍为 0.2.5，Fork 设为 0.2.6，版本解析脚本返回 0.2.6。

## 合并与定制保留

采用旧 tag、Fork HEAD、目标 tag 在临时目录三方预演，再应用到原工作树，HEAD/暂存区不动。39 个重叠路径、7 个文本冲突，逐块处理。
131 个允许路径纳入；93 个与官方目标字节一致，38 个保留版本/Fork/测试兼容差异。一个监控测试路径保留 Fork 已有动态平台数量断言，因此本次对 HEAD 无代码差异。
本次不涉及的 551 个可读取 Fork 文件在更新说明文档前均与初始 HEAD 字节一致；最终其中三份说明文档更新版本，其余 548 个保持原样。298 个历史 SQL 全部字节一致，无新增 SQL。独立只读审查另行确认 495 个非重叠后端/前端 Fork 文件不变。

- Kiro OAuth、credits/cache_read、独立 429 冷却、真实上游 ID、精简 DTO 六个状态字段与 Wire 依赖保留；XorPay、Access Ban 和 admin 登录失败自动封禁不变。
- 订阅重置卡的期限快照、来源幂等、退款事务、过期重开、消费与旧接口、兼容镜像、提示和 active_available 过滤不变。
- 提示词审计、Ops 删除、Claude 用户端别名、GLM 分类、支付金额分离/隐藏倍率、品牌、ConfirmDialog/Select、VersionBadge 禁止在线升级保留。
- 新 Wire SettingService 追加到原 Fork 参数；账号导出测试保留 Kiro 导入格式保护和官方票据脱敏回归。
- 金额输入采纳非法输入恢复，同时保持人民币、快捷金额与 max 校验；customText 仅在验证通过后更新，避免超额值成为恢复目标。
- gRPC 升至 1.83.2，x/crypto 与 x/net 等采用上游安全升级；保留 Fork tokenizer、x/image 0.45.0 与 direct x/text，未新增无关依赖。
- 新 Codex 票据默认关闭，后台开关/采集代理热更新、票据摘要、托管状态过滤与生命周期按上游合入；未启用实际采集。兑换分页保持当前用户隔离、全部兑换类型、稳定排序及旧无分页参数时的数组响应。
- 独立只读审查未发现具体兼容缺陷或定制丢失；该审查不替代运行时验证。

## 测试驱动

先将官方金额输入行为测试接入 Fork 原有测试：旧代码 8 项中 5 项预期失败（非法字符/超两位小数不能恢复）；Fork 金额上限测试和小数编辑测试通过。合入兼容实现后，同一文件 8 项全部通过。

## 已执行验证

Go 命令在 backend，pnpm 命令在 frontend，其余在项目根执行。

| 命令 | 结果 |
| --- | --- |
| `pnpm exec vitest run src/components/payment/__tests__/AmountInput.spec.ts` | 旧实现 5 fail / 3 pass；合入后 8 pass，退出 0 |
| `pnpm run test:run --maxWorkers=4 --minWorkers=2` | 312 文件 / 2348 项通过，退出 0 |
| `pnpm run typecheck && pnpm run lint:check && pnpm run build` | 全部通过，退出 0；构建含 i18n 完整性校验 |
| `go test ./... -count=1` | 52 个测试包通过，退出 0 |
| `go vet ./... && go build ./...` | 通过，退出 0 |
| `go test -tags=unit -p 1 ./... -count=1` | 60 个测试包通过，退出 0 |
| `sh backend/scripts/resolve-version.sh` | 0.2.6，退出 0 |
| `git diff --check`、新增文件 no-index whitespace、变更 Go 文件 `gofmt -l` | 无诊断 |

## 文件检查与限制

最终 138 个变更文本文件均为 UTF-8，无 BOM、异常替换字符、合并冲突标记或新增绝对个人路径。OpenSpec proposal/design/spec/tasks 的路径与 requirement/scenario 结构已检查；最终文档落盘后再次复核通过。

- `deploy/config.example.yaml` 属受保护配置，未读未改；对应新功能代码、默认值、管理端设置已经同步，高级部署配置上线前单独核对。
- 无 Docker/PostgreSQL 可执行文件，未运行数据库 integration；没有执行真实迁移或部署。无新增 SQL，历史 298 个 SQL 未变。
- 未调用真实 OAuth、模型供应商、票据采集或支付网关，自动化检查不等同于真实外部集成验证。
- 前端构建存在大 chunk 与 Browserslist 数据陈旧提示，未阻断构建；部分既有测试输出预期异常和 i18n runtime 警告，最终测试均通过。
- OpenSpec CLI 和 CodeGraph 不可用，沿用现有规划结构与 Git/rg；未安装额外工具，未运行 CLI validate。
- 未执行 git add/commit/push、修改秘密配置或进行生产操作。
