import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import ResultBatch from '../ResultBatch.vue'
import { groupResults } from '../../batches'
import type { ImageHistory } from '../../types'
vi.mock('../../useImageStudio', () => ({ downloadImage: vi.fn() }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
const record = (id: string, index: number): ImageHistory => ({ id, createdAt: 1, batch: { id: 'batch', index, total: 3 }, snapshot: { groupID: 1, model: 'image', prompt: 'Test', size: 'auto', quality: 'auto', count: 1 }, images: [new Blob(['png'], { type: 'image/png' })], references: [], status: 'success', favorite: false, categoryID: null })
function render() {
  return mount(ResultBatch, { props: { group: groupResults([record('a', 0), record('b', 1)], [])[0] }, slots: { default: '<p data-individual>individual actions</p>' }, global: { stubs: { BlobImage: true, Icon: true, BaseDialog: { props: ['show'], template: '<div v-if="show"><slot/><slot name="footer"/></div>' } } } })
}
describe('result batch interactions', () => {
  it('starts as a fan and exposes individual controls only on explicit expansion', async () => {
    const wrapper = render()
    expect(wrapper.findAll('.fan-sheet')).toHaveLength(2)
    expect(wrapper.find('[data-individual]').exists()).toBe(false)
    await wrapper.get('.batch-fan').trigger('click')
    expect(wrapper.get('.batch-fan').attributes('aria-expanded')).toBe('true')
    expect(wrapper.find('[data-individual]').exists()).toBe(true)
    await wrapper.get('.batch-collapse').trigger('click')
    expect(wrapper.find('[data-individual]').exists()).toBe(false)
    wrapper.unmount()
  })
  it('offers image selection and prunes images deleted elsewhere while open', async () => {
    const wrapper = render()
    await wrapper.findAll('button').find(button => button.text().includes('imageStudio.batchDownload'))!.trigger('click')
    expect(wrapper.emitted('dialog')?.at(-1)).toEqual([true])
    expect(wrapper.findAll('[data-download-choice]')).toHaveLength(2)
    await wrapper.findAll('button').find(button => button.text() === 'imageStudio.clearSelection')!.trigger('click')
    expect(wrapper.get('[data-download-selected]').attributes('disabled')).toBeDefined()
    await wrapper.findAll('[data-download-choice]')[0].setValue(true)
    expect(wrapper.get('[data-download-selected]').attributes('disabled')).toBeUndefined()
    await wrapper.setProps({ group: groupResults([record('b', 1)], [])[0] })
    expect(wrapper.findAll('[data-download-choice]')).toHaveLength(1)
    expect(wrapper.get('[data-download-selected]').attributes('disabled')).toBeDefined()
    wrapper.unmount()
  })
})
