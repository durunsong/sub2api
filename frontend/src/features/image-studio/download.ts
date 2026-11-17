// ZIP store mode keeps already-compressed image bytes intact without a dependency.
const crcTable = Uint32Array.from({ length: 256 }, (_, value) => {
  for (let i = 0; i < 8; i++) value = value & 1 ? 0xedb88320 ^ (value >>> 1) : value >>> 1
  return value >>> 0
})
const extension = (image: Blob) => image.type === 'image/jpeg' ? 'jpg' : image.type === 'image/webp' ? 'webp' : 'png'
function readBytes(blob: Blob): Promise<Uint8Array> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(new Uint8Array(reader.result as ArrayBuffer))
    reader.onerror = () => reject(new Error('imageStudio.downloadFailed'))
    reader.readAsArrayBuffer(blob)
  })
}
export async function prepareImageDownload(images: Blob[]): Promise<{ blob: Blob; name: string }> {
  if (!images.length || images.length > 65535 || images.reduce((sum, image) => sum + image.size, 0) > 512 * 1024 * 1024) throw new Error('imageStudio.downloadFailed')
  const stamp = Date.now()
  if (images.length === 1) return { blob: images[0], name: `image-${stamp}.${extension(images[0])}` }
  const files: BlobPart[] = [], directory: BlobPart[] = []
  let offset = 0, directorySize = 0
  for (let index = 0; index < images.length; index++) {
    const image = images[index], bytes = await readBytes(image)
    let crc = 0xffffffff
    for (const byte of bytes) crc = crcTable[(crc ^ byte) & 255] ^ (crc >>> 8)
    crc = (crc ^ 0xffffffff) >>> 0
    const name = new TextEncoder().encode(`image-${String(index + 1).padStart(2, '0')}.${extension(image)}`)
    const local = new Uint8Array(30 + name.length), header = new DataView(local.buffer)
    header.setUint32(0, 0x04034b50, true); header.setUint16(4, 20, true)
    header.setUint16(12, 33, true) // 1980-01-01, valid DOS date.
    header.setUint32(14, crc, true); header.setUint32(18, image.size, true); header.setUint32(22, image.size, true)
    header.setUint16(26, name.length, true); local.set(name, 30)
    files.push(local, image)
    const central = new Uint8Array(46 + name.length), entry = new DataView(central.buffer)
    entry.setUint32(0, 0x02014b50, true); entry.setUint16(4, 20, true); entry.setUint16(6, 20, true)
    entry.setUint16(14, 33, true); entry.setUint32(16, crc, true)
    entry.setUint32(20, image.size, true); entry.setUint32(24, image.size, true); entry.setUint16(28, name.length, true)
    entry.setUint32(42, offset, true); central.set(name, 46)
    directory.push(central); directorySize += central.length; offset += local.length + image.size
  }
  const end = new Uint8Array(22), footer = new DataView(end.buffer)
  footer.setUint32(0, 0x06054b50, true); footer.setUint16(8, images.length, true); footer.setUint16(10, images.length, true)
  footer.setUint32(12, directorySize, true); footer.setUint32(16, offset, true)
  return { blob: new Blob([...files, ...directory, end], { type: 'application/zip' }), name: `images-${stamp}.zip` }
}
