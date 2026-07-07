import { apiClient } from '../client'
import type {
  BasePaginationResponse,
  CreateIPBanRequest,
  FetchOptions,
  IPBan,
  UpdateIPBanRequest
} from '@/types'

export async function list(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    status?: string
    search?: string
    sort_by?: string
    sort_order?: 'asc' | 'desc'
  },
  options?: FetchOptions
): Promise<BasePaginationResponse<IPBan>> {
  const { data } = await apiClient.get<BasePaginationResponse<IPBan>>('/admin/ip-bans', {
    params: { page, page_size: pageSize, ...filters },
    signal: options?.signal
  })
  return data
}

export async function getById(id: number): Promise<IPBan> {
  const { data } = await apiClient.get<IPBan>(`/admin/ip-bans/${id}`)
  return data
}

export async function create(request: CreateIPBanRequest): Promise<IPBan> {
  const { data } = await apiClient.post<IPBan>('/admin/ip-bans', request)
  return data
}

export async function update(id: number, request: UpdateIPBanRequest): Promise<IPBan> {
  const { data } = await apiClient.put<IPBan>(`/admin/ip-bans/${id}`, request)
  return data
}

export async function deleteRule(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/admin/ip-bans/${id}`)
  return data
}

const ipBansAPI = {
  list,
  getById,
  create,
  update,
  delete: deleteRule
}

export default ipBansAPI
