import { describe, expect, it } from 'vitest'
import { parseRedeemCodes } from '../redeemCodes'

describe('parseRedeemCodes', () => {
  it('splits pasted lines and separators, drops blanks, and keeps the first of each code', () => {
    const raw = [
      '73df2500425a717ab144d904ee6cef94',
      '',
      '  6d4313e592b1d9386f45df0327e3a231  ',
      '73df2500425a717ab144d904ee6cef94',
      '"c50d54d4c818bcb808fd7ec8b30f8400", b99319f0ec86f6efb3efbf42f4795fae；BDED-26F1-C8D4-04AA',
      '1. ac4c89e4f3f4528003e13bacd4bb0d2f',
      '- 4362199b6f87277f7956a10ceaee8c4a',
    ].join('\n')

    expect(parseRedeemCodes(raw)).toEqual([
      '73df2500425a717ab144d904ee6cef94',
      '6d4313e592b1d9386f45df0327e3a231',
      'c50d54d4c818bcb808fd7ec8b30f8400',
      'b99319f0ec86f6efb3efbf42f4795fae',
      'BDED-26F1-C8D4-04AA',
      'ac4c89e4f3f4528003e13bacd4bb0d2f',
      '4362199b6f87277f7956a10ceaee8c4a',
    ])
  })

  it('keeps a single code exactly, including case', () => {
    expect(parseRedeemCodes(' REDEEM-CODE ')).toEqual(['REDEEM-CODE'])
  })

  it('returns nothing for blank input', () => {
    expect(parseRedeemCodes(' \n\t ')).toEqual([])
  })
})
