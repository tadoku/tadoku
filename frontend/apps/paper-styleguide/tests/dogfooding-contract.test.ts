import { readFileSync, readdirSync } from 'node:fs'
import { dirname, extname, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { DOGFOODING_DEBT, type DogfoodingDebtKind } from './dogfooding-inventory'
import { importedStyleSources } from './style-sources'

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
  kind: DogfoodingDebtKind,
  source: string,
  pattern: RegExp,
): string[] {
  return [...source.matchAll(pattern)].map(
    (_, index) => `${sourceKey(path)}:${kind}#${index + 1}`,
  )
}

function directSurfaceStyleKeys(): string[] {
  const keys = new Set<string>()

  for (const { key, source } of importedStyleSources) {
    const rules = source.matchAll(/([^{}]+)\{([^{}]*)\}/gu)

    for (const rule of rules) {
      if (!/background:\s*var\(--paper-color-surface-[^)]+\)/u.test(rule[2])) {
        continue
      }
      for (const selector of rule[1].split(',')) {
        const normalized = selector.trim().replace(/\s+/gu, ' ')
        if (normalized.startsWith('.')) {
          keys.add(`${key}:surface-style:${normalized}`)
        }
      }
    }
  }

  return [...keys]
}

function discoveredDogfoodingDebt(): string[] {
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

  return [...keys, ...directSurfaceStyleKeys()].sort()
}

function importedPackages(source: string): string[] {
  return [
    ...source.matchAll(/(?:from\s+|import\s*(?:\(\s*)?)['"]([^'"]+)['"]/gu),
  ].map((match) => match[1])
}

describe('Paper styleguide dogfooding boundary', () => {
  it('deduplicates React and form context across the workspace-linked Paper package', () => {
    expect(viteConfig).toContain(
      "dedupe: ['react', 'react-dom', 'react-hook-form']",
    )
  })

  it('imports only public Paper APIs rather than legacy or primitive internals', () => {
    const forbidden = applicationSources.flatMap((path) =>
      importedPackages(readFileSync(path, 'utf8'))
        .filter(
          (specifier) =>
            specifier === 'ui' ||
            specifier.startsWith('ui/') ||
            specifier.startsWith('@base-ui/') ||
            specifier.startsWith('@headlessui/') ||
            specifier === 'paper-ui/src' ||
            specifier.startsWith('paper-ui/src/') ||
            specifier.includes('packages/paper-ui/src'),
        )
        .map((specifier) => `${sourceKey(path)} -> ${specifier}`),
    )

    expect(forbidden).toEqual([])
  })

  it('keeps every remaining app-owned primitive and surface in the migration ledger', () => {
    const inventoryKeys = DOGFOODING_DEBT.map((item) => item.key).sort()
    expect(new Set(inventoryKeys).size).toBe(inventoryKeys.length)
    expect(DOGFOODING_DEBT.every((item) => item.context && item.destination)).toBe(true)
    expect(inventoryKeys).toEqual(discoveredDogfoodingDebt())
  })
})
