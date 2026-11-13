import { buildGatewayUrl } from './url'

// Model discovery uses the selected API key, not the dashboard JWT client.
export async function listKeyModels(
  apiKey: string,
  signal?: AbortSignal
): Promise<string[]> {
  // The shared catalog resolves account mappings and group authorization for every platform.
  const response = await fetch(buildGatewayUrl('/v1/models'), {
    method: 'GET',
    headers: { Accept: 'application/json', Authorization: `Bearer ${apiKey}` },
    cache: 'no-store',
    credentials: 'omit',
    signal
  })
  if (!response.ok) throw new Error(`Models request failed with status ${response.status}`)

  const payload: unknown = await response.json()
  if (!payload || typeof payload !== 'object' || !('data' in payload) ||
    !Array.isArray(payload.data) || !payload.data.every((model) =>
      model && typeof model.id === 'string' && model.id.trim().length > 0
    )) {
    throw new Error('Invalid models response')
  }
  return [...new Set<string>(payload.data.map((model) => model.id))]
}
