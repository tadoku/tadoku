import { cp, mkdir, readFile, rm, writeFile } from 'node:fs/promises'

const root = new URL('..', import.meta.url)
const dist = new URL('../dist/', import.meta.url)

await mkdir(dist, { recursive: true })

// StyleX requires `export const name = stylex.defineVars(...)`; tsup's bundle form exports names at the end instead.
const ts = (await import('typescript')).default
const tokenSource = await readFile(new URL('src/tokens/tokens.stylex.ts', root), 'utf8')
const compilerOptions = { module: ts.ModuleKind.ESNext, target: ts.ScriptTarget.ES2020 }
await writeFile(new URL('tokens.stylex.js', dist), ts.transpileModule(tokenSource, { compilerOptions }).outputText)

await rm(new URL('tokens.stylex.js.map', dist), { force: true })
await cp(new URL('src/foundations/fonts.css', root), new URL('fonts.css', dist))

for (const directory of ['src/assets']) {
  try {
    await cp(new URL(directory, root), new URL('assets', dist), { recursive: true })
  } catch (error) {
    if (error?.code !== 'ENOENT') throw error
  }
}

