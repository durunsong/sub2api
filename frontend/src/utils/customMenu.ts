import type { CustomMenuItem } from '@/types'

export function resolveCustomMenuLabel(item: Pick<CustomMenuItem, 'label' | 'visibility'>, tutorialLabel: string): string {
  return item.visibility === 'user' && item.label.trim() === '文档'
    ? tutorialLabel
    : item.label
}
