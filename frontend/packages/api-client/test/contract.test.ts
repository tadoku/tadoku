import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { load } from 'js-yaml'
import { describe, expect, it } from 'vitest'
import * as api from '../src'

type Response = { content?: Record<string, { schema?: { $ref?: string } }>; $ref?: string }
type Operation = { operationId: string; 'x-tadoku-exposure': string; responses: Record<string, Response> }
type Spec = {
  paths: Record<string, Record<string, Operation>>
  components: { responses: Record<string, Response> }
}
type Validator = { safeParse: (value: unknown) => { success: boolean } }
type Call = (options: Record<string, unknown>) => Promise<{ data?: unknown; error?: unknown; response?: globalThis.Response }>

const repository = new URL('../../../../', import.meta.url).pathname
const spec = load(readFileSync(join(repository, 'services/tadoku-api/spec/openapi.yaml'), 'utf8')) as Spec
const goldens = join(repository, 'services/tadoku-api/e2e/testdata')

const routes = Object.entries(spec.paths)
  .flatMap(([template, item]) => Object.entries(item).map(([method, operation]) => ({
    method: method.toUpperCase(),
    names: [...template.matchAll(/\{([^}]+)\}/g)].map(match => match[1]),
    pattern: new RegExp(`^${template.replace(/\{[^}]+\}/g, '([^/]+)')}$`),
    operation,
  })))
  // Literal segments win over parameters, as in the API router.
  .sort((a, b) => a.names.length - b.names.length)

const exportsByLowerName = new Map(Object.entries(api).map(([name, value]) => [name.toLowerCase(), value]))

function call(operationId: string) {
  return exportsByLowerName.get(operationId.toLowerCase()) as Call | undefined
}

function respond(status: number, headers: Record<string, string>, body: string | null) {
  return async () => new Response(body, { status, headers })
}

function exchanges() {
  return readdirSync(goldens, { recursive: true, encoding: 'utf8' })
    .filter(path => path.endsWith('golden.http'))
    .sort()
    .map(path => {
      const text = readFileSync(join(goldens, path), 'utf8')
      const [, method, target] = /^>>> (\S+) (\S+)\n/.exec(text) ?? []
      const head = text.slice(text.indexOf('\n') + 1)
      const split = head.indexOf('\n\n')
      const [statusLine, ...headerLines] = (split < 0 ? head : head.slice(0, split)).split('\n')
      const headers = Object.fromEntries(headerLines.filter(Boolean).map(line => {
        const colon = line.indexOf(':')
        return [line.slice(0, colon), line.slice(colon + 1).trim()]
      }))
      const body = split < 0 ? '' : head.slice(split + 2).replace(/\n$/, '')
      return { path, method, target, status: Number(statusLine.split(' ')[1]), headers, body }
    })
}

describe('generated client against recorded Tadoku API responses', () => {
  const cases = exchanges()

  it('reads every golden HTTP exchange', () => {
    expect(cases.length).toBeGreaterThan(1000)
    expect(cases.filter(exchange => !exchange.method || !exchange.status)).toEqual([])
  })

  it('accepts every documented response the API actually returns', async () => {
    const counts = {
      success: 0, error: 0, emptyError: 0, undocumented: 0, redirect: 0, download: 0, callback: 0, unrouted: 0,
    }
    const failures: string[] = []

    for (const exchange of cases) {
      const pathname = exchange.target.split('?')[0]
      const route = routes.find(candidate => candidate.method === exchange.method && candidate.pattern.test(pathname))
      if (!route) {
        counts.unrouted++
        continue
      }
      const { operation } = route
      if (operation['x-tadoku-exposure'] === 'callback') {
        counts.callback++
        continue
      }
      const documented = operation.responses[String(exchange.status)]
      const resolved = documented?.$ref
        ? spec.components.responses[documented.$ref.split('/').pop()!]
        : documented
      const types = Object.keys(resolved?.content ?? {})
      if (types.some(type => type !== 'application/json')) {
        counts.download++
        continue
      }
      if (exchange.status >= 300 && exchange.status < 400) {
        counts.redirect++
        continue
      }

      const fn = call(operation.operationId)
      if (!fn) {
        failures.push(`${operation.operationId}: no generated operation`)
        continue
      }
      const match = route.pattern.exec(pathname)!
      const path = Object.fromEntries(route.names.map((name, index) => [name, decodeURIComponent(match[index + 1])]))
      const empty = exchange.body === '' || [204, 205].includes(exchange.status)
      const result = await fn({
        baseUrl: 'https://api.test',
        path,
        fetch: respond(exchange.status, exchange.headers, empty ? null : exchange.body),
      })

      const label = `${exchange.path} ${operation.operationId} ${exchange.status}`
      if (!resolved) {
        counts.undocumented++
        if (result.error === undefined) failures.push(`${label}: undocumented status accepted`)
      } else if (exchange.status < 300) {
        counts.success++
        if (result.error !== undefined) failures.push(`${label}: ${String(result.error).slice(0, 300)}`)
        // The same answer under a success status the contract never documents, and with its body added or removed.
        const replay = (status: number, body: string | null) => fn({
          baseUrl: 'https://api.test',
          path,
          fetch: respond(status, empty ? {} : { 'Content-Type': 'application/json' }, body),
        })
        if ((await replay(203, empty ? null : exchange.body)).error === undefined) {
          failures.push(`${label}: undocumented 203 accepted`)
        }
        if (exchange.status !== 204 && (await replay(exchange.status, empty ? '{}' : null)).error === undefined) {
          failures.push(`${label}: ${empty ? 'body on a bodyless success' : 'empty body'} accepted`)
        }
      } else if (exchange.body === '') {
        // Some documented JSON errors arrive without a body; callers fall back to the status.
        counts.emptyError++
        if (result.error === undefined) failures.push(`${label}: error status accepted as success`)
      } else {
        counts.error++
        const reference = resolved.content?.['application/json']?.schema?.$ref ?? ''
        const schema = exportsByLowerName.get(`z${reference.split('/').pop()}`.toLowerCase()) as Validator | undefined
        if (result.error === undefined) failures.push(`${label}: error status accepted as success`)
        else if (reference && !schema) failures.push(`${label}: no generated validator for ${reference}`)
        else if (schema && !schema.safeParse(result.error).success) {
          failures.push(`${label}: error body rejected by its generated validator`)
        }
      }
    }

    expect(failures).toEqual([])
    expect(Object.values(counts).reduce((sum, count) => sum + count)).toBe(cases.length)
    expect(counts.success).toBeGreaterThan(600)
    expect(counts.error).toBeGreaterThan(0)
  })
})

describe('generated client on malformed answers', () => {
  const session = (status: number, headers: Record<string, string>, body: string | null) =>
    api.identitySessionGet({ baseUrl: 'https://api.test', fetch: respond(status, headers, body) })
  const json = { 'Content-Type': 'application/json' }

  it.each([
    ['an HTML page', 200, { 'Content-Type': 'text/html' }, '<html>Gateway</html>'],
    ['an empty body', 200, { 'Content-Length': '0' }, null],
    ['no content', 204, {}, null],
    ['a missing account', 200, json, '{}'],
    ['a missing recent_login', 200, json, '{"account":null}'],
    ['an account without an ID', 200, json, '{"account":{"email":"a@b.test","display_name":"Ab",'
      + '"created_at":"2026-01-01T00:00:00Z","authenticated_at":"2026-01-01T00:00:00Z"},"recent_login":true}'],
  ])('rejects %s', async (_, status, headers, body) => {
    const result = await session(status, headers, body)
    expect(result.data).toBeUndefined()
    expect(result.error).toBeDefined()
  })

  it('accepts a signed-out session', async () => {
    const result = await session(200, json, '{"account":null,"recent_login":false}')
    expect(result.error).toBeUndefined()
    expect(result.data).toEqual({ account: null, recent_login: false })
  })

  it('sends plain-object headers with a JSON write', async () => {
    let sent: Request | undefined
    await api.identityLoginCreate({
      baseUrl: 'https://api.test',
      headers: { Origin: 'https://account.tadoku.app', cookie: 'tadoku_session=a', 'x-dev-branch': 'b' },
      body: { email: 'a@b.test', password: 'password' },
      fetch: async input => {
        sent = input as Request
        return new Response(null, { status: 401 })
      },
    })
    expect(Object.fromEntries(sent!.headers)).toMatchObject({
      'content-type': 'application/json',
      origin: 'https://account.tadoku.app',
      cookie: 'tadoku_session=a',
      'x-dev-branch': 'b',
    })
  })

  it('accepts an empty answer where the contract has no body', async () => {
    const result = await api.identityLogout({ baseUrl: 'https://api.test', fetch: respond(204, {}, null) })
    expect(result.error).toBeUndefined()
  })

  it('sends logout without a body but with the JSON type every identity write needs', async () => {
    let sent: Request | undefined
    await api.identityLogout({
      baseUrl: 'https://api.test',
      fetch: async input => {
        sent = input as Request
        return new Response(null, { status: 204 })
      },
    })
    expect(sent!.headers.get('content-type')).toBe('application/json')
    expect(await sent!.text()).toBe('')
  })

  it('sends other bodyless writes without a content type', async () => {
    let sent: Request | undefined
    await api.immersionLogDeleteById({
      baseUrl: 'https://api.test',
      path: { id: '00000000-0000-4000-8000-000000000001' },
      fetch: async input => {
        sent = input as Request
        return new Response(null, { status: 200 })
      },
    })
    expect(sent!.headers.get('content-type')).toBeNull()
  })
})

describe('generated client on success answers the contract does not document', () => {
  const answer = (status: number, body: string | null, headers: Record<string, string> = {}) =>
    ({ baseUrl: 'https://api.test', fetch: respond(status, headers, body) })
  const json = { 'Content-Type': 'application/json' }
  const empty = { 'Content-Length': '0' }
  const password = { body: { password: 'password123' } }
  const recovery = { body: { email: 'reader@example.test', return_to: 'https://tadoku.app' } }
  const log = { path: { id: '00000000-0000-4000-8000-000000000001' } }
  const contest = { path: { id: '00000000-0000-4000-8000-000000000001' } }

  it.each([
    ['a session under 201', () => api.identitySessionGet(answer(201, '{"account":null,"recent_login":false}', json))],
    ['an empty 200 for a password change documented as 204',
      () => api.identityPasswordUpdate({ ...answer(200, null, empty), ...password })],
    ['JSON under 200 for a recovery request documented as an empty 202',
      () => api.identityRecoveryCreate({ ...answer(200, '{"unrelated":true}', json), ...recovery })],
    ['a body on the documented empty 202 of a recovery request',
      () => api.identityRecoveryCreate({ ...answer(202, '{}', json), ...recovery })],
    ['a pending 202 for a log deletion documented as an empty 200',
      () => api.immersionLogDeleteById({ ...answer(202, '{"pending":true}', json), ...log })],
    ['an empty 200 for a registration lookup documented as JSON 200 or empty 204',
      () => api.immersionContestFindRegistration({ ...answer(200, null, empty), ...contest })],
  ])('rejects %s', async (_, send) => {
    const result = await send()
    expect(result.error).toBeDefined()
    expect(result.data).toBeUndefined()
  })

  it.each([
    ['204 for a password change', () => api.identityPasswordUpdate({ ...answer(204, null), ...password })],
    ['an empty 202 for a recovery request', () => api.identityRecoveryCreate({ ...answer(202, null, empty), ...recovery })],
    ['an empty 200 for a log deletion', () => api.immersionLogDeleteById({ ...answer(200, null, empty), ...log })],
    ['204 for a registration lookup', () => api.immersionContestFindRegistration({ ...answer(204, null), ...contest })],
  ])('accepts %s', async (_, send) => {
    expect((await send()).error).toBeUndefined()
  })

  it('keeps an empty denial a denial with its status', async () => {
    const result = await api.identityPasswordUpdate({ ...answer(403, null, empty), ...password })
    expect(result.error).toBeDefined()
    expect(result.response?.status).toBe(403)
  })
})

describe('generated client on values the contract constrains', () => {
  const answer = (body: unknown) => ({
    baseUrl: 'https://api.test',
    fetch: respond(200, { 'Content-Type': 'application/json' }, JSON.stringify(body)),
  })
  const account = {
    id: '00000000-0000-4000-8000-000000000001',
    email: 'reader@example.test',
    display_name: 'Reader',
    created_at: '2026-01-01T00:00:00Z',
    authenticated_at: '2026-01-01T00:00:00Z',
  }
  const session = (fields: Record<string, string>) =>
    api.identitySessionGet(answer({ account: { ...account, ...fields }, recent_login: true }))

  it.each([
    ['empty account fields', { id: '', email: '', display_name: '', created_at: '', authenticated_at: '' }],
    ['an ID that is not a UUID', { id: '../arbitrary' }],
    ['an email without a domain', { email: 'reader' }],
    ['a display name of one character', { display_name: 'R' }],
    ['a display name of 33 characters', { display_name: '😀'.repeat(33) }],
    ['a login time that is not a date-time', { authenticated_at: 'yesterday' }],
  ])('rejects a session with %s', async (_, fields) => {
    expect((await session(fields)).error).toBeDefined()
  })

  it.each([
    // The API counts characters; 25 emoji are 50 UTF-16 units.
    ['a 25-emoji display name', { display_name: '😀'.repeat(25) }],
    ['a historical UUID without RFC version/variant bits', { id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' }],
    // Go's net/mail accepts a domain without a dot.
    ['a local-domain email', { email: 'reader@localhost' }],
    ['a time with a UTC offset and fractional seconds', { created_at: '2026-01-01T09:00:00.123456+09:00' }],
  ])('accepts a session with %s', async (_, fields) => {
    expect((await session(fields)).error).toBeUndefined()
  })

  it('rejects a post published at a time that is not a date-time', () => {
    const post = { slug: 'hello', title: 'Hello', content: 'Hello', published_at: 'not-a-date' }
    expect(api.zContentPost.safeParse(post).success).toBe(false)
  })

  it('accepts the recorded 100-post page and rejects a 101st post', () => {
    const recorded = readFileSync(join(goldens, 'ListPosts/200_page_size_capped/golden.http'), 'utf8')
    const page = JSON.parse(recorded.slice(recorded.indexOf('\n\n') + 2)) as { posts: unknown[] }
    expect(page.posts).toHaveLength(100)
    expect(api.zContentPosts.safeParse(page).success).toBe(true)
    expect(api.zContentPosts.safeParse({ ...page, posts: [...page.posts, page.posts[0]] }).success).toBe(false)
  })
})
