import { mkdtempSync, writeFileSync, rmSync } from 'node:fs';
import { join } from 'node:path';
import ts from 'typescript';
import process from 'node:process';
import { catalogRegistry } from '../dist/catalog.js';

// Compile what readers copy, against the published package declarations.
const directory = mkdtempSync(join(process.cwd(), '.example-check-'));
try {
  const files = catalogRegistry.fixtures.map((fixture) => {
    if (!fixture.code) throw new Error(`Missing consumer example: ${fixture.id}`);
    const filename = join(directory, `${fixture.id}.tsx`);
    writeFileSync(filename, fixture.code);
    return filename;
  });
  const cssDeclaration = join(directory, "styles.d.ts");
  writeFileSync(cssDeclaration, 'declare module "paper-ui/styles.css";\n');
  const program = ts.createProgram([...files, cssDeclaration], {
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
