-- Align Kiro cache emulation ratios and metering tokenizer with nianzs/sub2api.
-- Historical platform CHECK constraints stay dropped; Kiro remains registered in
-- domain/platforms.go. Channel monitor providers stay on the official set.

ALTER TABLE groups
  ADD COLUMN IF NOT EXISTS kiro_cache_emulation_mode VARCHAR(16) NOT NULL DEFAULT 'uniform',
  ADD COLUMN IF NOT EXISTS kiro_cache_creation_emulation_ratio DECIMAL(5,4),
  ADD COLUMN IF NOT EXISTS kiro_cache_read_emulation_ratio DECIMAL(5,4);

UPDATE groups
SET
  kiro_cache_emulation_mode = 'uniform',
  kiro_cache_creation_emulation_ratio = kiro_cache_emulation_ratio,
  kiro_cache_read_emulation_ratio = kiro_cache_emulation_ratio
WHERE kiro_cache_creation_emulation_ratio IS NULL
   OR kiro_cache_read_emulation_ratio IS NULL;

ALTER TABLE groups
  ALTER COLUMN kiro_cache_creation_emulation_ratio SET DEFAULT 1.0,
  ALTER COLUMN kiro_cache_creation_emulation_ratio SET NOT NULL,
  ALTER COLUMN kiro_cache_read_emulation_ratio SET DEFAULT 1.0,
  ALTER COLUMN kiro_cache_read_emulation_ratio SET NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'groups_kiro_cache_emulation_mode_valid'
  ) THEN
    ALTER TABLE groups
      ADD CONSTRAINT groups_kiro_cache_emulation_mode_valid
      CHECK (kiro_cache_emulation_mode IN ('uniform', 'independent'));
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'groups_kiro_cache_creation_emulation_ratio_range'
  ) THEN
    ALTER TABLE groups
      ADD CONSTRAINT groups_kiro_cache_creation_emulation_ratio_range
      CHECK (kiro_cache_creation_emulation_ratio >= 0 AND kiro_cache_creation_emulation_ratio <= 1);
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'groups_kiro_cache_read_emulation_ratio_range'
  ) THEN
    ALTER TABLE groups
      ADD CONSTRAINT groups_kiro_cache_read_emulation_ratio_range
      CHECK (kiro_cache_read_emulation_ratio >= 0 AND kiro_cache_read_emulation_ratio <= 1);
  END IF;
END $$;

INSERT INTO settings (key, value, updated_at)
VALUES ('gateway_kiro_tokenizer', 'official', NOW())
ON CONFLICT (key) DO NOTHING;
