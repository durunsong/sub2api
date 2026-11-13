import { afterEach, describe, expect, it, vi } from 'vitest'
import { listKeyModels } from '../keyModels'
import { buildGatewayUrl } from '../url'

describe('API key model discovery', () => {
  afterEach(() => vi.unstubAllGlobals())

  it('uses the gateway origin and selected API key, with cancellation and no cache', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({
      data: [{ id: 'allowed-model' }, { id: 'another-model' }, { id: 'allowed-model' }]
    })))
    vi.stubGlobal('fetch', fetchMock)
    const signal = new AbortController().signal

    expect(await listKeyModels('sk-selected', signal)).toEqual(['allowed-model', 'another-model'])
    expect(fetchMock).toHaveBeenCalledWith(buildGatewayUrl('/v1/models'), {
      method: 'GET',
      headers: { Accept: 'application/json', Authorization: 'Bearer sk-selected' },
      cache: 'no-store',
      credentials: 'omit',
      signal
    })
  })

  it('preserves custom aliases returned by the group-aware gateway catalog', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response('{"data":[{"id":"my-model"}]}'))
    vi.stubGlobal('fetch', fetchMock)
    expect(await listKeyModels('sk-selected')).toEqual(['my-model'])
    expect(fetchMock).toHaveBeenCalledWith(buildGatewayUrl('/v1/models'), expect.any(Object))
  })

  it('preserves an empty authorized list without inventing default models', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('{"data":[]}')))
    expect(await listKeyModels('sk-selected')).toEqual([])
  })

  it.each([401, 403, 500])('rejects HTTP %s without exposing a response body or changing the login session', async (status) => {
    localStorage.setItem('auth_token', 'user-session')
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('private upstream details', { status })))
    await expect(listKeyModels('sk-selected')).rejects.toThrow(`Models request failed with status ${status}`)
    expect(localStorage.getItem('auth_token')).toBe('user-session')
    localStorage.removeItem('auth_token')
  })

  it.each([null, {}, { models: [] }, { data: null }, { data: [{ id: '' }] }, { data: [{ id: 3 }] }])(
    'rejects a malformed catalog: %j', async (payload) => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify(payload))))
      await expect(listKeyModels('sk-selected')).rejects.toThrow('Invalid models response')
    }
  )
})
