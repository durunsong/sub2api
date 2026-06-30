-- 把 grok 平台加入 user_platform_quotas.platform 的 CHECK 约束。
--
-- 背景：grok 自 2026-06 起进入默认平台配额（default_platform_quotas /
-- auth_source_default_*_platform_quotas），但 142 建表时的 CHECK 仅允许
-- anthropic/openai/gemini/antigravity。自助注册时 snapshotPlatformQuotaDefaults
-- 会写入 grok 默认配额行 → 违反 CHECK → 整个注册事务被标记 aborted →
-- OAuth pending 路径 consume 会话时撞 "transaction aborted" → 500 → 清 cookie → 404。
--
-- 修复：把约束与代码平台列表（internal/domain/constants.go 的 PlatformGrok）对齐。
-- DROP ... IF EXISTS 保证可重入；新约束是旧约束的超集，存量行（仅 4 平台）瞬时校验通过。
--
-- 注意（本 fork）：145 迁移已将 kiro 加入该 CHECK 约束。157 在 145 之后执行且
-- 会 DROP 重建约束，若只列 grok 会丢掉 kiro，导致写入 platform='kiro' 配额行时
-- 违反 CHECK。故此处保留 kiro，约束为 4 基础平台 + kiro + grok。
ALTER TABLE user_platform_quotas
    DROP CONSTRAINT IF EXISTS user_platform_quotas_platform_check;

ALTER TABLE user_platform_quotas
    ADD CONSTRAINT user_platform_quotas_platform_check
    CHECK (platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'kiro', 'grok'));
