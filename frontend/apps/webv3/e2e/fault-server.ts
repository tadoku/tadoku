import { spawn } from 'node:child_process'
import { createServer, type IncomingMessage } from 'node:http'
import type { AddressInfo } from 'node:net'

/**
 * A failure the proxy answers in place of the API for requests whose path and query match `path`, counting each one in
 * `hits`.
 */
export type Fault = { path: RegExp; hits?: number } & (
  | { kind: 'status'; status: number }
  | { kind: 'reset' }
  | { kind: 'malformed'; body: string }
  | { kind: 'slow'; delayMs: number }
)

const hopHeaders = new Set(['host', 'connection', 'keep-alive', 'transfer-encoding', 'content-length', 'content-encoding'])

const read = (request: IncomingMessage) => new Promise<Buffer>((resolve, reject) => {
  const chunks: Buffer[] = []
  request.on('data', chunk => chunks.push(chunk)).on('end', () => resolve(Buffer.concat(chunks))).on('error', reject)
})

const freePort = () => new Promise<number>((resolve, reject) => {
  const server = createServer().listen(0, '127.0.0.1', () => {
    const { port } = server.address() as AddressInfo
    server.close(() => resolve(port))
  }).on('error', reject)
})

/**
 * Runs this checkout's Webv3 development server with its API behind a local proxy that forwards to `upstream` and
 * answers with `faults` instead, so the server's own upstream reads, including those of server rendering, meet them.
 * The browser cannot reach those reads, and the deployed branch's server cannot reach this machine.
 */
export async function faultyWebv3(appDir: string, upstream: string) {
  const faults: Fault[] = []
  const proxy = createServer(async (request, response) => {
    const path = request.url!.replace(/^\/api\/internal/, '')
    const fault = faults.find(each => each.path.test(path))
    if (fault) fault.hits = (fault.hits ?? 0) + 1
    if (fault?.kind === 'reset') return request.socket.destroy()
    if (fault?.kind === 'status') return response.writeHead(fault.status, { 'Content-Type': 'application/json' }).end('{}')
    if (fault?.kind === 'malformed') return response.writeHead(200, { 'Content-Type': 'application/json' }).end(fault.body)
    if (fault?.kind === 'slow') await new Promise(resolve => setTimeout(resolve, fault.delayMs))

    try {
      const headers = Object.entries(request.headers).filter(([name]) => !hopHeaders.has(name)) as [string, string][]
      const body = ['GET', 'HEAD'].includes(request.method!) ? undefined : new Uint8Array(await read(request))
      const answer = await fetch(`${upstream}${path}`, { method: request.method, headers, body, redirect: 'manual' })
      const forwarded = [...answer.headers].filter(([name]) => !hopHeaders.has(name) && name !== 'set-cookie')
      response.writeHead(answer.status, [...forwarded, ...answer.headers.getSetCookie().map(cookie => ['set-cookie', cookie])].flat())
      response.end(Buffer.from(await answer.arrayBuffer()))
    } catch {
      if (!response.headersSent) response.writeHead(502).end()
    }
  })
  await new Promise<void>(resolve => proxy.listen(0, '127.0.0.1', resolve))
  const apiUrl = `http://127.0.0.1:${(proxy.address() as AddressInfo).port}/api/internal`

  const port = await freePort()
  const server = spawn('pnpm', ['exec', 'vite', '--strictPort'], {
    cwd: appDir,
    detached: true,
    stdio: 'ignore',
    env: { ...process.env, HOST: '127.0.0.1', PORT: String(port), TADOKU_API_URL: apiUrl },
  })
  const url = `http://127.0.0.1:${port}`
  const close = async () => {
    try {
      process.kill(-server.pid!, 'SIGTERM')
    } catch {
      // Already stopped.
    }
    await new Promise(resolve => proxy.close(resolve))
  }

  for (const started = Date.now(); ; await new Promise(resolve => setTimeout(resolve, 500))) {
    const healthy = await fetch(`${url}/healthz`).then(response => response.ok, () => false)
    if (healthy) break
    if (server.exitCode !== null || Date.now() - started > 120_000) {
      await close()
      throw new Error(`The local Webv3 server did not start (exit ${server.exitCode})`)
    }
  }
  return { url, faults, close }
}
