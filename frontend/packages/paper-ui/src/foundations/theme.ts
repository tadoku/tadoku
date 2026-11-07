import * as stylex from '@stylexjs/stylex'
import { Fragment, createElement, type ReactNode } from 'react'
import { comfortableDensity, compactDensity, darkTheme, lightTheme } from '../tokens/themes'
import { color, density, font, themeScope } from '../tokens/tokens.stylex'

export const themes = { light: lightTheme, dark: darkTheme } as const
export const densities = { comfortable: comfortableDensity, compact: compactDensity } as const
export type PaperTheme = keyof typeof themes
export type PaperDensity = keyof typeof densities

const root = stylex.create({
  html: {
    backgroundColor: color.surfaceCanvas,
    color: color.textInk,
    colorScheme: color.scheme,
    fontFamily: font.body,
    fontSize: '100%',
    lineHeight: density.bodyLineHeight,
    textSizeAdjust: '100%',
    // Always draw the scrollbar track, so content does not shift when a page grows long enough to scroll.
    overflowY: 'scroll',
  },
  body: {
    margin: 0,
    backgroundColor: color.surfaceCanvas,
    color: color.textInk,
    fontSize: density.bodySize,
    minHeight: '100%',
  },
})

/** Spread on `<html>` and `<body>`; the theme and density classes are added by `themeBootScript` or `applyTheme`. */
export const rootProps = {
  html: () => stylex.props(themeScope, root.html),
  body: () => stylex.props(root.body),
}

const classNames = <Name extends string>(values: Record<Name, object>) =>
  Object.fromEntries(
    Object.entries(values).map(([name, theme]) => [name, stylex.props(theme as stylex.StyleXStyles).className ?? '']),
  ) as Record<Name, string>

const themeClassNames = classNames(themes)
const densityClassNames = classNames(densities)

function swap(element: HTMLElement, options: Record<string, string>, selected: string) {
  for (const className of Object.values(options)) element.classList.remove(...className.split(' ').filter(Boolean))
  element.classList.add(...options[selected].split(' ').filter(Boolean))
}

/** Applies a theme to an element, by default the document root, so portalled menus and dialogs inherit it. */
export function applyTheme(theme: PaperTheme, element: HTMLElement = document.documentElement) {
  swap(element, themeClassNames, theme)
  element.classList.add(...(stylex.props(themeScope).className ?? '').split(' ').filter(Boolean))
  element.dataset.theme = theme
}

export function applyDensity(value: PaperDensity, element: HTMLElement = document.documentElement) {
  swap(element, densityClassNames, value)
  element.dataset.density = value
}

/**
 * An inline script for `<head>` that applies the stored theme, or the system preference, before first paint.
 * Read the current theme back from `document.documentElement.dataset.theme`.
 */
export function themeBootScript(storageKey: string) {
  const classes = JSON.stringify(themeClassNames)
  return (
    `try{var c=${classes},s=localStorage.getItem(${JSON.stringify(storageKey)}),` +
    `t=s==='dark'||s==='light'?s:matchMedia('(prefers-color-scheme: dark)').matches?'dark':'light',` +
    `e=document.documentElement;e.dataset.theme=t;c[t].split(' ').forEach(function(n){n&&e.classList.add(n)})}catch(x){}`
  )
}


const variants = stylex.create({
  light: { display: { default: 'contents', [stylex.when.ancestor('[data-theme="dark"]', themeScope)]: 'none' } },
  dark: { display: { default: 'none', [stylex.when.ancestor('[data-theme="dark"]', themeScope)]: 'contents' } },
})

/**
 * Renders both variants and shows the one for the theme on the nearest `applyTheme` or `rootProps.html` element, so
 * server-rendered pages show the right logo or label before hydration.
 */
export function ThemeVariant({ light, dark }: { readonly light: ReactNode; readonly dark: ReactNode }) {
  return createElement(
    Fragment,
    null,
    createElement('span', stylex.props(variants.light), light),
    createElement('span', stylex.props(variants.dark), dark),
  )
}
