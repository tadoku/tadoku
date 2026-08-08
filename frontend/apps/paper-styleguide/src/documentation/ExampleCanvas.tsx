import { useEffect, useMemo, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { FormProvider, useForm } from 'react-hook-form'
import { Button, Select } from 'paper-ui'
import type {
  CatalogFixture,
  FixtureDensity,
  FixtureTheme,
} from 'paper-ui/catalog'
import type { PaperDensity, PaperTheme } from '../app/displayPreferences'

const VIEWPORTS = [
  { id: 'phone', name: 'Phone', width: 360 },
  { id: 'tablet', name: 'Tablet', width: 768 },
  { id: 'desktop', name: 'Desktop', width: 1280 },
] as const

const THEMES: readonly FixtureTheme[] = ['light', 'dark']
const DENSITIES: readonly FixtureDensity[] = ['comfortable', 'compact']

interface CanvasFields {
  theme: PaperTheme
  density: PaperDensity
}

function initialViewportId(): (typeof VIEWPORTS)[number]['id'] {
  return typeof window.matchMedia === 'function' &&
    window.matchMedia('(max-width: 48rem)').matches
    ? 'phone'
    : 'desktop'
}

function preferredTheme(themes: readonly FixtureTheme[]): PaperTheme {
  return themes.includes('light') ? 'light' : (themes[0] ?? 'light')
}

function preferredDensity(densities: readonly FixtureDensity[]): PaperDensity {
  return densities.includes('comfortable')
    ? 'comfortable'
    : (densities[0] ?? 'comfortable')
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

export function ExampleCanvas({ fixture }: { fixture?: CatalogFixture }) {
  const themes = fixture?.themes.length ? fixture.themes : THEMES
  const densities = fixture?.densities.length ? fixture.densities : DENSITIES
  const themeCollectionKey = themes.join(',')
  const densityCollectionKey = densities.join(',')
  const methods = useForm<CanvasFields>({
    defaultValues: {
      theme: preferredTheme(themes),
      density: preferredDensity(densities),
    },
  })
  const theme = methods.watch('theme')
  const density = methods.watch('density')
  const [viewportId, setViewportId] = useState<(typeof VIEWPORTS)[number]['id']>(
    initialViewportId,
  )
  const [previewRoot, setPreviewRoot] = useState<HTMLElement | null>(null)
  const frameDocumentRef = useRef<Document | null>(null)
  const viewport = useMemo(
    () => VIEWPORTS.find((candidate) => candidate.id === viewportId)!,
    [viewportId],
  )

  useEffect(() => {
    const current = methods.getValues()
    const nextTheme = themes.includes(current.theme)
      ? current.theme
      : preferredTheme(themes)
    const nextDensity = densities.includes(current.density)
      ? current.density
      : preferredDensity(densities)

    if (nextTheme !== current.theme || nextDensity !== current.density) {
      methods.reset({ theme: nextTheme, density: nextDensity })
    }
  }, [densities, densityCollectionKey, methods, themes, themeCollectionKey])

  useEffect(() => {
    const frameDocument = frameDocumentRef.current
    if (!frameDocument) return
    applyPreferences(frameDocument, theme, density)
  }, [density, theme])

  return (
    <FormProvider {...methods}>
      <section className="example-canvas" aria-labelledby="example-canvas-title">
        <div className="example-canvas__heading">
          <div>
            <p className="eyebrow">Isolated preview</p>
            <h2 id="example-canvas-title" className="paper-type-section">
              Real viewport canvas
            </h2>
          </div>
          <div className="canvas-controls">
            <Select
              name="theme"
              label="Theme"
              aria-label="Preview theme"
              options={themes.map((candidate) => ({
                value: candidate,
                label: candidate === 'dark' ? 'Dark' : 'Light',
              }))}
            />
            <Select
              name="density"
              label="Density"
              aria-label="Preview density"
              options={densities.map((candidate) => ({
                value: candidate,
                label: candidate === 'compact' ? 'Compact' : 'Comfortable',
              }))}
            />
            <fieldset>
              <legend>Viewport</legend>
              {VIEWPORTS.map((option) => {
                const selected = viewportId === option.id
                return (
                  <Button
                    key={option.id}
                    variant={selected ? 'default' : 'outline'}
                    className="canvas-viewport-button"
                    aria-pressed={selected}
                    aria-label={`${option.name}, ${option.width} pixels`}
                    onClick={() => setViewportId(option.id)}
                  >
                    {option.name}
                  </Button>
                )
              })}
            </fieldset>
          </div>
        </div>
        <p className="canvas-status paper-type-metadata" aria-live="polite">
          {viewport.name} · {viewport.width}px · {theme} · {density}
        </p>
        <div className="example-canvas__stage">
          <iframe
            title="Paper responsive component preview"
            width={viewport.width}
            data-fixture-id={fixture?.id}
            srcDoc={'<!doctype html><html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head><body><div id="paper-preview-root"></div></body></html>'}
            data-preview-width={viewport.width}
            style={{
              inlineSize: `${viewport.width}px`,
              boxSizing: 'content-box',
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
                <div className="paper-fixture-stage">{fixture.render()}</div>
              ) : (
                <PreviewSpecimen />
              ),
              previewRoot,
            )
          : null}
      </section>
    </FormProvider>
  )
}
