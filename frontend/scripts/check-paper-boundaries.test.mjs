import assert from 'node:assert/strict'
import { mkdtemp, mkdir, readFile, rm, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { dirname, join } from 'node:path'
import process from 'node:process'
import { spawnSync } from 'node:child_process'
import test from 'node:test'

const entry = '@import "paper-ui/theme.css";\n@tailwind utilities;\n@source "./";\n'

async function check(files) {
  const directory = await mkdtemp(join(tmpdir(), 'paper-boundaries-'))
  try {
    await mkdir(join(directory, 'scripts'))
    await mkdir(join(directory, 'apps', 'example'), { recursive: true })
    await writeFile(join(directory, 'scripts', 'check-paper-boundaries.mjs'), await readFile(new URL('./check-paper-boundaries.mjs', import.meta.url)))
    await writeFile(join(directory, 'paper-boundaries.json'), JSON.stringify({ schemaVersion: 1, applications: { example: 'paper' } }))
    for (const [name, source] of Object.entries(files)) {
      const path = name.startsWith('packages/') ? join(directory, name) : join(directory, 'apps', 'example', name)
      await mkdir(dirname(path), { recursive: true })
      await writeFile(path, source)
    }
    return spawnSync(process.execPath, [join(directory, 'scripts', 'check-paper-boundaries.mjs')], { encoding: 'utf8' })
  } finally {
    await rm(directory, { recursive: true, force: true })
  }
}

test('accepts one declarative entry stylesheet that imports Paper’s theme', async () => {
  const fonts = '@font-face {\n  font-family: "Local";\n  src: url("./local.woff2");\n}\n@custom-variant print (@media print);\n'
  const result = await check({ 'src/styles.css': entry + fonts, 'src/main.ts': 'import "./styles.css";' })
  assert.equal(result.status, 0, result.stderr)
})

test('rejects a handwritten stylesheet beside the entry', async () => {
  const result = await check({ 'src/styles.css': entry, 'src/pages/home.css': '.home { display: grid; }' })
  assert.equal(result.status, 1)
  assert.match(result.stderr, /home\.css: handwritten stylesheet/)
})

test('rejects style rules, @apply and @utility in the entry stylesheet', async () => {
  for (const rule of ['.card { padding: 1rem; }', '.card { @apply p-4; }', '@utility card { padding: 1rem; }', ':root { --gap: 3px; }']) {
    const result = await check({ 'src/styles.css': entry + rule })
    assert.equal(result.status, 1, rule)
    assert.match(result.stderr, /styles\.css: entry stylesheet may only/)
  }
})

test('requires the entry stylesheet to import Paper’s theme', async () => {
  const result = await check({ 'src/styles.css': '@import "tailwindcss";\n' })
  assert.equal(result.status, 1)
  assert.match(result.stderr, /does not import paper-ui\/theme\.css/)
})

test('rejects any stylesheet in paper-ui outside its build output', async () => {
  const result = await check({
    'src/styles.css': entry,
    'packages/paper-ui/package.json': '{}',
    'packages/paper-ui/src/components/card.css': '.card { padding: 1rem; }',
    'packages/paper-ui/dist/theme.css': '@theme reference { --color-ink: #000; }',
  })
  assert.equal(result.status, 1)
  assert.match(result.stderr, /card\.css: handwritten stylesheet/)
  assert.doesNotMatch(result.stderr, /theme\.css/)
})

test('counts executable imports without treating examples, comments, or prose as imports', async () => {
  const result = await check({
    'src/styles.css': entry,
    'examples.tsx': [
      'const snippet = \'import { Input } from "ui";\';',
      'const example = `import { Input } from "ui";\nimport { Button } from "paper-ui/src/private";`;',
      '// import { Button } from "ui";',
      '/* import Link from "next/link"; */',
    ].join('\n'),
  })
  assert.equal(result.status, 0, result.stderr)
})

test('still rejects real multiline, re-export, dynamic, and private imports', async () => {
  const result = await check({
    'src/styles.css': entry,
    'bad.ts': 'import {\n Input\n} from /* migration mistake */ "ui";\nexport { Link } from "next/link";\nconst menu = import("@headlessui/react");\nconst button = require("paper-ui/src/private");',
  })
  assert.equal(result.status, 1)
  assert.match(result.stderr, /legacy ui import/)
  assert.match(result.stderr, /Next or Headless UI import/)
  assert.match(result.stderr, /private paper-ui source import/)
})
