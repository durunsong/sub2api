import { flushPromises, mount } from '@vue/test-utils'
import { expect, it, vi } from 'vitest'
import PromptLibraryDialog from './PromptLibraryDialog.vue'
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('./catalog', async original => {
  const actual = await original<typeof import('./catalog')>()
  return {...actual,loadSource: vi.fn(async (source: typeof actual.sources[number]) => ({items:actual.normalizePrompts([{id:'a',title:source.name,prompt:'Exact prompt',tags:['work']}],source),stale:false}))}
})
it('defaults to the upstream catalog and selects exact text without generating', async () => {
 const wrapper=mount(PromptLibraryDialog,{props:{show:true},slots:{custom:'<input value="my private prompt" />'},global:{stubs:{BaseDialog:{template:'<div><slot /></div>'},Icon:true}}})
 await flushPromises()
 expect(wrapper.get('[role="tab"][aria-selected="true"]').text()).toBe('promptLibrary.upstream')
 expect(wrapper.findAll('.prompt-card')).toHaveLength(7)
 await wrapper.get('.prompt-card button').trigger('click')
 expect(wrapper.emitted('use')).toEqual([['Exact prompt']])
 expect(wrapper.emitted('close')).toHaveLength(1)
 await wrapper.findAll('[role="tab"]')[1].trigger('click')
 expect(wrapper.get('input').element.value).toBe('my private prompt')
 wrapper.unmount()
})
