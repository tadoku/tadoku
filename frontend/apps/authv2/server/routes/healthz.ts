import { defineHandler } from 'nitro'

export default defineHandler(() => new Response('ok\n', {
  headers: { 'Content-Type': 'text/plain', 'Cache-Control': 'no-store' },
}))
