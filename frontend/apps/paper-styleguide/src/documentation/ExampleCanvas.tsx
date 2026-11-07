import * as stylex from '@stylexjs/stylex'
import { useStore } from '@tanstack/react-form'
import { useEffect, useId, useMemo, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { applyDensity, applyTheme, rootProps, typography, usePaperForm } from 'paper-ui'
import type {
  CatalogFixture,
  CatalogFixtureViewport,
  FixtureDensity,
  FixtureTheme,
} from 'paper-ui/catalog'
import type { PaperDensity, PaperTheme } from '../app/displayPreferences'
import { shellStyles } from '../styles/shell.styles'
import { descendantStyles } from '../descendants.styles'
import { size } from 'paper-ui/tokens.stylex'

/** The frame takes the chosen viewport's size, capped so a tall viewport doesn't push the page away. */
const frame = stylex.create({
  base: { borderWidth: 0, boxSizing: 'content-box' },
  size: (inlineSize: string, blockSize: string) => ({ blockSize, inlineSize }),
})

const FALLBACK_VIEWPORTS = [
  { id: 'phone', label: 'Phone', width: 360, height: 720 },
  { id: 'tablet', label: 'Tablet', width: 768, height: 800 },
  { id: 'desktop', label: 'Desktop', width: 1280, height: 800 },
] as const satisfies readonly CatalogFixtureViewport[]

const THEMES: readonly FixtureTheme[] = ['light', 'dark']
const DENSITIES: readonly FixtureDensity[] = ['comfortable', 'compact']

interface CanvasFields {
  fixtureId: string
  theme: PaperTheme
  density: PaperDensity
  viewportId: string
}

interface ExampleCanvasProps {
  readonly fixture?: CatalogFixture
  readonly fixtures?: readonly CatalogFixture[]
  readonly onFixtureChange?: (fixture: CatalogFixture) => void
}

function preferredTheme(themes: readonly FixtureTheme[]): PaperTheme {
  return themes.includes('light') ? 'light' : (themes[0] ?? 'light')
}

function preferredDensity(densities: readonly FixtureDensity[]): PaperDensity {
  return densities.includes('comfortable')
    ? 'comfortable'
    : (densities[0] ?? 'comfortable')
}

function titleCase(value: string): string {
  return `${value.charAt(0).toUpperCase()}${value.slice(1)}`
}

function PreviewSpecimen() {
  return (
    <article {...stylex.props(shellStyles.previewSpecimen)}>
      <p {...stylex.props(typography.metadata)}>Responsive Paper specimen</p>
      <h3 {...stylex.props(typography.section, descendantStyles.previewSpecimenU_uH3)}>Read more, one page at a time.</h3>
      <p>
        The canvas is a separate document. Change its width to exercise real
        media queries without shrinking the catalogue itself.
      </p>
      <div {...stylex.props(shellStyles.previewBreakpoint)} aria-live="polite">
        <span {...stylex.props(shellStyles.previewBreakpoint__narrow)}>Narrow layout active</span>
        <span {...stylex.props(shellStyles.previewBreakpoint__wide)}>Wide layout active</span>
      </div>
    </article>
  )
}

function copyPaperStyles(target: Document) {
  const sources = document.querySelectorAll('style, link[rel="stylesheet"]')
  for (const source of sources) {
    target.head.append(source.cloneNode(true))
  }
}

function applyPreferences(
  target: Document,
  theme: PaperTheme,
  density: PaperDensity,
) {
  target.documentElement.className = rootProps.html().className ?? ''
  target.body.className = rootProps.body().className ?? ''
  applyTheme(theme, target.documentElement)
  applyDensity(density, target.documentElement)
}

export function ExampleCanvas({
  fixture: requestedFixture,
  fixtures,
  onFixtureChange,
}: ExampleCanvasProps) {
  const initialFixture = requestedFixture ?? fixtures?.[0]
  const methods = usePaperForm({ defaultValues: {
    fixtureId: initialFixture?.id ?? '', theme: preferredTheme(initialFixture?.themes ?? THEMES),
    density: preferredDensity(initialFixture?.densities ?? DENSITIES), viewportId: 'fit',
  } as CanvasFields })
  const settingsTitleId = useId()
  const fixtureId = useStore(methods.store, state => state.values['fixtureId'])
  const availableFixtures = useMemo(
    () =>
      fixtures?.length
        ? fixtures
        : requestedFixture
          ? [requestedFixture]
          : [],
    [fixtures, requestedFixture],
  )
  const fixture =
    availableFixtures.find((candidate) => candidate.id === fixtureId) ??
    requestedFixture ??
    availableFixtures[0]
  const themes = fixture?.themes.length ? fixture.themes : THEMES
  const densities = fixture?.densities.length ? fixture.densities : DENSITIES
  const viewports = fixture?.viewports.length
    ? fixture.viewports
    : FALLBACK_VIEWPORTS
  const themeCollectionKey = themes.join(',')
  const densityCollectionKey = densities.join(',')
  const viewportCollectionKey = viewports
    .map(({ id, width, height }) => `${id}:${width}:${height}`)
    .join(',')
  const theme = useStore(methods.store, state => state.values['theme'])
  const density = useStore(methods.store, state => state.values['density'])
  const viewportId = useStore(methods.store, state => state.values['viewportId'])
  const [previewRoot, setPreviewRoot] = useState<HTMLElement | null>(null)
  const frameDocumentRef = useRef<Document | null>(null)
  const stageRef = useRef<HTMLDivElement | null>(null)
  const [availableWidth, setAvailableWidth] = useState(360)
  const [previewDestination, setPreviewDestination] = useState<{ fixtureId: string; href: string } | null>(null)
  const isFit = viewportId === 'fit'
  const viewport = isFit
    ? { id: 'fit', label: 'Fit', width: availableWidth, height: 800 }
    : viewports.find((candidate) => candidate.id === viewportId) ?? viewports[0]

  useEffect(() => {
    const stage = stageRef.current
    if (!stage || typeof ResizeObserver === 'undefined') return
    const observer = new ResizeObserver(([entry]) => {
      if (entry && entry.contentRect.width > 0) {
        setAvailableWidth(Math.floor(entry.contentRect.width))
      }
    })
    observer.observe(stage)
    return () => observer.disconnect()
  }, [])

  useEffect(() => {
    const selectedFixtureId = methods.getFieldValue('fixtureId')
    if (
      availableFixtures.length &&
      !availableFixtures.some((candidate) => candidate.id === selectedFixtureId)
    ) {
      methods.setFieldValue('fixtureId', availableFixtures[0].id)
    }
  }, [availableFixtures, methods])

  useEffect(() => {
    if (fixture) onFixtureChange?.(fixture)
  }, [fixture, onFixtureChange])

  useEffect(() => {
    const current = methods.state.values
    if (!themes.includes(current.theme)) {
      methods.setFieldValue('theme', preferredTheme(themes))
    }
    if (!densities.includes(current.density)) {
      methods.setFieldValue('density', preferredDensity(densities))
    }
    if (current.viewportId !== 'fit' && !viewports.some((candidate) => candidate.id === current.viewportId)) {
      methods.setFieldValue('viewportId', 'fit')
    }
  }, [
    densities,
    densityCollectionKey,
    methods,
    themes,
    themeCollectionKey,
    viewportCollectionKey,
    viewports,
  ])

  useEffect(() => {
    const frameDocument = frameDocumentRef.current
    if (!frameDocument) return
    applyPreferences(frameDocument, theme, density)
  }, [density, theme])

  if (!viewport) return null

  return (
    <methods.AppForm><section {...stylex.props(shellStyles.exampleCanvas)} aria-label="Isolated component preview">
      <div
        {...stylex.props(shellStyles.canvasControls)}
        role="group"
        aria-labelledby={settingsTitleId}
        onChange={(event) => {
          if ((event.target as HTMLSelectElement).name === 'fixtureId') setPreviewDestination(null)
        }}
      >
        <h3
          id={settingsTitleId}
          {...stylex.props(shellStyles.canvasControls__title, typography.label)}
        >
          Preview settings
        </h3>
        {availableFixtures.length > 1 ? (
          <div {...stylex.props(shellStyles.canvasControls__fixture)}>
            <methods.AppField name="fixtureId">{field => <field.Select label="Fixture" options={availableFixtures.map((candidate) => ({
                value: candidate.id,
                label: candidate.name,
              }))} />}</methods.AppField>
          </div>
        ) : null}
        <div {...stylex.props(shellStyles.canvasControls__theme)}>
          <methods.AppField name="theme">{field => <field.Select label="Theme" options={themes.map((candidate) => ({
              value: candidate,
              label: titleCase(candidate),
            }))} />}</methods.AppField>
        </div>
        <div {...stylex.props(shellStyles.canvasControls__density)}>
          <methods.AppField name="density">{field => <field.Select label="Density" options={densities.map((candidate) => ({
              value: candidate,
              label: titleCase(candidate),
            }))} />}</methods.AppField>
        </div>
        <div {...stylex.props(shellStyles.canvasControls__viewport)}>
          <methods.AppField name="viewportId">{field => <field.RadioSelect label="Viewport" variant="segmented" options={[{ value: 'fit', label: 'Fit' }, ...viewports.map((option) => ({
              value: option.id,
              label: option.label,
            }))]} />}</methods.AppField>
        </div>
      </div>
      <p {...stylex.props(shellStyles.canvasStatus, typography.metadata)} aria-live="polite">
        {viewport.label} · {viewport.width} px · {titleCase(theme)} ·{' '}
        {titleCase(density)}
      </p>
      {previewDestination && previewDestination.fixtureId === fixture?.id && (
        <p {...stylex.props(shellStyles.canvasStatus, typography.metadata)} role="status">
          Preview destination: {previewDestination.href}. Navigation stays in this example.
        </p>
      )}
      <div {...stylex.props(shellStyles.exampleCanvas__stage)} ref={stageRef}>
        <iframe {...stylex.props(descendantStyles.exampleCanvasU_uIframe, frame.base, frame.size(isFit ? '100%' : `${viewport.width}px`, `min(${viewport.height}px, ${size[352]})`))}
          title="Paper responsive component preview"
          width={isFit ? '100%' : viewport.width}
          height={viewport.height}
          data-fixture-id={fixture?.id}
          srcDoc={'<!doctype html><html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head><body><div id="paper-preview-root"></div></body></html>'}
          data-preview-width={viewport.width}
          data-preview-height={viewport.height}
          onLoad={(event) => {
            const targetDocument = event.currentTarget.contentDocument
            if (!targetDocument) return
            copyPaperStyles(targetDocument)
            applyPreferences(targetDocument, theme, density)
            frameDocumentRef.current = targetDocument
            setPreviewRoot(targetDocument.getElementById('paper-preview-root'))
          }}
        />
      </div>
      {previewRoot
        ? createPortal(
            fixture ? (
              <div key={fixture.id} {...stylex.props(shellStyles.paperFixtureStage)} onClick={(event) => {
                const link = (event.target as Element).closest?.('a[href]')
                if (!link || event.defaultPrevented) return
                event.preventDefault()
                const href = link.getAttribute('href')!
                setPreviewDestination({ fixtureId: fixture.id, href })
                if (href.startsWith('#')) {
                  try {
                    link.ownerDocument.getElementById(decodeURIComponent(href.slice(1)))?.scrollIntoView?.({ block: 'nearest' })
                  } catch { /* A malformed example fragment has no scroll target. */ }
                }
              }}>{fixture.render()}</div>
            ) : (
              <PreviewSpecimen />
            ),
            previewRoot,
          )
        : null}
    </section></methods.AppForm>
  )
}
