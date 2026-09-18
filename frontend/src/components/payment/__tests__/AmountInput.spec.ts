import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import { createI18n } from 'vue-i18n'
import AmountInput from '../AmountInput.vue'

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackWarn: false,
  missingWarn: false,
  messages: {
    en: {
      payment: {
        customAmount: 'Custom amount',
        enterAmount: 'Enter amount',
        quickAmounts: 'Quick amounts',
      },
    },
  },
})

function mountAmountInput(props: Partial<InstanceType<typeof AmountInput>['$props']> = {}) {
  return mount(AmountInput, {
    props: {
      modelValue: null,
      ...props,
    },
    global: { plugins: [i18n] },
  })
}

describe('AmountInput', () => {
  it('defaults quick amounts to 1000 max', () => {
    const wrapper = mountAmountInput()

    expect(wrapper.text()).toContain('1000')
    expect(wrapper.text()).not.toContain('2000')
    expect(wrapper.text()).not.toContain('5000')
  })

  it('ignores custom amount above max', async () => {
    const wrapper = mountAmountInput({ max: 1000 })
    const input = wrapper.get('input')

    await input.setValue('1000')
    await input.setValue('1001')

    expect((input.element as HTMLInputElement).value).toBe('1000')
    expect(wrapper.emitted('update:modelValue')).toEqual([[1000]])
  })
})

describe('recharge amount input', () => {
  it.each(['10abc', '10.555', '-10', '1e2'])('restores the accepted amount after rejecting %s', async (value) => {
    const wrapper = mountAmountInput({ modelValue: 10 })
    const input = wrapper.get('input')
    await input.setValue(value)
    expect((input.element as HTMLInputElement).value).toBe('10')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })

  it('restores the last typed amount rather than a stale prop', async () => {
    const wrapper = mountAmountInput()
    const input = wrapper.get('input')
    await input.setValue('12.50')
    await input.setValue('12.500')
    expect((input.element as HTMLInputElement).value).toBe('12.50')
    expect(wrapper.emitted('update:modelValue')).toEqual([[12.5]])
  })

  it('preserves decimal editing and allows clearing the amount', async () => {
    const wrapper = mountAmountInput()
    const input = wrapper.get('input')
    for (const value of ['0', '0.', '0.5', '0.50', '']) await input.setValue(value)
    expect(wrapper.emitted('update:modelValue')).toEqual([[null], [null], [0.5], [0.5], [null]])
    expect((input.element as HTMLInputElement).value).toBe('')
  })
})
