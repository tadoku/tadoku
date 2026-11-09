import { defineHandler } from 'nitro'
import { isWebHost } from '../../src/lib/urls.server'

/** Crawlers may read the web host; every other host, such as `next.` or a branch, is closed to them. */
export default defineHandler(event => new Response(
  `User-agent: *\nDisallow:${isWebHost(event.req.headers.get('host')) ? '' : ' /'}\n`,
  { headers: { 'Content-Type': 'text/plain; charset=utf-8', 'Cache-Control': 'public, max-age=3600' } },
))
