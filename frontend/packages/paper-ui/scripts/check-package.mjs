import { execFileSync } from 'node:child_process'
import { mkdtempSync, readFileSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const destination = mkdtempSync(join(tmpdir(), 'paper-ui-pack-'))
const output = execFileSync('pnpm', ['pack', '--json', '--pack-destination', destination], {
  encoding: 'utf8',
})
const parsedOutput = JSON.parse(output)
const result = Array.isArray(parsedOutput) ? parsedOutput[0] : parsedOutput
const files = new Set(result.files.map((file) => file.path))
const required = [
  'package.json',
  'dist/index.js',
  'dist/index.d.ts',
  'dist/icons.js',
  'dist/icons.d.ts',
  'dist/catalog.js',
  'dist/catalog.d.ts',
  'dist/tokens.stylex.js',
  'dist/tokens.stylex.d.ts',
  'dist/fonts.css',
  'eslint/index.mjs',
  'dist/assets/brand/cut-meter.svg',
  'dist/assets/brand/wordmark-accent.svg',
  'dist/assets/brand/wordmark-reversed.svg',
  'dist/assets/brand/wordmark.svg',
  'dist/assets/fonts/merriweather-700.woff2',
  'dist/assets/fonts/open-sans-400.woff2',
  'dist/assets/fonts/open-sans-600.woff2',
  'dist/assets/fonts/open-sans-700.woff2',
]

for (const file of required) {
  if (!files.has(file)) throw new Error(`Package is missing ${file}`)
}

for (const forbidden of ['next', '@headlessui/react']) {
  const forbiddenSpecifier = new RegExp(`["']${forbidden.replace('/', '\\/')}(?:\\/[^"']*)?["']`)
  for (const file of ['dist/index.js', 'dist/index.d.ts', 'dist/icons.js', 'dist/catalog.js']) {
    if (forbiddenSpecifier.test(readFileSync(new URL(`../${file}`, import.meta.url), 'utf8'))) {
      throw new Error(`${file} exposes forbidden dependency ${forbidden}`)
    }
  }
}

// Paper ships StyleX source for the consumer's compiler; fonts.css is its only stylesheet.
const stylesheets = [...files].filter((file) => file.endsWith('.css'))
if (stylesheets.join() !== 'dist/fonts.css') throw new Error(`Unexpected stylesheets: ${stylesheets.join(', ')}`)
const fonts = readFileSync(new URL('../dist/fonts.css', import.meta.url), 'utf8')
for (const forbidden of ['fonts.googleapis.com', 'fonts.gstatic.com']) {
  if (fonts.includes(forbidden)) throw new Error(`fonts.css contains forbidden value ${forbidden}`)
}

console.log(`Paper package contains ${files.size} validated files.`)
rmSync(destination, { recursive: true })
