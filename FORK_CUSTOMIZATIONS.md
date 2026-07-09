# sub2api Fork 自定义功能清单（索引）

> **完整文档已迁移至** [`docs/FORK_VS_UPSTREAM.md`](docs/FORK_VS_UPSTREAM.md)  
> **AI 改代码前必读** [`AGENTS.md`](AGENTS.md)

本文件保留作快捷入口；详细文件列表、迁移红线、合并流程请以 **`docs/FORK_VS_UPSTREAM.md`** 为准。

---

## 一句话

本仓库 = [Wei-Shaw/sub2api](https://github.com/Wei-Shaw/sub2api) **v0.1.147** + **Kiro** + **XorPay** + **Access Ban** + **Grok/Kiro 配额** + **UI/支付/版本弹窗** 等定制（相对官方 **300** 已提交文件差异；工作区另见文档 §8.2）。

## 核心模块

| 模块 | 详见文档章节 |
|------|-------------|
| Kiro 平台 | §3 |
| XorPay 支付 | §4 |
| Grok 配额口径 | §5 |
| UI / 品牌 / VersionBadge | §6 |
| Ops 删错误日志 | §7 |
| **Access Ban 全局封禁** | **§8** |
| 迁移 157 红线 | §9 |
| 冲突高危文件 | §10 |

## 同步上游

见 `docs/FORK_VS_UPSTREAM.md` §14 — merge 时 **kiro 在前、grok 在后**，**157 须含 kiro+grok**，**159/160 Access Ban 不可删**。

---

*最后更新：2026-07-09*
