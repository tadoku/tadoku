import { readdirSync, readFileSync } from 'node:fs'
import { join, resolve } from 'node:path'
import ts from 'typescript'
import { expect, it } from 'vitest'

it('typechecks the actual foundation example modules against the public package', () => {
  const directory = resolve('src/documentation/foundation-examples')
  const files = readdirSync(directory).filter(name => name.endsWith('.tsx')).map(name => join(directory, name))
  expect(files).toHaveLength(10)
  const program = ts.createProgram(files, {
    noEmit: true, strict: true, skipLibCheck: true,
    jsx: ts.JsxEmit.ReactJSX, target: ts.ScriptTarget.ES2020,
    module: ts.ModuleKind.ESNext, moduleResolution: ts.ModuleResolutionKind.Bundler,
    esModuleInterop: true, types: ['vite/client', 'react'],
  })
  const diagnostics = ts.getPreEmitDiagnostics(program)
  expect(ts.formatDiagnosticsWithColorAndContext(diagnostics, {
    getCurrentDirectory: () => process.cwd(), getCanonicalFileName: name => name, getNewLine: () => '\n',
  })).toBe('')
}, 15000)

it('keeps foundation tables readable and stack spacing faithful to the public recipe', () => {
  const css = readFileSync(resolve('src/documentation/foundation-guide.css'), 'utf8')
  expect(css).toMatch(/\.foundation-reference table\s*\{[^}]*min-inline-size: 32rem/)
  expect(css).toMatch(/\.foundation-reference th\s*\{[^}]*min-inline-size: 8rem/)
  expect(css).toMatch(/\.foundation-guide \.paper-stack > h3\s*\{[^}]*margin-block: 0/)
})
