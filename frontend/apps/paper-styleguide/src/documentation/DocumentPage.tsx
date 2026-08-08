import type { CatalogDocument } from 'paper-ui/catalog'
import { ExampleCanvas } from './ExampleCanvas'
import { type OutlineItem, TableOfContents } from './TableOfContents'

const OUTLINE: readonly OutlineItem[] = [
  { id: 'overview', label: 'Overview' },
  { id: 'guidance', label: 'Guidance' },
  { id: 'accessibility', label: 'Accessibility' },
  { id: 'preview', label: 'Preview canvas' },
  { id: 'contract', label: 'Public contract' },
  { id: 'metadata', label: 'Metadata' },
]

function readable(value: unknown): string {
  if (typeof value === 'string') return value
  if (Array.isArray(value)) return value.map(readable).filter(Boolean).join(' ')
  if (value && typeof value === 'object') {
    return Object.values(value).map(readable).filter(Boolean).join(' ')
  }
  return ''
}

function SectionCopy({ value, fallback }: { value: unknown; fallback: string }) {
  const copy = readable(value)
  return <p>{copy || fallback}</p>
}

export function DocumentPage({ document }: { document: CatalogDocument }) {
  return (
    <div className="document-layout">
      <article className="document-page">
        <header id="overview" className="document-hero paper-accent-rail">
          <p className="eyebrow">
            {document.kind} · {document.category}
          </p>
          <h1 className="paper-type-page">{document.name}</h1>
          <p className="document-summary">{document.summary}</p>
          <dl className="metadata-strip">
            <div>
              <dt>Status</dt>
              <dd>{document.lifecycle}</dd>
            </div>
            <div>
              <dt>Owner</dt>
              <dd>{document.owner}</dd>
            </div>
            <div>
              <dt>Version</dt>
              <dd>{document.packageVersion}</dd>
            </div>
            <div>
              <dt>Reviewed</dt>
              <dd>{document.reviewDate}</dd>
            </div>
          </dl>
        </header>

        <section id="guidance" className="document-section">
          <h2 className="paper-type-section">Guidance</h2>
          <SectionCopy
            value={document.guidance}
            fallback="Detailed usage guidance will arrive with this catalogue slice."
          />
        </section>

        <section id="accessibility" className="document-section">
          <h2 className="paper-type-section">Accessibility</h2>
          <SectionCopy
            value={document.accessibility}
            fallback="Accessibility requirements are tracked in the catalogue registry."
          />
        </section>

        <div id="preview" className="document-section document-section--wide">
          <ExampleCanvas />
        </div>

        <section id="contract" className="document-section">
          <h2 className="paper-type-section">Public contract</h2>
          <SectionCopy
            value={document.api}
            fallback="This foundation does not expose a component API."
          />
        </section>

        <section id="metadata" className="document-section">
          <h2 className="paper-type-section">Metadata</h2>
          <dl className="metadata-list">
            <div>
              <dt>Canonical route</dt>
              <dd>
                <code>{document.route}</code>
              </dd>
            </div>
            <div>
              <dt>Source</dt>
              <dd>
                <code>{document.sourcePath}</code>
              </dd>
            </div>
            <div>
              <dt>Dependencies</dt>
              <dd>
                {[
                  ...document.dependencies.documents,
                  ...document.dependencies.packages,
                ].join(', ') || 'None'}
              </dd>
            </div>
          </dl>
        </section>
      </article>
      <TableOfContents items={OUTLINE} />
    </div>
  )
}
