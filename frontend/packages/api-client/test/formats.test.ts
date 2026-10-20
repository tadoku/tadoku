import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import * as api from '../src'

type Verdicts = { accepted: string[]; rejected?: string[]; readByGoFallbackOnly?: string[] }

// Go's verdict on each value. `email`: what the email feature's `validAddress` stores as a recipient and answers as
// written. `account-email`: what the API answers as an account's email for stored values Kratos v26.2.0 or the
// identity import admits; it writes the address `net/mail` reads from the stored value, or the stored value itself
// when `net/mail` can't read it, such as an imported address with a leading vertical tab. `date-time`: what
// `time.Parse(time.RFC3339, …)` reads; `readByGoFallbackOnly` lists times outside RFC 3339 that Go's layout fallback
// still reads but never writes, and the client rejects them.
const go = JSON.parse(readFileSync(new URL('formats.json', import.meta.url), 'utf8')) as Record<string, Verdicts>

const account = {
  id: '00000000-0000-4000-8000-000000000001',
  email: 'reader@example.test',
  display_name: 'Reader',
  created_at: '2026-01-01T00:00:00Z',
  authenticated_at: '2026-01-01T00:00:00Z',
}

const message = {
  id: '00000000-0000-4000-8000-000000000002',
  tenant: 'tadoku/prod',
  template: 'account.recovery',
  recipient: 'reader@example.test',
  subject: 'Recover your Tadoku account',
  body: 'Hello',
  params: {},
  status: 'held',
  sent_at: null,
  provider_message_id: null,
  error: null,
  created_at: '2026-01-01T00:00:00Z',
}

const label = (value: string) =>
  value.length > 40 ? `${JSON.stringify(value.slice(0, 12))}… (${value.length} characters)` : JSON.stringify(value)

const answer = (body: unknown) => ({
  baseUrl: 'https://api.test',
  fetch: async () =>
    new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } }),
})

const session = (fields: Record<string, string>, accountApp = 'v1') =>
  api.identitySessionGet(answer({ account_app: accountApp, account: { ...account, ...fields } }))

const recipient = (value: string) =>
  api.emailGet({ ...answer({ ...message, recipient: value }), path: { id: message.id } })

describe.each([
  ['email', (value: string) => recipient(value)],
  ['date-time', (value: string) => session({ authenticated_at: value })],
])('generated validator on %s values Go has judged', (format, send) => {
  it.each(go[format].accepted.map(value => [label(value), value]))('accepts %s', async (_, value) => {
    expect((await send(value)).error).toBeUndefined()
  })

  const rejected = [...go[format].rejected ?? [], ...go[format].readByGoFallbackOnly ?? []]
  it.each(rejected.map(value => [label(value), value]))('rejects %s', async (_, value) => {
    expect((await send(value)).error).toBeDefined()
  })
})

describe.each(['v1', 'v2'])('generated %s session validator', accountApp => {
  it.each(go['account-email'].accepted.map(value => [label(value), value]))(
    'accepts the account email the API sends: %s',
    async (_, value) => {
      expect((await session({ email: value }, accountApp)).error).toBeUndefined()
    },
  )
})
