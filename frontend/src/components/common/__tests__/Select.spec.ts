import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import Select from '../Select.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

describe('Select', () => {
  it('支持用隐藏搜索关键字匹配展示别名', async () => {
    const wrapper = mount(Select, {
      props: {
        modelValue: null,
        searchable: true,
        options: [
          { value: 1, label: '高级分组', searchKeywords: 'kiro claude' },
          { value: 2, label: 'OpenAI 分组' }
        ]
      },
      attachTo: document.body
    })

    await wrapper.get('button').trigger('click')
    const input = document.body.querySelector<HTMLInputElement>('.select-search-input')
    expect(input).not.toBeNull()
    input!.value = 'claude'
    input!.dispatchEvent(new Event('input', { bubbles: true }))
    await nextTick()

    expect(document.body.textContent).toContain('高级分组')
    expect(document.body.textContent).not.toContain('OpenAI 分组')
    wrapper.unmount()
  })
})
