import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'
import GenerationPending from '../GenerationPending.vue'

const props = {
  status: 'queued' as const,
  title: 'Queued',
  subtitle: 'Waiting for an available slot',
  elapsed: '0:12'
}

afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})

describe('GenerationPending', () => {
  it('announces the actual state without repeatedly announcing elapsed time or inventing progress', async () => {
    const wrapper = mount(GenerationPending, { props })
    const status = wrapper.get('[role="status"]')
    expect(status.text()).toContain('Queued')
    expect(status.text()).toContain('Waiting for an available slot')
    expect(status.text()).not.toContain('0:12')

    await wrapper.setProps({ status: 'running', title: 'Generating', subtitle: 'You can keep creating', elapsed: '0:13' })
    expect(status.text()).toContain('Generating')
    expect(status.text()).not.toContain('Queued')
    expect(status.text()).not.toContain('0:13')
    expect(wrapper.text()).toContain('0:13')
    expect(wrapper.find('[role="progressbar"], progress, [aria-valuenow]').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('%')
    expect(wrapper.find('svg').attributes('aria-hidden')).toBe('true')
    wrapper.unmount()
  })

  it('pauses motion when the document is hidden and resumes on return', async () => {
    const hidden = vi.spyOn(document, 'hidden', 'get').mockReturnValue(false)
    const wrapper = mount(GenerationPending, { props })
    expect(wrapper.attributes('data-motion')).toBe('running')
    hidden.mockReturnValue(true)
    document.dispatchEvent(new Event('visibilitychange'))
    await nextTick()
    expect(wrapper.attributes('data-motion')).toBe('paused')
    hidden.mockReturnValue(false)
    document.dispatchEvent(new Event('visibilitychange'))
    await nextTick()
    expect(wrapper.attributes('data-motion')).toBe('running')
    wrapper.unmount()
  })

  it('pauses an offscreen card without hiding its status', async () => {
    let reportVisibility: IntersectionObserverCallback = () => {}
    vi.stubGlobal('IntersectionObserver', class {
      constructor(callback: IntersectionObserverCallback) { reportVisibility = callback }
      observe() {}
      disconnect() {}
    })
    const wrapper = mount(GenerationPending, { props })
    reportVisibility([{ isIntersecting: false } as IntersectionObserverEntry], {} as IntersectionObserver)
    await nextTick()
    expect(wrapper.attributes('data-motion')).toBe('paused')
    expect(wrapper.get('[role="status"]').text()).toContain('Queued')
    reportVisibility([{ isIntersecting: true } as IntersectionObserverEntry], {} as IntersectionObserver)
    await nextTick()
    expect(wrapper.attributes('data-motion')).toBe('running')
    wrapper.unmount()
  })
})
