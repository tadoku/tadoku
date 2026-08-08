import { readFileSync, readdirSync } from 'node:fs'
import { dirname, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const testDirectory = dirname(fileURLToPath(import.meta.url))
export const styleguideRoot = resolve(testDirectory, '..')
export const stylesDirectory = resolve(styleguideRoot, 'src/styles')
export const styleEntryPath = resolve(stylesDirectory, 'shell.css')
export const styleEntrySource = readFileSync(styleEntryPath, 'utf8')

export const expectedStyleImports = [
  './base.css',
  './shell-layout.css',
  './controls.css',
  './navigation.css',
  './documents.css',
  './canvas-frame.css',
  './workbench.css',
  './canvas.css',
  './overlays.css',
  './responsive.css',
] as const

export const importedStyleSpecifiers = [
  ...styleEntrySource.matchAll(/@import\s+['"]([^'"]+)['"];?/gu),
].map((match) => match[1])

export interface StyleSource {
  readonly path: string
  readonly key: string
  readonly source: string
}

export const importedStyleSources: readonly StyleSource[] =
  importedStyleSpecifiers.map((specifier) => {
    const path = resolve(stylesDirectory, specifier)
    return {
      path,
      key: relative(styleguideRoot, path).replace(/\\/gu, '/'),
      source: readFileSync(path, 'utf8'),
    }
  })

export const styleguideStyles = importedStyleSources
  .map(({ source }) => source)
  .join('\n')

export const stylesheetNames = readdirSync(stylesDirectory)
  .filter((name) => name.endsWith('.css'))
  .sort()
