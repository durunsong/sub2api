import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import MaskEditor from '../MaskEditor.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

describe('MaskEditor', () => {
  it('does not submit an edit without instructions or a marked region', async () => {
    const wrapper = mount(MaskEditor, {
      props: { show: true, blob: new Blob(['image'], { type: 'image/png' }) },
      global: { stubs: { BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }, BlobImage: true } }
    })
    await wrapper.get('[data-testid="mask-submit"]').trigger('click')
    expect(wrapper.emitted('confirm')).toBeUndefined()
    expect(wrapper.text()).toContain('imageStudio.maskRequirements')
    await wrapper.get('textarea').setValue('Turn the marked region red')
    await wrapper.get('[data-testid="mask-submit"]').trigger('click')
    expect(wrapper.emitted('confirm')).toBeUndefined()
    wrapper.unmount()
  })
})
