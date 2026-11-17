import { describe, expect, it, vi, beforeEach } from 'vitest'
import { decodeGeneratedImages, generateImages, generateWorkflow, imageToDataURL } from './api'

const { post } = vi.hoisted(() => ({ post: vi.fn() }))
vi.mock('@/api/client', () => ({ default: { post }, apiClient: { post } }))
const png = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII='

describe('image studio transport', () => {
  beforeEach(() => { post.mockReset() })
  it('rejects invalid AI results instead of offering incomplete templates or series', async () => {
    for (const data of [{ draft: null }, { draft: { title: 'title', prompt: '', category: '', tags: [], seriesPrompts: [] } }, { prompts: ['only one'] }, { prompts: ['one', 2] }]) {
      post.mockResolvedValueOnce({ data })
      await expect(generateWorkflow({ model_id: 'text', mode: 'draft' in data ? 'workflow' : 'series', prompt: 'brief', count: 2 }, new AbortController().signal)).rejects.toThrow('imageStudio.aiFailed')
    }
  })
  it('uses a cancellable non-replayed request for AI series generation', async () => {
    post.mockResolvedValueOnce({ data: { prompts: ['one', 'two'] } })
    const signal = new AbortController().signal
    expect(await generateWorkflow({ model_id: 'text', mode: 'series', prompt: 'brief', count: 2 }, signal)).toEqual({ prompts: ['one', 'two'] })
    expect(post).toHaveBeenCalledWith('/image-studio/workflow', { model_id: 'text', mode: 'series', prompt: 'brief', count: 2 }, expect.objectContaining({ signal, _retry: true }))
  })
  it('converts returned base64 images into local blobs without fetching URLs', () => {
    const images = decodeGeneratedImages({ images: [{ b64_json: png, mime_type: 'image/png' }] })
    expect(images).toHaveLength(1)
    expect(images[0]?.type).toBe('image/png')
    expect(images[0]?.size).toBeGreaterThan(0)
  })
  it('rejects missing output and remote-only or unsupported images', () => {
    expect(() => decodeGeneratedImages({ images: [] })).toThrow()
    expect(() => decodeGeneratedImages({ images: [{ url: 'https://example.org/image.png' }] })).toThrow()
    expect(() => decodeGeneratedImages({ images: [{ b64_json: 'xxx', mime_type: 'image/svg+xml' }] })).toThrow()
  })
  it('sends only the selected catalog identity and native parameters without API credentials', async () => {
    post.mockResolvedValue({ data: { images: [{ b64_json: png, mime_type: 'image/png' }] } })
    const controller = new AbortController()
    const result = await generateImages({ model_id: '7:gpt-image-2', prompt: 'sea', n: 1, size: '1024x1024', quality: 'auto', references: [] }, controller.signal)
    expect(result.images[0]?.type).toBe('image/png')
    expect(post).toHaveBeenCalledWith('/image-studio/generate', expect.objectContaining({ model_id: '7:gpt-image-2' }), expect.objectContaining({ signal: controller.signal, timeout: 600000 }))
  })
  it('does not retry a failed or cancelled billable request', async () => {
    post.mockRejectedValue(new Error('upstream failed'))
    await expect(generateImages({ model_id: '7:gpt-image-2', prompt: 'sea', n: 1 }, new AbortController().signal)).rejects.toThrow('upstream failed')
    expect(post).toHaveBeenCalledTimes(1)
  })
  it('encodes local references and rejects non-image data', async () => {
    expect(await imageToDataURL(new Blob(['png'], { type: 'image/png' }))).toBe('data:image/png;base64,cG5n')
    await expect(imageToDataURL(new Blob(['<svg/>'], { type: 'image/svg+xml' }))).rejects.toThrow()
  })
})
