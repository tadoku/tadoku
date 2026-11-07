import { readFileSync, readdirSync } from 'node:fs'
import { dirname, extname, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const testDirectory = dirname(fileURLToPath(import.meta.url))
const appRoot = resolve(testDirectory, '..')
const sourceRoot = resolve(appRoot, 'src')
const viteConfig = readFileSync(resolve(appRoot, 'vite.config.ts'), 'utf8')

function sourceFiles(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = resolve(directory, entry.name)
    return entry.isDirectory() ? sourceFiles(path) : [path]
  })
}

const applicationSources = sourceFiles(sourceRoot).filter((path) =>
  ['.ts', '.tsx'].includes(extname(path)),
)

function sourceKey(path: string): string {
  return relative(appRoot, path).replace(/\\/gu, '/')
}

function ordinalKeys(
  path: string,
  kind: string,
  source: string,
  pattern: RegExp,
): string[] {
  return [...source.matchAll(pattern)].map(
    (_, index) => `${sourceKey(path)}:${kind}#${index + 1}`,
  )
}

function discoveredForbiddenPrimitives(): string[] {
  const keys = applicationSources.flatMap((path) => {
    const source = readFileSync(path, 'utf8')
    return [
      ...ordinalKeys(path, 'native-button', source, /<button\b/gu),
      ...ordinalKeys(path, 'native-input', source, /<input\b/gu),
      ...ordinalKeys(path, 'native-select', source, /<select\b/gu),
      ...ordinalKeys(path, 'manual-tablist', source, /role\s*=\s*['"]tablist['"]/gu),
      ...ordinalKeys(
        path,
        'manual-focus-trap',
        source,
        /(?:function|const)\s+(?:[Tt]rap.*[Ff]ocus|[Ff]ocus.*[Tt]rap)[\w$]*/gu,
      ),
    ]
  })

  return keys.sort()
}

function importedPackages(source: string): string[] {
  return [
    ...source.matchAll(/(?:from\s+|import\s*(?:\(\s*)?)['"]([^'"]+)['"]/gu),
  ].map((match) => match[1])
}

describe('Paper styleguide dogfooding boundary', () => {
  it('deduplicates React and form context across the workspace-linked Paper package', () => {
    expect(viteConfig).toContain(
      "dedupe: ['react', 'react-dom', '@tanstack/react-form']",
    )
  })

  it('imports only public Paper APIs rather than legacy or primitive internals', () => {
    const paperSourcePath = ['paper-ui', 'src'].join('/')
    const workspacePaperSourcePath = ['packages', paperSourcePath].join('/')
    const forbidden = applicationSources.flatMap((path) =>
      importedPackages(readFileSync(path, 'utf8'))
        .filter(
          (specifier) =>
            specifier === 'ui' ||
            specifier.startsWith('ui/') ||
            specifier.startsWith('@base-ui/') ||
            specifier.startsWith('@headlessui/') ||
            specifier === paperSourcePath ||
            specifier.startsWith(`${paperSourcePath}/`) ||
            specifier.includes(workspacePaperSourcePath),
        )
        .map((specifier) => `${sourceKey(path)} -> ${specifier}`),
    )

    expect(forbidden).toEqual([])
  })

  it('has no app-owned form primitives, tablist, or overlay focus trap', () => {
    expect(discoveredForbiddenPrimitives()).toEqual([])
  })
})
