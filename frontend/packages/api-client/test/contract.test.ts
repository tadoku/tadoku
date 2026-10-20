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
    ['an unknown account app', 200, json, '{"account_app":"v3","account":null}'],
    ['a missing account', 200, json, '{"account_app":"v2"}'],
    ['an account without an ID', 200, json, '{"account_app":"v2","account":{"email":"a@b.test","display_name":"Ab",'
      + '"created_at":"2026-01-01T00:00:00Z","authenticated_at":"2026-01-01T00:00:00Z"}}'],
  ])('rejects %s', async (_, status, headers, body) => {
    const result = await session(status, headers, body)
    expect(result.data).toBeUndefined()
    expect(result.error).toBeDefined()
  })

  it('accepts a signed-out session', async () => {
    const result = await session(200, json, '{"account_app":"v2","account":null}')
    expect(result.error).toBeUndefined()
    expect(result.data).toEqual({ account_app: 'v2', account: null })
  })

  it('accepts an empty answer where the contract has no body', async () => {
    const result = await api.identityLogout({ baseUrl: 'https://api.test', body: {}, fetch: respond(204, {}, null) })
    expect(result.error).toBeUndefined()
  })
})
