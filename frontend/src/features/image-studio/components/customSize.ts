export function normalizeDimensions(width: number, height: number, align: boolean, step: number): string {
  const normalize = (value: number) => {
    const integer = Math.max(1, Math.floor(Number(value) || 1))
    return align ? Math.ceil(integer / step) * step : integer
  }
  return `${normalize(width)}x${normalize(height)}`
}

export function isValidCustomSize(size: string, step: number): boolean {
  const match = /^(\d+)x(\d+)$/.exec(size)
  if (!match) return false
  const width = Number(match[1]), height = Number(match[2])
  const pixels = width * height
  return width > 0 && height > 0 && width <= 3840 && height <= 3840 && width % step === 0 && height % step === 0 && pixels >= 655360 && pixels <= 8294400 && Math.max(width, height) / Math.min(width, height) <= 3
}
