import { afterEach, describe, expect, it, vi } from 'vitest'
import { createI18n } from 'vue-i18n'
import { enableAutoUnmount, mount } from '@vue/test-utils'
import { vTooltip } from '../tooltip'

enableAutoUnmount(afterEach)
afterEach(() => { vi.useRealTimers(); document.body.innerHTML = '' })

describe('icon tooltips', () => {
  afterEach(() => { document.body.innerHTML = '' })
  it('describes the actual button on hover/focus, dismisses on Escape and cleans up', async () => {
    const wrapper = mount({ directives: { tooltip: vTooltip }, template: '<button v-tooltip aria-label="下载图片">↓</button>' }, { attachTo: document.body })
    const button = wrapper.get('button')
    await button.trigger('mouseenter')
    const bubble = document.querySelector('[role="tooltip"]')!
    expect(bubble?.textContent).toBe('下载图片')
    expect(button.attributes('aria-describedby')).toBe(bubble.id)
    await button.trigger('mouseleave')
    await new Promise(resolve => setTimeout(resolve, 170))
    expect(document.querySelector('[role="tooltip"]')).toBeNull()
    await button.trigger('focus')
    expect(document.querySelector('[role="tooltip"]')).not.toBeNull()
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    expect(document.querySelector('[role="tooltip"]')).toBeNull()
    await button.trigger('blur')
    await button.trigger('focus')
    wrapper.unmount()
    expect(document.querySelector('[role="tooltip"]')).toBeNull()
  })
})

import BaseDialog from '../BaseDialog.vue'
it('keeps dialog close hints opt-in', async () => {
  const ordinary = mount(BaseDialog, { attachTo: document.body, props: { show: true, title: 'Ordinary' } })
  const ordinaryClose = document.querySelector('[aria-label="Close modal"]')!
  ordinaryClose.dispatchEvent(new MouseEvent('mouseenter'))
  expect(document.querySelector('[role="tooltip"]')).toBeNull()
  ordinary.unmount()
  const i18n = createI18n({ legacy: false, locale: 'zh', messages: { zh: { common: { close: '关闭' } }, en: { common: { close: 'Close' } } } })
  const studio = mount(BaseDialog, { attachTo: document.body, global: { plugins: [i18n] }, props: { show: true, title: 'Studio', closeTooltip: true } })
  const studioClose = document.querySelector('[aria-label="关闭"]')!
  studioClose.dispatchEvent(new MouseEvent('mouseenter'))
  expect(document.querySelector('[role="tooltip"]')?.textContent).toBe('关闭')
  document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
  expect(document.querySelector('[role="tooltip"]')).toBeNull()
  expect(studio.emitted('close')).toHaveLength(1)
  i18n.global.locale.value = 'en'
  await studio.vm.$nextTick()
  studioClose.dispatchEvent(new MouseEvent('mouseenter'))
  expect(document.querySelector('[role="tooltip"]')?.textContent).toBe('Close')
  studio.unmount()
})


it('keeps only the latest focused or hovered action described', async () => {
  const wrapper = mount({ directives: { tooltip: vTooltip }, template: '<div><button v-tooltip aria-label="first">1</button><button v-tooltip aria-label="second">2</button></div>' }, { attachTo: document.body })
  const [first, second] = wrapper.findAll('button')
  await first!.trigger('focus')
  await second!.trigger('mouseenter')
  expect(document.querySelectorAll('[role="tooltip"]')).toHaveLength(1)
  expect(document.querySelector('[role="tooltip"]')?.textContent).toBe('second')
  expect(first!.attributes('aria-describedby')).toBeUndefined()
  await second!.trigger('focus')
  await first!.trigger('mouseenter')
  expect(document.querySelectorAll('[role="tooltip"]')).toHaveLength(1)
  expect(document.querySelector('[role="tooltip"]')?.textContent).toBe('first')
  wrapper.unmount()
})


it('lets the pointer cross the gap, hover the bubble and leave without leaking timers', async () => {
  vi.useFakeTimers()
  const wrapper = mount({ directives: { tooltip: vTooltip }, template: '<button v-tooltip aria-label="下载">↓</button>' }, { attachTo: document.body })
  const button = wrapper.get('button')
  await button.trigger('mouseenter')
  const bubble = document.querySelector('[role="tooltip"]')!
  await button.trigger('mouseleave')
  vi.advanceTimersByTime(100)
  bubble.dispatchEvent(new MouseEvent('mouseenter'))
  vi.advanceTimersByTime(200)
  expect(document.querySelector('[role="tooltip"]')).toBe(bubble)
  bubble.dispatchEvent(new MouseEvent('mouseleave'))
  vi.advanceTimersByTime(151)
  expect(document.querySelector('[role="tooltip"]')).toBeNull()
  await button.trigger('mouseenter')
  await button.trigger('mouseleave')
  wrapper.unmount()
  expect(vi.getTimerCount()).toBe(0)
  vi.useRealTimers()
})
