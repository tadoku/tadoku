import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import * as api from '../src'

type Verdicts = { accepted: string[]; rejected: string[]; readByGoFallbackOnly?: string[] }

// Go's verdict on each value: what `identity.ParseEmail` stores unchanged up to letter case, and what
// `time.Parse(time.RFC3339, …)` reads. `readByGoFallbackOnly` lists times outside RFC 3339 that Go's layout fallback
// still reads but never writes; the client rejects them.
const go = JSON.parse(readFileSync(new URL('formats.json', import.meta.url), 'utf8')) as Record<string, Verdicts>

const account = {
  id: '00000000-0000-4000-8000-000000000001',
  email: 'reader@example.test',
  display_name: 'Reader',
  created_at: '2026-01-01T00:00:00Z',
  authenticated_at: '2026-01-01T00:00:00Z',
}

const label = (value: string) =>
  value.length > 40 ? `${JSON.stringify(value.slice(0, 12))}… (${value.length} characters)` : JSON.stringify(value)

async function session(fields: Record<string, string>) {
  const body = JSON.stringify({ account_app: 'v1', account: { ...account, ...fields } })
  return api.identitySessionGet({
    baseUrl: 'https://api.test',
    fetch: async () => new Response(body, { status: 200, headers: { 'Content-Type': 'application/json' } }),
  })
}

describe.each([
  ['email', 'email'],
  ['date-time', 'authenticated_at'],
])('generated session validator on %s values Go has judged', (format, field) => {
  it.each(go[format].accepted.map(value => [label(value), value]))('accepts %s', async (_, value) => {
    expect((await session({ [field]: value })).error).toBeUndefined()
  })

  const rejected = [...go[format].rejected, ...go[format].readByGoFallbackOnly ?? []]
  it.each(rejected.map(value => [label(value), value]))('rejects %s', async (_, value) => {
    expect((await session({ [field]: value })).error).toBeDefined()
  })
})
