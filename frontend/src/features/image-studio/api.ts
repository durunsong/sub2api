import { apiClient } from '@/api/client'
import type { AxiosRequestConfig } from 'axios'

export interface StudioModel {
  id: string
  group_id: number
  group_name: string
  model: string
  label: string
  protocol: 'images' | 'gemini'
  sizes: string[]
  qualities: string[]
  max_images: number
  max_references: number
  supports_mask: boolean
  supports_custom_size?: boolean
  size_multiple?: number
  pricing: {
    currency: 'USD'
    mode: 'image' | 'token'
    rate_multiplier: number
    image_price_1k: number | null
    image_price_2k: number | null
    image_price_4k: number | null
  }
}
export interface StudioCatalog { enabled: boolean; models: StudioModel[] }
export interface StudioModelSetting {
  group_id: number
  model: string
  enabled: boolean
  label: string
  protocol?: 'images' | 'gemini'
}
export interface StudioConfig { enabled: boolean; models: StudioModelSetting[] }
export interface WorkflowModel {
  id: string
  group_id: number
  group_name: string
  model: string
  label: string
  supports_references: boolean
}
export interface WorkflowCatalog { enabled: boolean; models: WorkflowModel[] }
export interface WorkflowDraft { title: string; prompt: string; category: string; tags: string[]; seriesPrompts: string[] }
export interface WorkflowRequest {
  model_id: string
  mode: 'workflow' | 'series'
  prompt: string
  references?: string[]
  count?: number
}
export type WorkflowResult = { draft: WorkflowDraft } | { prompts: string[] }
export interface GenerateRequest {
  model_id: string
  prompt: string
  size?: string
  quality?: string
  n: number
  aspect_ratio?: string
  resolution?: string
  references?: string[]
  mask?: string
}

const mimeTypes = ['image/png', 'image/jpeg', 'image/webp']
export function decodeGeneratedImages(response: unknown): Blob[] {
  const data = response as { images?: { b64_json?: unknown; mime_type?: unknown }[] } | null
  if (!Array.isArray(data?.images) || !data.images.length || data.images.length > 16) {
    throw new Error('imageStudio.errors.noImages')
  }
  return data.images.map(image => {
    if (!mimeTypes.includes(String(image.mime_type)) || typeof image.b64_json !== 'string' || !image.b64_json.length || image.b64_json.length > 45 * 1024 * 1024) {
      throw new Error('imageStudio.errors.invalidImage')
    }
    try {
      const bytes = Uint8Array.from(atob(image.b64_json), c => c.charCodeAt(0))
      if (!bytes.length) throw new Error('empty image')
      return new Blob([bytes], { type: image.mime_type as string })
    } catch { throw new Error('imageStudio.errors.invalidImage') }
  })
}

export async function imageToDataURL(image: Blob): Promise<string> {
  if (!mimeTypes.includes(image.type) || !image.size || image.size > 32 * 1024 * 1024) throw new Error('imageStudio.errors.invalidImage')
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result))
    reader.onerror = () => reject(new Error('imageStudio.errors.readImage'))
    reader.readAsDataURL(image)
  })
}

export async function generateImages(request: GenerateRequest, signal: AbortSignal): Promise<{ images: Blob[]; text?: string }> {
  // A billable request must never be replayed by the shared 401 refresh interceptor.
  const config: AxiosRequestConfig & { _retry: boolean } = { signal, timeout: 600000, _retry: true }
  const { data } = await apiClient.post('/image-studio/generate', request, config)
  return { images: decodeGeneratedImages(data), text: typeof data?.text === 'string' ? data.text : undefined }
}

export async function generateWorkflow(request: WorkflowRequest, signal: AbortSignal): Promise<WorkflowResult> {
  const config: AxiosRequestConfig & { _retry: boolean } = { signal, timeout: 180000, _retry: true }
  const data = (await apiClient.post('/image-studio/workflow', request, config)).data
  const text = (value: unknown, max: number, empty = false) => typeof value === 'string' && value.length <= max && (empty || !!value.trim())
  const texts = (value: unknown, limit: number, length: number): value is string[] => Array.isArray(value) && value.length <= limit && value.every(item => text(item, length))
  if (request.mode === 'series' && texts(data?.prompts, 20, 10000) && data.prompts.length === request.count) return { prompts: data.prompts }
  const draft = data?.draft
  if (request.mode === 'workflow' && draft && text(draft.title, 120) && text(draft.prompt, 10000) && text(draft.category, 120, true) && texts(draft.tags, 20, 80) && texts(draft.seriesPrompts, 20, 10000)) {
    return { draft: { title: draft.title, prompt: draft.prompt, category: draft.category, tags: draft.tags, seriesPrompts: draft.seriesPrompts } }
  }
  throw new Error('imageStudio.aiFailed')
}

export const imageStudioAPI = {
  workflowModels: async (signal?: AbortSignal): Promise<WorkflowCatalog> => (await apiClient.get('/image-studio/workflow-models', { signal })).data,
  adminWorkflowModels: async (): Promise<WorkflowCatalog> => (await apiClient.get('/admin/image-studio/workflow-models')).data,
  catalog: async (signal?: AbortSignal): Promise<StudioCatalog> => (await apiClient.get('/image-studio/catalog', { signal })).data,
  adminCatalog: async (): Promise<StudioCatalog> => (await apiClient.get('/admin/image-studio/catalog')).data,
  config: async (): Promise<StudioConfig> => (await apiClient.get('/admin/image-studio/config')).data,
  saveConfig: async (config: StudioConfig): Promise<StudioConfig> => (await apiClient.put('/admin/image-studio/config', config)).data,
}
