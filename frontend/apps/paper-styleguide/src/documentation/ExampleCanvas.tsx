import { useEffect, useId, useMemo, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { FormProvider, useForm, useFormContext } from 'react-hook-form'
import { RadioSelect, Select } from 'paper-ui'
import type {
  CatalogFixture,
  CatalogFixtureViewport,
  FixtureDensity,
  FixtureTheme,
} from 'paper-ui/catalog'
import type { PaperDensity, PaperTheme } from '../app/displayPreferences'

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
    <article className="preview-specimen">
      <p className="paper-type-metadata">Responsive Paper specimen</p>
      <h3 className="paper-type-section">Read more, one page at a time.</h3>
      <p>
        The canvas is a separate document. Change its width to exercise real
        media queries without shrinking the catalogue itself.
      </p>
      <div className="preview-breakpoint" aria-live="polite">
        <span className="preview-breakpoint__narrow">Narrow layout active</span>
        <span className="preview-breakpoint__wide">Wide layout active</span>
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
  target.documentElement.dataset.theme = theme
  target.documentElement.dataset.density = density
}

function CanvasContent({
  fixture: requestedFixture,
  fixtures,
  onFixtureChange,
}: ExampleCanvasProps) {
  const methods = useFormContext<CanvasFields>()
  const settingsTitleId = useId()
  const fixtureId = methods.watch('fixtureId')
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
  const theme = methods.watch('theme')
  const density = methods.watch('density')
  const viewportId = methods.watch('viewportId')
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
    const selectedFixtureId = methods.getValues('fixtureId')
    if (
      availableFixtures.length &&
      !availableFixtures.some((candidate) => candidate.id === selectedFixtureId)
    ) {
      methods.setValue('fixtureId', availableFixtures[0].id)
    }
  }, [availableFixtures, methods])

  useEffect(() => {
    if (fixture) onFixtureChange?.(fixture)
  }, [fixture, onFixtureChange])

  useEffect(() => {
    const current = methods.getValues()
    if (!themes.includes(current.theme)) {
      methods.setValue('theme', preferredTheme(themes))
    }
    if (!densities.includes(current.density)) {
      methods.setValue('density', preferredDensity(densities))
    }
    if (current.viewportId !== 'fit' && !viewports.some((candidate) => candidate.id === current.viewportId)) {
      methods.setValue('viewportId', 'fit')
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
    <section className="example-canvas" aria-label="Isolated component preview">
      <div
        className="canvas-controls"
        role="group"
        aria-labelledby={settingsTitleId}
        onChange={(event) => {
          if ((event.target as HTMLSelectElement).name === 'fixtureId') setPreviewDestination(null)
        }}
      >
        <h3
          id={settingsTitleId}
          className="canvas-controls__title paper-type-label"
        >
          Preview settings
        </h3>
        {availableFixtures.length > 1 ? (
          <div className="canvas-controls__fixture">
            <Select
              name="fixtureId"
              label="Fixture"
              options={availableFixtures.map((candidate) => ({
                value: candidate.id,
                label: candidate.name,
              }))}
            />
          </div>
        ) : null}
        <div className="canvas-controls__theme">
          <Select
            name="theme"
            label="Theme"
            options={themes.map((candidate) => ({
              value: candidate,
              label: titleCase(candidate),
            }))}
          />
        </div>
        <div className="canvas-controls__density">
          <Select
            name="density"
            label="Density"
            options={densities.map((candidate) => ({
              value: candidate,
              label: titleCase(candidate),
            }))}
          />
        </div>
        <div className="canvas-controls__viewport">
          <RadioSelect
            name="viewportId"
            label="Viewport"
            variant="segmented"
            options={[{ value: 'fit', label: 'Fit' }, ...viewports.map((option) => ({
              value: option.id,
              label: option.label,
            }))]}
          />
        </div>
      </div>
      <p className="canvas-status paper-type-metadata" aria-live="polite">
        {viewport.label} · {viewport.width} px · {titleCase(theme)} ·{' '}
        {titleCase(density)}
      </p>
      {previewDestination && previewDestination.fixtureId === fixture?.id && (
        <p className="canvas-status paper-type-metadata" role="status">
          Preview destination: {previewDestination.href}. Navigation stays in this example.
        </p>
      )}
      <div className="example-canvas__stage" ref={stageRef}>
        <iframe
          title="Paper responsive component preview"
          width={isFit ? '100%' : viewport.width}
          height={viewport.height}
          data-fixture-id={fixture?.id}
          srcDoc={'<!doctype html><html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head><body><div id="paper-preview-root"></div></body></html>'}
          data-preview-width={viewport.width}
          data-preview-height={viewport.height}
          style={{
            inlineSize: isFit ? '100%' : `${viewport.width}px`,
            blockSize: `min(${viewport.height}px, 22rem)`,
            boxSizing: 'content-box',
            border: 0,
          }}
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
              <div key={fixture.id} className="paper-fixture-stage" onClick={(event) => {
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
    </section>
  )
}

export function ExampleCanvas(props: ExampleCanvasProps) {
  const initialFixture = props.fixture ?? props.fixtures?.[0]
  const methods = useForm<CanvasFields>({
    defaultValues: {
      fixtureId: initialFixture?.id ?? '',
      theme: preferredTheme(initialFixture?.themes ?? THEMES),
      density: preferredDensity(initialFixture?.densities ?? DENSITIES),
      viewportId: 'fit',
    },
  })

  return (
    <FormProvider {...methods}>
      <CanvasContent {...props} />
    </FormProvider>
  )
}
