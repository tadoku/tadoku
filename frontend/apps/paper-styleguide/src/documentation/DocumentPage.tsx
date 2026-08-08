import {
  REQUIRED_COMPONENT_SECTION_KEYS,
  catalogRegistry,
  type CatalogDocument,
  type CatalogDocumentationSection,
} from 'paper-ui/catalog'
import { Fragment } from 'react'
import { ComponentWorkbench } from './ComponentWorkbench'
import { ExampleCanvas } from './ExampleCanvas'
import { type OutlineItem, TableOfContents } from './TableOfContents'

const FOUNDATION_OUTLINE: readonly OutlineItem[] = [
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

function sourceHref(sourcePath: string): string {
  const repositoryPath = sourcePath.startsWith('docs/')
    ? sourcePath
    : `frontend/packages/paper-ui/${sourcePath}`
  return `https://github.com/tadoku/tadoku/blob/main/${repositoryPath}`
}

function Metadata({ document }: { document: CatalogDocument }) {
  return (
    <dl className="metadata-list">
      <div>
        <dt>Canonical route</dt>
        <dd><code>{document.route}</code></dd>
      </div>
      <div>
        <dt>Source</dt>
        <dd>
          <a className="text-link paper-focus-ring" href={sourceHref(document.sourcePath)}>
            <code>{document.sourcePath}</code>
          </a>
        </dd>
      </div>
      <div>
        <dt>Dependencies</dt>
        <dd>{[...document.dependencies.documents, ...document.dependencies.packages].join(', ') || 'None'}</dd>
      </div>
    </dl>
  )
}

function Hero({ document }: { document: CatalogDocument }) {
  return (
    <header className="document-hero paper-accent-rail">
      <p className="eyebrow">{document.kind} · {document.category}</p>
      <h1 className="paper-type-page">{document.name}</h1>
      <p className="document-summary">{document.summary}</p>
      <dl className="metadata-strip">
        <div><dt>Status</dt><dd>{document.lifecycle}</dd></div>
        <div><dt>Owner</dt><dd>{document.owner}</dd></div>
        <div><dt>Version</dt><dd>{document.packageVersion}</dd></div>
        <div><dt>Reviewed</dt><dd>{document.reviewDate}</dd></div>
      </dl>
    </header>
  )
}

function ComponentSection({
  id,
  section,
}: {
  id: string
  section: CatalogDocumentationSection
}) {
  return (
    <section id={id} className="document-section">
      <h2 className="paper-type-section">{section.heading}</h2>
      {section.content.map((paragraph) => <p key={paragraph}>{paragraph}</p>)}
    </section>
  )
}

function ComponentDocumentPage({ document }: { document: CatalogDocument }) {
  const fixtures = catalogRegistry.fixtures.filter((fixture) =>
    document.fixtureIds.includes(fixture.id),
  )
  const sections = document.sections?.required ?? {}
  const outline: OutlineItem[] = REQUIRED_COMPONENT_SECTION_KEYS.flatMap((key) => {
    const section = sections[key]
    return section ? [{ id: key, label: section.heading }] : []
  })
  outline.push({ id: 'metadata', label: 'Metadata' })

  return (
    <div className="document-layout">
      <article className="document-page">
        <Hero document={document} />
        {REQUIRED_COMPONENT_SECTION_KEYS.map((key) => {
          const section = sections[key]
          if (!section) return null
          return (
            <Fragment key={key}>
              <ComponentSection id={key} section={section} />
              {key === 'recommendedExample' ? (
                <div className="document-section--wide">
                  <ComponentWorkbench document={document} fixtures={fixtures} />
                </div>
              ) : null}
            </Fragment>
          )
        })}
        <section id="metadata" className="document-section">
          <h2 className="paper-type-section">Metadata</h2>
          <Metadata document={document} />
        </section>
      </article>
      <TableOfContents items={outline} />
    </div>
  )
}

function GeneralDocumentPage({ document }: { document: CatalogDocument }) {
  const fixture = catalogRegistry.fixtures.find((candidate) =>
    document.fixtureIds.includes(candidate.id),
  )

  return (
    <div className="document-layout">
      <article className="document-page">
        <div id="overview"><Hero document={document} /></div>
        <section id="guidance" className="document-section">
          <h2 className="paper-type-section">Guidance</h2>
          <SectionCopy value={document.guidance} fallback="No additional guidance is registered for this document." />
        </section>
        <section id="accessibility" className="document-section">
          <h2 className="paper-type-section">Accessibility</h2>
          <SectionCopy value={document.accessibility} fallback="Accessibility requirements are tracked in the catalogue registry." />
        </section>
        <div id="preview" className="document-section document-section--wide">
          <ExampleCanvas fixture={fixture} />
        </div>
        <section id="contract" className="document-section">
          <h2 className="paper-type-section">Public contract</h2>
          <SectionCopy value={document.api} fallback="This foundation does not expose a component API." />
        </section>
        <section id="metadata" className="document-section">
          <h2 className="paper-type-section">Metadata</h2>
          <Metadata document={document} />
        </section>
      </article>
      <TableOfContents items={FOUNDATION_OUTLINE} />
    </div>
  )
}

export function DocumentPage({ document }: { document: CatalogDocument }) {
  return document.kind === 'component'
    ? <ComponentDocumentPage document={document} />
    : <GeneralDocumentPage document={document} />
}
