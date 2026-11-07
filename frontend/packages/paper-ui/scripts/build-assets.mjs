import { cp, mkdir, writeFile } from 'node:fs/promises'

const root = new URL('..', import.meta.url)
const dist = new URL('../dist/', import.meta.url)
const { themeCss } = await import(new URL('tokens.js', dist).href + `?t=${Date.now()}`)

await mkdir(dist, { recursive: true })
await writeFile(new URL('theme.css', dist), themeCss())
await cp(new URL('src/assets', root), new URL('assets', dist), { recursive: true })
