import { definePlugin } from 'nitro'

/**
 * Answers 400 for a path that does not decode, such as `/pages/%FF` or a lone `%`. Nitro's h3 throws an uncaught
 * URIError on one and never answers, which leaves the request hanging until the edge times it out.
 */
export default definePlugin(nitroApp => {
  const fetch = nitroApp.fetch
  nitroApp.fetch = request => {
    try {
      decodeURI(new URL(request.url).pathname)
    } catch {
      return new Response('Bad Request\n', { status: 400, headers: { 'Content-Type': 'text/plain', 'Cache-Control': 'no-store' } })
    }
    return fetch(request)
  }
})
