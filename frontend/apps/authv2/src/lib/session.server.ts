import { getRequestHeader, setResponseHeader } from '@tanstack/react-start/server'
import { z } from 'zod'
import { requestUrls } from './urls.server'

const sessionSchema = z.object({
  account_app: z.enum(['v1', 'v2']),
  account: z.object({
    id: z.string().min(1),
    email: z.email(),
    display_name: z.string().min(2).max(32),
    created_at: z.iso.datetime({ offset: true }),
    authenticated_at: z.iso.datetime({ offset: true }),
  }).nullable(),
})

export async function readSession() {
  setResponseHeader('Cache-Control', 'private, no-store')
  const { apiUrl, ...urls } = requestUrls()
  const headers = new Headers({ Accept: 'application/json' })
  for (const name of ['cookie', 'x-dev-branch']) {
    const value = getRequestHeader(name)
    if (value) headers.set(name, value)
  }

  try {
    const response = await fetch(`${apiUrl}/identity/session`, {
      method: 'GET',
      headers,
      cache: 'no-store',
      redirect: 'error',
      signal: AbortSignal.timeout(10_000),
    })
    if (!response.ok) return { status: 'unavailable' as const, ...urls }
    return { status: 'ready' as const, ...urls, session: sessionSchema.parse(await response.json()) }
  } catch {
    return { status: 'unavailable' as const, ...urls }
  }
}
