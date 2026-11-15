import { definePlugin } from 'nitro'
import { logRequest } from '../../src/lib/log.server'

/** Gives every request, pages, server functions and Nitro routes alike, its log record and slow-response line. */
export default definePlugin(nitroApp => {
  const fetch = nitroApp.fetch
  nitroApp.fetch = request => logRequest(request, () => fetch(request))
})
