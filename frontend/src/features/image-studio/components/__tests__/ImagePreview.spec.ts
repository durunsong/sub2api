import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ImagePreview from '../ImagePreview.vue'

vi.mock('vue-i18n', async original => ({ ...(await original<typeof import('vue-i18n')>()), useI18n: () => ({ t: (key: string) => key }) }))
afterEach(() => { vi.restoreAllMocks(); vi.unstubAllGlobals(); document.body.innerHTML = '' })
enableAutoUnmount(afterEach)
let created: ReturnType<typeof vi.fn>, revoked: ReturnType<typeof vi.fn>
beforeEach(() => {
  created = vi.fn(() => `blob:preview-${Math.random()}`)
  revoked = vi.fn()
  vi.stubGlobal('URL', { createObjectURL: created, revokeObjectURL: revoked })
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(function (this: HTMLElement) {
    return { width: 800, height: 400, left: 0, top: 0, right: 800, bottom: 400, x: 0, y: 0, toJSON() {} } as DOMRect
  })
  HTMLElement.prototype.setPointerCapture = vi.fn()
})
const blobs = () => [new Blob(['first'], { type: 'image/png' }), new Blob(['second'], { type: 'image/jpeg' })]
const root = () => document.querySelector<HTMLElement>('[data-image-preview]')!
const stage = () => root().querySelector<HTMLElement>('[data-preview-stage]')!
const button = (label: string) => [...root().querySelectorAll<HTMLButtonElement>('button')].find(item => item.getAttribute('aria-label') === label || item.textContent?.trim() === label)!
async function loaded(width = 1600, height = 800) {
  const image = stage().querySelector('img')!
  Object.defineProperty(image, 'naturalWidth', { value: width, configurable: true })
  Object.defineProperty(image, 'naturalHeight', { value: height, configurable: true })
  image.dispatchEvent(new Event('load'))
  await flushPromises()
}
function pointer(type: string, id: number, x: number, y: number) {
  const event = new Event(type, { bubbles: true, cancelable: true })
  Object.assign(event, { pointerId: id, pointerType: 'touch', button: 0, clientX: x, clientY: y })
  stage().dispatchEvent(event)
}

describe('ImagePreview', () => {
  it('fits uncropped image dimensions, switches through the gallery and resets view state', async () => {
    const images = blobs()
    const wrapper = mount(ImagePreview, { attachTo: document.body, global: { stubs: { transition: true } }, props: { show: true, images, captions: ['First image', 'Second image'] } })
    await flushPromises(); await loaded()
    expect(stage().dataset.zoom).toBe('0.5')
    expect(stage().querySelector('img')!.style.width).toBe('1600px')
    expect(root().textContent).toContain('1600 × 800')
    button('imageStudio.actualSize').click(); await flushPromises()
    expect(stage().dataset.zoom).toBe('1')
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowRight', bubbles: true }))
    await flushPromises(); await loaded(500, 1000)
    expect(stage().dataset.zoom).toBe('0.4')
    expect(root().textContent).toContain('Second image')
    expect(button('imageStudio.nextImage').disabled).toBe(true)
    button('imageStudio.download').click()
    expect(wrapper.emitted('download')).toEqual([[images[1]]])
    await wrapper.setProps({ show: false }); await flushPromises()
    expect(document.querySelector('[data-image-preview]')).toBeNull()
    expect(revoked).toHaveBeenCalledTimes(created.mock.calls.length)
  })

  it('keeps zoom and drag bounded and restores fit after pointer cancellation', async () => {
    mount(ImagePreview, { attachTo: document.body, global: { stubs: { transition: true } }, props: { show: true, images: blobs() } })
    await flushPromises(); await loaded()
    pointer('pointerdown', 1, 300, 200)
    pointer('pointerdown', 2, 500, 200)
    pointer('pointermove', 2, 700, 200)
    await flushPromises()
    expect(stage().dataset.zoom).toBe('1')
    expect(stage().dataset.panX).toBe('100')
    pointer('pointercancel', 2, 700, 200)
    pointer('pointermove', 1, 3000, 2000)
    await flushPromises()
    expect(stage().dataset.panX).toBe('400')
    expect(stage().dataset.panY).toBe('200')
    pointer('pointerup', 1, 3000, 2000)
    button('imageStudio.fitImage').click(); await flushPromises()
    expect(stage().dataset.zoom).toBe('0.5')
    expect(stage().dataset.panX).toBe('0')
    expect(stage().dataset.panY).toBe('0')
    stage().dispatchEvent(new WheelEvent('wheel', { deltaY: -10000, clientX: 400, clientY: 200, cancelable: true }))
    await flushPromises()
    expect(stage().dataset.zoom).toBe('8')
    expect(button('imageStudio.zoomIn').disabled).toBe(true)
  })

  it('retries a local decode error and releases the failed object URL', async () => {
    mount(ImagePreview, { attachTo: document.body, global: { stubs: { transition: true } }, props: { show: true, images: blobs() } })
    await flushPromises()
    const original = stage().querySelector('img')!.src
    stage().querySelector('img')!.dispatchEvent(new Event('error'))
    await flushPromises()
    expect(root().textContent).toContain('imageStudio.previewFailed')
    expect(button('imageStudio.zoomIn').disabled).toBe(true)
    button('imageStudio.retryPreview').click(); await flushPromises()
    expect(revoked).toHaveBeenCalledWith(original)
    expect(stage().querySelector('img')!.src).not.toBe(original)
    await loaded()
    expect(button('imageStudio.zoomIn').disabled).toBe(false)
  })

  it('skips hidden thumbnails when trapping focus in a short viewport', async () => {
    mount(ImagePreview, { attachTo: document.body, global: { stubs: { transition: true } }, props: { show: true, images: blobs() } })
    await flushPromises(); await loaded()
    root().querySelector<HTMLElement>('.preview-thumbnails')!.style.display = 'none'
    const close = button('imageStudio.closePreview')
    close.focus()
    close.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true, cancelable: true }))
    expect(document.activeElement).toBe(stage())
  })

  it('contains keyboard focus and returns it to the opener after Escape', async () => {
    const opener = document.createElement('button'); document.body.appendChild(opener); opener.focus()
    const wrapper = mount(ImagePreview, { attachTo: document.body, global: { stubs: { transition: true } }, props: { show: true, images: blobs() } })
    await flushPromises(); await loaded()
    const close = button('imageStudio.closePreview')
    expect(document.activeElement).toBe(close)
    close.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true, cancelable: true }))
    expect(document.activeElement).toBe(root().querySelector('.preview-thumbnails')!.lastElementChild)
    document.activeElement!.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true, cancelable: true }))
    expect(document.activeElement).toBe(close)
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    expect(wrapper.emitted('close')).toHaveLength(1)
    await wrapper.setProps({ show: false }); await flushPromises()
    expect(document.activeElement).toBe(opener)
    expect(document.body.classList.contains('modal-open')).toBe(false)
  })
})
