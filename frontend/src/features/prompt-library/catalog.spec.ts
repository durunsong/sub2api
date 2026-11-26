import { expect, it } from 'vitest'
import { normalizePrompts, filterPrompts, safeURL, sources } from './catalog'
it('preserves upstream text and tags, deduplicates source IDs and rejects executable URLs', () => {
  const items = normalizePrompts([{id:'a',title:'海报',prompt:'Exact {{text}}',tags:['工作'],coverUrl:'javascript:alert(1)'},{id:'a',title:'duplicate',prompt:'ignore'},null,{title:'missing'}],sources[0])
  expect(items).toHaveLength(1)
  expect(items[0]).toMatchObject({prompt:'Exact {{text}}',tags:['工作'],coverUrl:''})
  expect(safeURL('/a.png',sources[0].url)).toBe('https://raw.githubusercontent.com/a.png')
  expect(safeURL('data:text/html,hello',sources[0].url)).toBe('')
})
it('combines source and search filters with upstream OR tag selection', () => {
  const items=normalizePrompts([{title:'海报',prompt:'Apple',tags:['工作']},{title:'风景',prompt:'Lake',tags:['自然']}],sources[0])
  expect(filterPrompts(items,sources[0].id,'apple',['工作','自然'])).toHaveLength(1)
  expect(filterPrompts(items,sources[1].id,'',[])).toHaveLength(0)
  expect(filterPrompts(items,'','',[])).toHaveLength(2)
})
