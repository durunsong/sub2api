export interface ImageSnapshot {
  groupID: number
  model: string
  modelID?: string
  prompt: string
  size: string
  quality: string
  count: number
  aspectRatio?: string
  resolution?: string
  modelLabel?: string
  mask?: Blob
}

export interface ImageHistory {
  id: string
  createdAt: number
  status: 'success'
  snapshot: ImageSnapshot
  images: Blob[]
  references: Blob[]
  favorite: boolean
  categoryID: string | null
  categoryIDs?: string[]
}

export interface ImageAsset {
  id: string
  name: string
  createdAt: number
  image: Blob
  categoryID: string | null
  categoryIDs?: string[]
}

export interface ImageCategory {
  id: string
  name: string
}

export interface ImageDraft {
  id: 'current'
  snapshot: ImageSnapshot
  references: Blob[]
  updatedAt: number
  layout?: 'side' | 'bottom'
}

export interface ImagePrompt {
  id: string
  title: string
  prompt: string
  createdAt: number
  updatedAt?: number
  category?: string
  tags?: string[]
  seriesPrompts?: string[]
}

export type ImageTaskStatus = 'queued' | 'running' | 'failed' | 'cancelled' | 'interrupted' | 'success'

export interface ImageTask {
  id: string
  ownerID?: string
  createdAt: number
  status: ImageTaskStatus
  snapshot: ImageSnapshot
  references: Blob[]
  error?: string
}

export interface ImageStudioBackup {
  version: 1
  history: ImageHistory[]
  assets: ImageAsset[]
  categories: ImageCategory[]
  drafts: ImageDraft[]
  prompts: ImagePrompt[]
  tasks: ImageTask[]
}
