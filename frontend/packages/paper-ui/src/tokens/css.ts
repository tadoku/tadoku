import { fontFaces } from './fonts'
import { effects } from './effects'
import { theme, themeDefaults, themeKeyframes, type Dense, type Responsive, type Themed, type TokenValue } from './theme'

const isThemed = (value: TokenValue): value is Themed => typeof value === 'object' && 'light' in value
const isDense = (value: TokenValue): value is Dense => typeof value === 'object' && 'comfortable' in value
const isResponsive = (value: TokenValue): value is Responsive => typeof value === 'object' && 'base' in value

/** Breakpoints from widest to narrowest, so a narrower override comes later and wins. */
const breakpoints = Object.entries(theme.breakpoint).sort(([, left], [, right]) => parseFloat(right) - parseFloat(left))

/** Runtime properties that Base UI sets on its positioners and popups. */
export const runtimeProperties = ['--anchor-width', '--transform-origin', '--available-height'] as const

type Mode = 'light' | 'dark' | 'comfortable' | 'compact' | (typeof breakpoints)[number][0]

interface Declaration {
  readonly name: string
  readonly value: TokenValue
}

function declarations(): Declaration[] {
  const all: Declaration[] = []
  for (const [namespace, tokens] of Object.entries(theme)) {
    for (const [key, value] of Object.entries(tokens as Record<string, TokenValue>)) {
      all.push({ name: `--${namespace}-${key}`, value })
    }
  }
  for (const [key, value] of Object.entries(effects)) all.push({ name: `--effect-${key}`, value })
  return all
}

const texts = (value: TokenValue) => (typeof value === 'string' ? [value] : Object.values(value))
const references = (value: TokenValue) =>
  texts(value).flatMap((text) => [...text.matchAll(/var\((--[\w-]+)/g)].map(([, reference]) => reference))

/** Sizing utilities read spacing before container and height, so one key in both would hide the other. */
function checkCollisions() {
  const spacing = new Set(Object.keys(theme.spacing))
  for (const namespace of ['container', 'height'] as const) {
    for (const key of Object.keys(theme[namespace])) {
      if (spacing.has(key)) throw new Error(`--${namespace}-${key} is hidden by --spacing-${key}; rename one`)
    }
  }
}

function check(all: Declaration[]) {
  checkCollisions()
  const known = new Set([...all.map(({ name }) => name), ...runtimeProperties])
  for (const { name, value } of all) {
    for (const reference of references(value)) {
      if (!known.has(reference)) throw new Error(`${name} references unknown token ${reference}`)
    }
  }
}

/** Which axes a token changes along: directly, or through the tokens its value reads. */
function axesOf(all: Declaration[]) {
  const byName = new Map(all.map((declaration) => [declaration.name, declaration]))
  const memo = new Map<string, { themed: boolean; dense: boolean; responsive: boolean }>()
  const visit = (name: string) => {
    const known = memo.get(name)
    if (known) return known
    const result = { themed: false, dense: false, responsive: false }
    memo.set(name, result)
    const declaration = byName.get(name)
    if (!declaration) return result
    result.themed = isThemed(declaration.value)
    result.dense = isDense(declaration.value)
    result.responsive = isResponsive(declaration.value)
    for (const reference of references(declaration.value)) {
      const axes = visit(reference)
      result.themed ||= axes.themed
      result.dense ||= axes.dense
      result.responsive ||= axes.responsive
    }
    return result
  }
  return visit
}

function pick(value: TokenValue, mode?: Mode): string {
  if (isThemed(value)) return mode === 'dark' ? value.dark : value.light
  if (isDense(value)) return mode === 'compact' ? value.compact : value.comfortable
  if (isResponsive(value)) {
    const narrowing = breakpoints.map(([key]) => key)
    const index = mode ? narrowing.indexOf(mode) : -1
    for (let step = index; step >= 0; step -= 1) {
      const override = value[narrowing[step] as keyof Responsive]
      if (override) return override
    }
    return value.base
  }
  return value
}

const block = (selector: string, lines: string[]) => `${selector} {\n${lines.map((line) => `  ${line}`).join('\n')}\n}`

/**
 * The stylesheet a Paper application imports after Tailwind. It registers Paper's tokens as Tailwind's whole theme,
 * declares the `dark:` and `compact:` variants, declares every token on the root with its light, comfortable and
 * widest value, re-declares what changes inside each `data-theme` and `data-density` scope and below each breakpoint,
 * registers Paper's built components as a Tailwind source and loads Paper's fonts. An application that renders the
 * catalogue adds `@source` for `paper-ui/dist/catalog.js` itself.
 */
export function themeCss(): string {
  const all = declarations()
  check(all)
  const axes = axesOf(all)

  const defaults = Object.entries(themeDefaults).map(([name, value]) => `--${name}: ${value};`)
  const registered = all.filter(({ name }) => !name.startsWith('--effect-')).map(({ name, value }) => `${name}: ${pick(value)};`)
  const scoped = (mode: Mode, axis: 'themed' | 'dense' | 'responsive') =>
    all.filter(({ name }) => axes(name)[axis]).map(({ name, value }) => `${name}: ${pick(value, mode)};`)
  const byName = new Map(all.map((declaration) => [declaration.name, declaration.value]))
  const changesAt = (name: string, key: string): boolean => {
    const value = byName.get(name)
    if (!value) return false
    return (isResponsive(value) && key in value) || references(value).some((reference) => changesAt(reference, key))
  }
  const narrower = breakpoints
    .map(([key, width]) => {
      const lines = all.filter(({ name }) => changesAt(name, key)).map(({ name, value }) => `${name}: ${pick(value, key)};`)
      return lines.length ? `@media (width < ${width}) {\n${block(':root', lines).replace(/^/gm, '  ')}\n}` : ''
    })
    .filter(Boolean)
  const keyframes = Object.entries(themeKeyframes).map(
    ([name, steps]) =>
      `@keyframes ${name} {\n${Object.entries(steps).map(([step, rule]) => `  ${step} { ${rule} }`).join('\n')}\n}`,
  )
  const faces = fontFaces.map(({ family, weight, file }) =>
    block('@font-face', [
      `font-family: "${family}";`,
      'font-style: normal;',
      `font-weight: ${weight};`,
      'font-display: swap;',
      `src: url("./assets/fonts/${file}") format("woff2");`,
    ]),
  )

  return [
    '/* Generated by paper-ui from src/tokens. Do not edit. */',
    '@source "./index.js";',
    '@source "./icons.js";',
    '@source "./chunks";',
    '@custom-variant dark (&:where([data-theme="dark"], [data-theme="dark"] *));',
    '@custom-variant compact (&:where([data-density="compact"], [data-density="compact"] *));',
    block('@theme reference', ['--*: initial;', ...defaults, ...registered]),
    '@layer theme {',
    block(':root', ['color-scheme: light;', ...defaults, ...all.map(({ name, value }) => `${name}: ${pick(value)};`)]),
    block('[data-theme="light"]', ['color-scheme: light;', ...scoped('light', 'themed')]),
    block('[data-theme="dark"]', ['color-scheme: dark;', ...scoped('dark', 'themed')]),
    block('[data-density="comfortable"]', scoped('comfortable', 'dense')),
    block('[data-density="compact"]', scoped('compact', 'dense')),
    ...narrower,
    '}',
    ...keyframes,
    ...faces,
    '',
  ].join('\n')
}
