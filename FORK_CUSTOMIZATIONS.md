# sub2api Fork 自定义功能清单（索引）

> **完整文档已迁移至** [`docs/FORK_VS_UPSTREAM.md`](docs/FORK_VS_UPSTREAM.md)<br>
> **AI 改代码前必读** [`AGENTS.md`](AGENTS.md)

本文件保留作快捷入口；详细文件列表、迁移红线、合并流程请以 **`docs/FORK_VS_UPSTREAM.md`** 为准。

---

## 一句话

本仓库 = [Wei-Shaw/sub2api](https://github.com/Wei-Shaw/sub2api) **v0.1.146** + **Kiro** + **XorPay** + **Grok/Kiro 配额** + **UI/支付/版本弹窗** 等定制（相对官方 **281** 文件差异）。

## 核心模块

| 模块 | 详见文档章节 |
|------|-------------|
| Kiro 平台 | §3 |
| XorPay 支付 | §4 |
| Grok 配额口径 | §5 |
| UI / 品牌 / VersionBadge | §6 |
| Ops 删错误日志 | §7 |
| 迁移 157 红线 | §8 |
| 冲突高危文件 | §9 |

## 同步上游

见 `docs/FORK_VS_UPSTREAM.md` §13 — merge 时 **kiro 在前、grok 在后**，**157 须含 kiro+grok**。

---

*最后更新：2026-07-07*
