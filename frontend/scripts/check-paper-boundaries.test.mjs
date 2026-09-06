import assert from 'node:assert/strict'
import { mkdtemp, mkdir, readFile, rm, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import process from 'node:process'
import { spawnSync } from 'node:child_process'
import test from 'node:test'

async function check(files) {
  const directory = await mkdtemp(join(tmpdir(), 'paper-boundaries-'))
  try {
    await mkdir(join(directory, 'scripts'))
    await mkdir(join(directory, 'apps', 'example'), { recursive: true })
    await writeFile(join(directory, 'scripts', 'check-paper-boundaries.mjs'), await readFile(new URL('./check-paper-boundaries.mjs', import.meta.url)))
    await writeFile(join(directory, 'paper-boundaries.json'), JSON.stringify({ schemaVersion: 1, applications: { example: 'paper' } }))
    for (const [name, source] of Object.entries(files)) await writeFile(join(directory, 'apps', 'example', name), source)
    return spawnSync(process.execPath, [join(directory, 'scripts', 'check-paper-boundaries.mjs')], { encoding: 'utf8' })
  } finally {
    await rm(directory, { recursive: true, force: true })
  }
}

test('counts executable imports without treating examples, comments, or prose as imports', async () => {
  const result = await check({
    'main.ts': 'import "paper-ui/styles.css";',
    'examples.tsx': [
      'const stylesheet = "paper-ui/styles.css";',
      "const snippet = 'import \"paper-ui/styles.css\";';",
      'const example = `import { Input } from "ui";\nimport "paper-ui/styles.css";\nimport { Button } from "paper-ui/src/private";`;',
      '// import "paper-ui/styles.css"; import { Button } from "ui";',
      '/* import "paper-ui/styles.css"; import Link from "next/link"; */',
    ].join('\n'),
  })
  assert.equal(result.status, 0, result.stderr)
})

test('still detects real duplicate stylesheet imports including dynamic and require calls', async () => {
  const result = await check({
    'main.ts': 'import /* root stylesheet */ "paper-ui/styles.css";',
    'other.ts': 'import("paper-ui/styles.css"); require("paper-ui/styles.css");',
  })
  assert.equal(result.status, 1)
  assert.match(result.stderr, /styles\.css imported 3 times/)
})

test('still rejects real multiline, re-export, dynamic, and private imports', async () => {
  const result = await check({
    'main.ts': 'import "paper-ui/styles.css";',
    'bad.ts': 'import {\n Input\n} from /* migration mistake */ "ui";\nexport { Link } from "next/link";\nconst menu = import("@headlessui/react");\nconst button = require("paper-ui/src/private");',
  })
  assert.equal(result.status, 1)
  assert.match(result.stderr, /legacy ui import/)
  assert.match(result.stderr, /Next or Headless UI import/)
  assert.match(result.stderr, /private paper-ui source import/)
})
