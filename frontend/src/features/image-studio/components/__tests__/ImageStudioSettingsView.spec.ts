import { mount, flushPromises } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ImageStudioSettingsView from '@/views/admin/ImageStudioSettingsView.vue'

const mocks = vi.hoisted(() => ({ save: vi.fn(async config => config), settings: vi.fn() }))
vi.mock('@/components/layout/AppLayout.vue', () => ({ default: { template: '<div><slot /></div>' } }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/api/admin/groups', () => ({ getAll: async () => [{ id: 7, name: 'Text group' }] }))
vi.mock('../../api', () => ({ imageStudioAPI: {
  config: mocks.settings,
  adminCatalog: async () => ({ enabled: true, models: [] }),
  adminWorkflowModels: async () => ({ enabled: true, models: [{ id: '7:text', group_id: 7, group_name: 'Text group', model: 'gpt-4', label: 'GPT', supports_references: true }] }),
  saveConfig: mocks.save
} }))

const stubs = { AppLayout: { template: '<div><slot /></div>' }, RouterLink: true, BaseDialog: true, Icon: true }
function render() { return mount(ImageStudioSettingsView, { global: { stubs } }) }
beforeEach(() => {
  mocks.save.mockClear()
  mocks.settings.mockResolvedValue({ enabled: true, models: [] })
})
afterEach(() => { vi.useRealTimers() })

describe('studio text model settings', () => {
  it('changes only the filtered models in available groups and persists only on save', async () => {
    mocks.settings.mockResolvedValue({ enabled: true, models: [
      { group_id: 7, model: 'gpt-4', label: 'Creative', enabled: true },
      { group_id: 7, model: 'image-alias', label: 'Retouch', protocol: 'gemini', enabled: false },
      { group_id: 8, model: 'other', label: 'Other group', protocol: 'images', enabled: true },
      { group_id: 99, model: 'old', label: 'Old group', protocol: 'gemini', enabled: true }
    ] })
    const wrapper = render()
    await flushPromises()
    const button = (key: string) => wrapper.findAll('button').find(item => item.text() === key)!
    await wrapper.get('input[aria-label="studioAdmin.search"]').setValue('Creative')
    await button('studioAdmin.hideFiltered').trigger('click')
    expect(mocks.save).not.toHaveBeenCalled()
    expect((wrapper.get('input[aria-label="studioAdmin.visible gpt-4"]').element as HTMLInputElement).checked).toBe(false)
    await wrapper.get('input[aria-label="studioAdmin.search"]').setValue('')
    await wrapper.getComponent({ name: 'Select' }).vm.$emit('update:modelValue', 7)
    await button('studioAdmin.showFiltered').trigger('click')
    await button('common.save').trigger('click')
    await flushPromises()
    expect(mocks.save).toHaveBeenCalledWith({ enabled: true, models: [
      { group_id: 7, model: 'gpt-4', label: 'Creative', enabled: true },
      { group_id: 7, model: 'image-alias', label: 'Retouch', protocol: 'gemini', enabled: true },
      { group_id: 8, model: 'other', label: 'Other group', protocol: 'images', enabled: true },
      { group_id: 99, model: 'old', label: 'Old group', protocol: 'gemini', enabled: true }
    ] })
    wrapper.unmount()
  })
  it('disables bulk actions for no editable results and during saving', async () => {
    mocks.settings.mockResolvedValue({ enabled: true, models: [{ group_id: 99, model: 'old', label: 'Old group', enabled: true }] })
    const wrapper = render()
    await flushPromises()
    const button = (key: string) => wrapper.findAll('button').find(item => item.text() === key)!
    await wrapper.get('input[aria-label="studioAdmin.search"]').setValue('Old group')
    expect(button('studioAdmin.hideFiltered').attributes('disabled')).toBeDefined()
    await wrapper.get('input[aria-label="studioAdmin.search"]').setValue('missing')
    expect(button('studioAdmin.showFiltered').attributes('disabled')).toBeDefined()
    await wrapper.get('input[aria-label="studioAdmin.search"]').setValue('')
    await button('studioAdmin.hideFiltered').trigger('click')
    expect((wrapper.get('input[aria-label="studioAdmin.visible old"]').element as HTMLInputElement).checked).toBe(true)
    expect((wrapper.get('input[aria-label="studioAdmin.visible gpt-4"]').element as HTMLInputElement).checked).toBe(false)
    mocks.save.mockImplementationOnce(() => new Promise(() => {}))
    await button('common.save').trigger('click')
    expect(button('studioAdmin.showFiltered').attributes('disabled')).toBeDefined()
    expect(button('studioAdmin.hideFiltered').attributes('disabled')).toBeDefined()
    wrapper.unmount()
  })
  it('expires success after five seconds, resets on another save and cleans up on unmount', async () => {
    vi.useFakeTimers()
    const wrapper = render()
    await flushPromises()
    const save = () => wrapper.findAll('button').find(item => item.text() === 'common.save')!.trigger('click')
    await save(); await flushPromises()
    expect(wrapper.get('[role="status"]').text()).toBe('studioAdmin.saved')
    await vi.advanceTimersByTimeAsync(4000)
    await save(); await flushPromises()
    await vi.advanceTimersByTimeAsync(4999)
    expect(wrapper.get('[role="status"]').text()).toBe('studioAdmin.saved')
    await vi.advanceTimersByTimeAsync(1)
    expect(wrapper.find('[role="status"]').exists()).toBe(false)
    await save(); await flushPromises()
    wrapper.unmount()
    expect(vi.getTimerCount()).toBe(0)
  })
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
