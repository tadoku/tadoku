import { readFile, readdir } from 'node:fs/promises'
import { extname, join, relative } from 'node:path'
import process from 'node:process'

const root = new URL('..', import.meta.url)
const config = JSON.parse(await readFile(new URL('../paper-boundaries.json', import.meta.url), 'utf8'))
if (config.schemaVersion !== 1) throw new Error(`Unsupported Paper boundary schema: ${config.schemaVersion}`)
const sourceExtensions = new Set(['.js', '.jsx', '.mjs', '.cjs', '.ts', '.tsx', '.css'])
const ignoredDirectories = new Set(['.next', 'dist', 'node_modules', 'coverage'])
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

  let paperStyleImports = 0
  for (const file of await sourceFiles(path)) {
    const source = await readFile(file, 'utf8')
    if (/paper-ui\/src\//.test(source)) report(file, 'private paper-ui source import')
    if (/['"]paper-ui\/styles\.css['"]/.test(source)) paperStyleImports += 1

    if (system === 'paper') {
      if (/['"]ui(?:\/[^'"]*)?['"]/.test(source)) report(file, 'legacy ui import in Paper code')
      if (/['"](?:next(?:\/[^'"]*)?|@headlessui\/react)['"]/.test(source)) report(file, 'Next or Headless UI import in Paper code')
    } else if (/['"]paper-ui(?:\/[^'"]*)?['"]/.test(source)) {
      report(file, 'paper-ui import in an unmigrated application')
    }
  }
  if (paperStyleImports > 1) report(path, `paper-ui/styles.css imported ${paperStyleImports} times`)
}

for (const [application, system] of Object.entries(config.applications)) {
  const path = new URL(`../apps/${application}`, import.meta.url).pathname
  try {
    await inspectTree(path, system)
  } catch (error) {
    if (error?.code !== 'ENOENT' || application !== 'paper-styleguide') throw error
  }
}

const paperPackage = new URL('../packages/paper-ui', import.meta.url).pathname
try {
  await inspectTree(paperPackage, 'paper')
} catch (error) {
  if (error?.code !== 'ENOENT') throw error
}

if (violations.length) {
  console.error(violations.join('\n'))
  process.exitCode = 1
} else {
  console.log('Paper package boundaries are valid.')
}
