import { useEffect, useMemo, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import type { CatalogFixture } from 'paper-ui/catalog'
import type { PaperDensity, PaperTheme } from '../app/displayPreferences'

const VIEWPORTS = [
  { id: 'phone', name: 'Phone', width: 360 },
  { id: 'tablet', name: 'Tablet', width: 768 },
  { id: 'desktop', name: 'Desktop', width: 1280 },
] as const

function initialViewportId(): (typeof VIEWPORTS)[number]['id'] {
  return typeof window.matchMedia === 'function' &&
    window.matchMedia('(max-width: 48rem)').matches
    ? 'phone'
    : 'desktop'
}

function PreviewSpecimen() {
  return (
    <article className="preview-specimen paper-surface paper-accent-rail">
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

export function ExampleCanvas({ fixture }: { fixture?: CatalogFixture }) {
  const [theme, setTheme] = useState<PaperTheme>('light')
  const [density, setDensity] = useState<PaperDensity>('comfortable')
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
    const frameDocument = frameDocumentRef.current
    if (!frameDocument) return
    frameDocument.documentElement.dataset.theme = theme
    frameDocument.documentElement.dataset.density = density
  }, [density, theme])

  return (
    <section className="example-canvas" aria-labelledby="example-canvas-title">
      <div className="example-canvas__heading">
        <div>
          <p className="eyebrow">Isolated preview</p>
          <h2 id="example-canvas-title" className="paper-type-section">
            Real viewport canvas
          </h2>
        </div>
        <div className="canvas-controls">
          <label>
            <span>Theme</span>
            <select
              aria-label="Preview theme"
              value={theme}
              onChange={(event) =>
                setTheme(event.target.value === 'dark' ? 'dark' : 'light')
              }
            >
              <option value="light">Light</option>
              <option value="dark">Dark</option>
            </select>
          </label>
          <label>
            <span>Density</span>
            <select
              aria-label="Preview density"
              value={density}
              onChange={(event) =>
                setDensity(
                  event.target.value === 'compact' ? 'compact' : 'comfortable',
                )
              }
            >
              <option value="comfortable">Comfortable</option>
              <option value="compact">Compact</option>
            </select>
          </label>
          <fieldset>
            <legend>Viewport</legend>
            {VIEWPORTS.map((option) => (
              <button
                key={option.id}
                type="button"
                className={`canvas-viewport-button paper-focus-ring${
                  viewportId === option.id ? ' canvas-viewport-button--active' : ''
                }`}
                aria-pressed={viewportId === option.id}
                aria-label={`${option.name}, ${option.width} pixels`}
                onClick={() => setViewportId(option.id)}
              >
                {option.name}
              </button>
            ))}
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
  )
}
