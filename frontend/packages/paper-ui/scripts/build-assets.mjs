import { cp, mkdir, readFile, readdir, writeFile } from 'node:fs/promises'
import { extname, join } from 'node:path'

const root = new URL('..', import.meta.url)
const dist = new URL('../dist/', import.meta.url)

await mkdir(dist, { recursive: true })

async function existingText(relativePath) {
  try {
    return await readFile(new URL(relativePath, root), 'utf8')
  } catch (error) {
    if (error?.code === 'ENOENT') return ''
    throw error
  }
}

async function cssFiles(directory) {
  const found = []
  try {
    for (const entry of await readdir(new URL(directory, root), { withFileTypes: true })) {
      const relativePath = join(directory, entry.name)
      if (entry.isDirectory()) found.push(...(await cssFiles(relativePath)))
      else if (extname(entry.name) === '.css') found.push(relativePath)
    }
  } catch (error) {
    if (error?.code !== 'ENOENT') throw error
  }
  return found.sort()
}

const fonts = await existingText('src/foundations/fonts.css')
const tokens = await existingText('src/foundations/tokens.css')
const base = await existingText('src/foundations/base.css')
const styleParts = []
for (const file of await cssFiles('styles')) {
  if (!file.endsWith('index.css')) styleParts.push(await existingText(file))
}
for (const file of await cssFiles('src/components')) styleParts.push(await existingText(file))

await writeFile(new URL('fonts.css', dist), fonts)
await writeFile(new URL('tokens.css', dist), tokens)
await writeFile(
  new URL('styles.css', dist),
  [fonts, tokens, base, ...styleParts]
    .filter(Boolean)
    .join('\n\n'),
)

for (const directory of ['src/assets', 'styles/assets']) {
  try {
    await cp(new URL(directory, root), new URL('assets', dist), { recursive: true })
  } catch (error) {
    if (error?.code !== 'ENOENT') throw error
  }
}

await cp(new URL('tailwind-preset.cjs', root), new URL('tailwind-preset.cjs', dist))
