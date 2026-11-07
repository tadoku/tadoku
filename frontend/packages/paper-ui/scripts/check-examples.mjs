import { cpSync, mkdtempSync, readdirSync, rmSync } from 'node:fs';
import { join } from 'node:path';
import ts from 'typescript';
import process from 'node:process';

// Compile what readers copy, against the published package declarations. Every catalog fixture's code is the raw
// source of one file in src/catalog/examples, copied outside src so it resolves `paper-ui` like a consumer would.
const source = join(process.cwd(), 'src/catalog/examples');
const directory = mkdtempSync(join(process.cwd(), '.example-check-'));
try {
  const files = readdirSync(source).filter((name) => name.endsWith('.tsx')).map((name) => {
    cpSync(join(source, name), join(directory, name));
    return join(directory, name);
  });
  const program = ts.createProgram(files, {
    noEmit: true, strict: true, noImplicitAny: false, skipLibCheck: true,
    jsx: ts.JsxEmit.ReactJSX, target: ts.ScriptTarget.ES2020,
    module: ts.ModuleKind.ESNext, moduleResolution: ts.ModuleResolutionKind.Bundler,
    esModuleInterop: true, types: ['react', 'react-dom'],
  });
  const diagnostics = ts.getPreEmitDiagnostics(program);
  if (diagnostics.length) {
    console.error(ts.formatDiagnosticsWithColorAndContext(diagnostics, {
      getCurrentDirectory: () => process.cwd(),
      getCanonicalFileName: (name) => name,
      getNewLine: () => '\n',
    }));
    process.exitCode = 1;
  } else console.log(`${files.length} copyable examples typechecked against Paper's published API.`);
} finally {
  rmSync(directory, { recursive: true, force: true });
}
