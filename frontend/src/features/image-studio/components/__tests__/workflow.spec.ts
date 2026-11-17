import { describe, expect, it } from 'vitest'
import { extractVariables, renderTemplate, splitSeries } from '../workflow'

describe('local workflow templates', () => {
  it('extracts each variable once and preserves user values literally', () => {
    expect(extractVariables('{{ product }} / {{场景}} / {{product}}')).toEqual(['product', '场景'])
    expect(renderTemplate('{{ product }} at {{场景}}', { product: '$& mug', 场景: '海边' })).toBe('$& mug at 海边')
  })
  it('splits manually reviewed series on a separator without breaking multiline prompts', () => {
    expect(splitSeries(' one\nline two\n---\n\nsecond\n---\n')).toEqual(['one\nline two', 'second'])
  })
})
