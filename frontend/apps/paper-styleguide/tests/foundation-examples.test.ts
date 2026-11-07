import { readdirSync } from 'node:fs'
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
