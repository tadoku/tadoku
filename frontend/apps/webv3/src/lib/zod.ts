import { z } from 'zod'

// Zod probes for compiled parsers with `new Function` when a schema is built, which the CSP in `start.ts` reports as
// `unsafe-eval`. Import `z` from here, never from `zod`, so the configuration runs before any schema is built.
z.config({ jitless: true })

export { z }
