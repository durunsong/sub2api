import { describe, expect, it } from 'vitest'
import { prepareImageDownload } from './download'
const bytes = (blob: Blob) => new Promise<ArrayBuffer>(resolve => { const reader = new FileReader(); reader.onload = () => resolve(reader.result as ArrayBuffer); reader.readAsArrayBuffer(blob) })
describe('local batch download', () => {
  it('keeps one image intact and packs multiple originals with a valid ZIP directory', async () => {
    const image = new Blob(['123456789'], { type: 'image/png' })
    const single = await prepareImageDownload([image])
    expect(single.blob).toBe(image)
    expect(single.name).toMatch(/\.png$/)
    const zip = await prepareImageDownload([image, new Blob(['second'], { type: 'image/jpeg' })])
    const data = await bytes(zip.blob), view = new DataView(data)
    expect(zip.name).toMatch(/\.zip$/)
    expect(view.getUint32(0, true)).toBe(0x04034b50)
    expect(view.getUint32(14, true)).toBe(0xcbf43926) // Known CRC32 test vector.
    const end = data.byteLength - 22
    expect(view.getUint32(end, true)).toBe(0x06054b50)
    expect(view.getUint16(end + 10, true)).toBe(2)
    const directory = view.getUint32(end + 16, true)
    expect(view.getUint32(directory, true)).toBe(0x02014b50)
    expect(view.getUint32(directory + 42, true)).toBe(0)
  })
  it('rejects empty and oversized selections before reading files', async () => {
    await expect(prepareImageDownload([])).rejects.toThrow()
    const huge = new Blob(['x'])
    Object.defineProperty(huge, 'size', { value: 513 * 1024 * 1024 })
    await expect(prepareImageDownload([huge, huge])).rejects.toThrow()
  })
})
