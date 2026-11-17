import { mount, flushPromises } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import WorkflowLibrary from '../WorkflowLibrary.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
const transport = vi.hoisted(() => ({ get: vi.fn(), post: vi.fn() }))
vi.mock('@/api/client', () => ({ apiClient: transport }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ user: { id: 7 } }) }))

describe('WorkflowLibrary', () => {
  it('reviews full AI series prompts without prepending the template or starting images', async () => {
    transport.get.mockResolvedValue({ data: { enabled: true, models: [{ id: 'text-model', group_id: 7, group_name: 'Text', model: 'gpt-4', label: 'GPT', supports_references: false }] } })
    transport.post.mockResolvedValue({ data: { prompts: ['AI front view', 'AI detail view'] } })
    const wrapper = mount(WorkflowLibrary, {
      props: { show: true, canRun: true, prompts: [{ id: 'one', title: 'Series', prompt: 'A tea cup', createdAt: 1 }] },
      global: { stubs: { BaseDialog: { template: '<div><slot /></div>' }, ConfirmDialog: true, Select: true, Icon: true } }
    })
    const button = (label: string) => wrapper.findAll('button').find(item => item.text() === label)!
    await button('imageStudio.useTemplate').trigger('click')
    await button('imageStudio.aiSeries').trigger('click')
    await flushPromises()
    await wrapper.get('input[type="number"]').setValue(2)
    await button('imageStudio.aiGenerateSeries').trigger('click')
    await flushPromises()
    expect(wrapper.findAll('textarea').map(item => item.element.value)).toEqual(['AI front view', 'AI detail view'])
    expect(wrapper.emitted('run')).toBeUndefined()
    await wrapper.findAll('textarea')[1].setValue('Reviewed detail view')
    await button('imageStudio.generateSeries').trigger('click')
    expect(wrapper.emitted('run')?.[0]).toEqual([['AI front view', 'Reviewed detail view']])
    wrapper.unmount()
  })
  it('creates an AI template for editing and saves only after explicit review', async () => {
    transport.get.mockResolvedValue({ data: { enabled: true, models: [{ id: 'text-model', group_id: 7, group_name: 'Text', model: 'gpt-4', label: 'GPT', supports_references: false }] } })
    transport.post.mockResolvedValue({ data: { draft: { title: 'AI product shoot', prompt: '{{product}} in daylight', category: 'Product', tags: ['photo'], seriesPrompts: ['Front', 'Side'] } } })
    const wrapper = mount(WorkflowLibrary, {
      props: { show: true, canRun: true, prompts: [] },
      global: { stubs: { BaseDialog: { template: '<div><slot /></div>' }, ConfirmDialog: true, Select: true, Icon: true } }
    })
    const button = (label: string) => wrapper.findAll('button').find(item => item.text() === label)!
    expect(button('imageStudio.aiCreate')).toBeDefined()
    await button('imageStudio.aiCreate').trigger('click')
    await flushPromises()
    await wrapper.get('textarea').setValue('Create a reusable product campaign')
    await button('imageStudio.aiCreateDraft').trigger('click')
    await flushPromises()
    expect(wrapper.findAll('input').some(input => input.element.value === 'AI product shoot')).toBe(true)
    expect(wrapper.emitted('save')).toBeUndefined()
    expect(wrapper.emitted('run')).toBeUndefined()
    await wrapper.findAll('textarea')[0].setValue('{{product}} in reviewed light')
    await button('imageStudio.saveTemplate').trigger('click')
    expect(wrapper.emitted('save')?.[0][0]).toMatchObject({ title: 'AI product shoot', prompt: '{{product}} in reviewed light', seriesPrompts: ['Front', 'Side'] })
    expect(wrapper.emitted('run')).toBeUndefined()
    wrapper.unmount()
  })
  it('requires variables and lets the user review a series before requesting generation', async () => {
    const wrapper = mount(WorkflowLibrary, {
      props: { show: true, canRun: true, prompts: [{ id: 'template', title: 'Product photos', prompt: '{{product}} on a table', seriesPrompts: ['Warm daylight', 'Cool moonlight'], createdAt: 1 }] },
      global: { stubs: { BaseDialog: { template: '<div><slot /></div>' }, ConfirmDialog: true, Select: true, Icon: true } }
    })
    const button = (label: string) => wrapper.findAll('button').find(item => item.text() === label)!
    await button('imageStudio.useTemplate').trigger('click')
    expect(button('imageStudio.prepareSeries').attributes('disabled')).toBeDefined()
    await wrapper.get('input').setValue('Tea cup')
    await button('imageStudio.prepareSeries').trigger('click')
    expect(wrapper.findAll('textarea').map(item => item.element.value)).toEqual(['Tea cup on a table\n\nWarm daylight', 'Tea cup on a table\n\nCool moonlight'])
    expect(wrapper.emitted('run')).toBeUndefined()
    await wrapper.findAll('textarea')[1].setValue('A reviewed composition')
    await button('imageStudio.generateSeries').trigger('click')
    expect(wrapper.emitted('run')?.[0]).toEqual([['Tea cup on a table\n\nWarm daylight', 'A reviewed composition']])
    wrapper.unmount()
  })
})
