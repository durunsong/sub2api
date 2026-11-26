import { afterEach, expect, it, vi } from 'vitest'
import { readVideoMetadata, videoFilename } from './media'

it('uses the actual container extension for downloaded assets', () => {
  expect(videoFilename('asset', new Blob([], { type: 'video/webm' }))).toBe('video-asset.webm')
  expect(videoFilename('asset', new Blob([], { type: 'video/mp4' }))).toBe('video-asset.mp4')
})

afterEach(() => vi.restoreAllMocks())

function fixture(width: number, duration: number) {
  const video = document.createElement('video')
  Object.defineProperties(video, { videoWidth: { value: width }, videoHeight: { value: 180 }, duration: { value: duration } })
  vi.spyOn(document, 'createElement').mockReturnValue(video)
  vi.spyOn(video, 'load').mockImplementation(() => {})
  URL.createObjectURL = vi.fn(() => 'blob:test')
  URL.revokeObjectURL = vi.fn()
  return video
}

it('uses decoded dimensions and duration and releases its local URL', async () => {
  const video = fixture(320, 1.5)
  const result = readVideoMetadata(new Blob(['video']))
  video.dispatchEvent(new Event('loadedmetadata'))
  await expect(result).resolves.toEqual({ width: 320, height: 180, videoDurationMs: 1500 })
  expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:test')
})

it('rejects undecodable output rather than persisting it as a successful video', async () => {
  const video = fixture(0, NaN)
  const result = readVideoMetadata(new Blob(['invalid']))
  video.dispatchEvent(new Event('loadedmetadata'))
  await expect(result).rejects.toThrow('videoStudio.invalidVideo')
  expect(URL.revokeObjectURL).toHaveBeenCalled()
})

it('releases the URL when account switching aborts inspection', async () => {
  fixture(320, 1)
  const abort = new AbortController()
  const result = readVideoMetadata(new Blob(['video']), abort.signal)
  abort.abort()
  await expect(result).rejects.toMatchObject({ name: 'AbortError' })
  expect(URL.revokeObjectURL).toHaveBeenCalled()
})
