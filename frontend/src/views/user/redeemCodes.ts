export const MAX_REDEEM_BATCH = 50

const CODE_SPLIT = /[\s,;，；]+/

/** Split a pasted block into unique redeem codes, preserving order and case. */
export function parseRedeemCodes(raw: string): string[] {
  const seen = new Set<string>()
  const codes: string[] = []
  for (const line of raw.split(/\r?\n/)) {
    const cleaned = line.replace(/^\s*(?:[-*•]|\d+[.)、])\s+/, '')
    for (const part of cleaned.split(CODE_SPLIT)) {
      const code = part.trim().replace(/^['"“”‘’`]+|['"“”‘’`]+$/g, '')
      if (!code || code.length > 128 || seen.has(code)) continue
      seen.add(code)
      codes.push(code)
    }
  }
  return codes
}
