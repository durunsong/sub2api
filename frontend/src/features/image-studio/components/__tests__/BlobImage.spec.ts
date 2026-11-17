import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import BlobImage from '../BlobImage.vue'

afterEach(() => vi.unstubAllGlobals())

describe('BlobImage', () => {
  it('replaces and releases image URLs without retaining old image data', async () => {
    const create = vi.fn().mockReturnValueOnce('blob:first').mockReturnValueOnce('blob:second')
    const revoke = vi.fn()
    vi.stubGlobal('URL', { createObjectURL: create, revokeObjectURL: revoke })
    const wrapper = mount(BlobImage, { props: { blob: new Blob(['a']), alt: 'Reference' } })
    expect(wrapper.get('img').attributes('src')).toBe('blob:first')
    await wrapper.setProps({ blob: new Blob(['b']) })
    expect(wrapper.get('img').attributes('src')).toBe('blob:second')
    expect(revoke).toHaveBeenCalledWith('blob:first')
    wrapper.unmount()
    expect(revoke).toHaveBeenCalledWith('blob:second')
  })
})
