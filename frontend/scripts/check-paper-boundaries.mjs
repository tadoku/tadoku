import { readFile, readdir } from 'node:fs/promises'
import { extname, join, relative } from 'node:path'
import process from 'node:process'

const root = new URL('..', import.meta.url)
const config = JSON.parse(await readFile(new URL('../paper-boundaries.json', import.meta.url), 'utf8'))
if (config.schemaVersion !== 1) throw new Error(`Unsupported Paper boundary schema: ${config.schemaVersion}`)
const sourceExtensions = new Set(['.js', '.jsx', '.mjs', '.cjs', '.ts', '.tsx', '.css'])
const ignoredDirectories = new Set(['.next', '.output', '.tanstack', 'dist', 'node_modules', 'coverage'])
const legacyOnlyClasses = new Set([
  'auto-format',
  'h-stack',
  'input-frame',
  'modal-actions',
  'modal-body',
  'secondary',
  'table-container',
  'v-stack',
])
const violations = []

async function sourceFiles(directory) {
  const files = []
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    if (ignoredDirectories.has(entry.name)) continue
    const path = join(directory, entry.name)
    if (entry.isDirectory()) files.push(...(await sourceFiles(path)))
    else if (sourceExtensions.has(extname(entry.name))) files.push(path)
  }
  return files
}

function report(path, message) {
  violations.push(`${relative(new URL('..', root).pathname, path)}: ${message}`)
}

function legacyClassesInMarkup(source) {
  const found = new Set()
  const attributes = source.matchAll(/class(?:Name)?\s*=\s*(?:["']([^"']*)["']|\{`([^`]*)`\})/g)
  for (const attribute of attributes) {
    const value = attribute[1] || attribute[2] || ''
    const classes = value.split(/\s+/)
    for (const className of classes) {
      if (legacyOnlyClasses.has(className)) found.add(className)
    }
    if (classes.includes('btn') && classes.includes('primary')) found.add('btn primary')
  }
  return [...found]
}

function moduleSpecifiers(source) {
  // Consume literals whole so tutorial snippets cannot introduce import tokens.
  // Comments are discarded between tokens, including between `from` and a path.
  const tokens = (source.match(/\/\*[\s\S]*?\*\/|\/\/[^\n]*|"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\])*'|`(?:\\[\s\S]|[^`\\])*`|[\w$]+|[^\s]/g) || [])
    .filter((token) => !token.startsWith('//') && !token.startsWith('/*'))
  const modules = []
  for (let index = 0; index < tokens.length; index += 1) {
    const token = tokens[index]
    if (token !== 'import' && token !== 'from' && token !== 'require') continue
    const next = tokens[index + 1] === '(' ? tokens[index + 2] : tokens[index + 1]
    if (next?.startsWith('"') || next?.startsWith("'")) modules.push(next.slice(1, -1))
  }
  return modules
}

const ENTRY_STYLESHEET = 'src/styles.css'
const ENTRY_RULE = 'entry stylesheet may only contain @import, @tailwind, @source, @custom-variant, @layer statements and @font-face'

/** The top-level statements of an entry stylesheet that are not declarative Tailwind or font set-up. */
function entryProblems(source) {
  const text = source.replace(/\/\*[\s\S]*?\*\//g, '')
  const problems = []
  let index = 0
  const blockEnd = (start) => {
    let depth = 0
    for (let at = start; at < text.length; at += 1) {
      if (text[at] === '{') depth += 1
      else if (text[at] === '}' && (depth -= 1) === 0) return at + 1
    }
    return text.length
  }
  while (index < text.length) {
    if (/\s/.test(text[index])) { index += 1; continue }
    const rest = text.slice(index)
    const name = rest.match(/^@([\w-]+)/)?.[1]
    if (name === 'font-face' || (name === 'custom-variant' && /^[^;{]*\{/.test(rest))) {
      index = blockEnd(index)
    } else if (name === 'import' || name === 'tailwind' || name === 'source' || name === 'layer' || name === 'custom-variant') {
      if (name === 'layer' && /^@layer[^;{]*\{/.test(rest)) {
        problems.push(rest.slice(0, 40).trim())
        index = blockEnd(index)
        continue
      }
      const end = text.indexOf(';', index)
      index = end < 0 ? text.length : end + 1
    } else {
      problems.push(rest.slice(0, 40).trim())
      index = rest.includes('{') ? blockEnd(index) : text.length
    }
  }
  return problems
}

async function inspectStylesheets(path, application) {
  for (const file of await sourceFiles(path)) {
    if (extname(file) !== '.css') continue
    if (application && relative(path, file) === ENTRY_STYLESHEET) {
      const source = await readFile(file, 'utf8')
      if (entryProblems(source).length) report(file, ENTRY_RULE)
      if (!/@import\s+["']paper-ui\/theme\.css["']/.test(source)) report(file, 'entry stylesheet does not import paper-ui/theme.css')
    } else {
      report(file, `handwritten stylesheet; style with Paper's token utilities, and only ${ENTRY_STYLESHEET} may exist`)
    }
  }
}

async function inspectTree(path, system) {
  try {
    const manifestPath = join(path, 'package.json')
    const manifest = JSON.parse(await readFile(manifestPath, 'utf8'))
    const dependencies = {
      ...manifest.dependencies,
      ...manifest.devDependencies,
      ...manifest.peerDependencies,
    }
    if (system === 'paper') {
      for (const dependency of ['ui', 'next', '@headlessui/react']) {
        if (dependency in dependencies) report(manifestPath, `${dependency} dependency in Paper code`)
      }
    } else if ('paper-ui' in dependencies) {
      report(manifestPath, 'paper-ui dependency in an unmigrated application')
    }
  } catch (error) {
    if (error?.code !== 'ENOENT') throw error
  }

  for (const file of await sourceFiles(path)) {
    const source = await readFile(file, 'utf8')
    const modules = moduleSpecifiers(source)
    if (modules.some((name) => name.startsWith('paper-ui/src/'))) report(file, 'private paper-ui source import')

    if (system === 'paper') {
      if (modules.some((name) => name === 'ui' || name.startsWith('ui/'))) report(file, 'legacy ui import in Paper code')
      if (modules.some((name) => name === 'next' || name.startsWith('next/') || name === '@headlessui/react' || name.startsWith('@headlessui/react/'))) report(file, 'Next or Headless UI import in Paper code')
      const legacyClasses = legacyClassesInMarkup(source)
      if (legacyClasses.length) report(file, `legacy-only classes in Paper markup: ${legacyClasses.join(', ')}`)
    } else if (modules.some((name) => name === 'paper-ui' || name.startsWith('paper-ui/'))) {
      report(file, 'paper-ui import in an unmigrated application')
    }
  }
}

for (const [application, system] of Object.entries(config.applications)) {
  const path = new URL(`../apps/${application}`, import.meta.url).pathname
  try {
    await inspectTree(path, system)
    if (system === 'paper') await inspectStylesheets(path, true)
  } catch (error) {
    if (error?.code !== 'ENOENT' || application !== 'paper-styleguide') throw error
  }
}

const paperPackage = new URL('../packages/paper-ui', import.meta.url).pathname
try {
  await inspectTree(paperPackage, 'paper')
  await inspectStylesheets(paperPackage, false)
} catch (error) {
  if (error?.code !== 'ENOENT') throw error
}

if (violations.length) {
  console.error(violations.join('\n'))
  process.exitCode = 1
} else {
  console.log('Paper package boundaries are valid, and Paper code has no handwritten stylesheets.')
}
