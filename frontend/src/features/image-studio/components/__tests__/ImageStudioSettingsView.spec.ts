import { mount, flushPromises } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import ImageStudioSettingsView from '@/views/admin/ImageStudioSettingsView.vue'

const mocks = vi.hoisted(() => ({ save: vi.fn(async config => config) }))
vi.mock('@/components/layout/AppLayout.vue', () => ({ default: { template: '<div><slot /></div>' } }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/api/admin/groups', () => ({ getAll: async () => [{ id: 7, name: 'Text group' }] }))
vi.mock('../../api', () => ({ imageStudioAPI: {
  config: async () => ({ enabled: true, models: [] }),
  adminCatalog: async () => ({ enabled: true, models: [] }),
  adminWorkflowModels: async () => ({ enabled: true, models: [{ id: '7:text', group_id: 7, group_name: 'Text group', model: 'gpt-4', label: 'GPT', supports_references: true }] }),
  saveConfig: mocks.save
} }))

describe('studio text model settings', () => {
  it('lets admins hide a discovered text model without giving it an image protocol', async () => {
    const wrapper = mount(ImageStudioSettingsView, { global: { stubs: { AppLayout: { template: '<div><slot /></div>' }, RouterLink: true, BaseDialog: true, Select: true, Icon: true } } })
    await flushPromises()
    expect(wrapper.text()).toContain('gpt-4')
    await wrapper.get('input[aria-label="studioAdmin.visible gpt-4"]').setValue(false)
    await wrapper.findAll('button').find(button => button.text() === 'common.save')!.trigger('click')
    await flushPromises()
    expect(mocks.save).toHaveBeenCalledWith({ enabled: true, models: [{ group_id: 7, model: 'gpt-4', label: 'gpt-4', enabled: false }] })
    wrapper.unmount()
  })
})
