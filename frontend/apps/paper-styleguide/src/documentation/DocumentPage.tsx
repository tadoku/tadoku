import {
  catalogRegistry,
  type CatalogDocument,
  type CatalogDocumentationSection,
  type ComponentPageSectionKey,
} from 'paper-ui/catalog'
import { ComponentWorkbench } from './ComponentWorkbench'
import { FoundationSpecimen } from './FoundationSpecimen'
import { GovernanceGuide } from './FoundationGuide'
import { type OutlineItem, TableOfContents } from './TableOfContents'
import { accentRail, cn, focusRing, typography } from 'paper-ui'

const documentSection = 'document-section max-w-docs-section scroll-mt-24 [&+&]:mt-12 [&>h2]:mb-5 [&>h2]:mt-0 [&>h3]:mb-2.5 [&>h3]:mt-7 [&>h2+h3]:mt-0 [&>p]:mb-4 [&>p]:mt-0'
const wideDocumentSection = cn(documentSection, 'max-w-none')

const FOUNDATION_OUTLINE: readonly OutlineItem[] = [
  { id: 'overview', label: 'Overview' },
  { id: 'specimen', label: 'Examples and usage' },
  { id: 'guidance', label: 'Guidance' },
  { id: 'accessibility', label: 'Accessibility' },
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

function LabeledList({
  heading,
  items,
}: {
  heading: string
  items: readonly string[]
}) {
  if (!items.length) return null
  return (
    <div className="document-list-group [&+&]:mt-7 [&_h3]:mb-2.5 [&_h3]:mt-0 [&_ul]:my-0 [&_ul]:ps-5 [&_li+li]:mt-1.8">
      <h3>{heading}</h3>
      {items.length > 0 ? (
        <ul>
          {items.map((item) => <li key={item}>{item}</li>)}
        </ul>
      ) : (
        <p>None registered.</p>
      )}
    </div>
  )
}

function sourceHref(sourcePath: string): string {
  const repositoryPath = sourcePath.startsWith('docs/')
    ? sourcePath
    : `frontend/packages/paper-ui/${sourcePath}`
  return `https://github.com/tadoku/tadoku/blob/main/${repositoryPath}`
}

function Metadata({ document }: { document: CatalogDocument }) {
  return (
    <dl className="[&>div]:grid [&>div]:grid-cols-metadata-list [&>div]:gap-4 [&>div]:border-b [&>div]:border-rule-subtle [&>div]:py-3 max-tablet:[&>div]:grid-cols-1 max-tablet:[&>div]:gap-0 [&_dt]:text-metadata [&_dt]:text-muted [&_dd]:mb-0 [&_dd]:ms-0 [&_dd]:mt-0.8">
      <div>
        <dt>Canonical route</dt>
        <dd><code>{document.route}</code></dd>
      </div>
      <div>
        <dt>Source</dt>
        <dd>
          <a className={cn(focusRing(), 'text-link')} href={sourceHref(document.sourcePath)}>
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
    <header className={cn(accentRail(), 'mb-16 ps-4 [&_h1]:mb-4 [&_h1]:mt-1.2')}>
      <p className="m-0 text-metadata font-semibold tracking-caps text-link uppercase">{document.kind} · {document.category}</p>
      <h1 className={typography({ role: 'page' })}>{document.name}</h1>
      <p className="text-component leading-165 text-muted">{document.summary}</p>
      <dl className="mt-8 grid grid-cols-thirds gap-4 border-t border-rule-subtle pt-4 max-tablet:grid-cols-pair [&_dt]:text-metadata [&_dt]:text-muted [&_dd]:mb-0 [&_dd]:ms-0 [&_dd]:mt-0.8">
        <div><dt>Status</dt><dd>{document.lifecycle}</dd></div>
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
    <section id={id} className={documentSection}>
      <h2 className={typography({ role: 'section' })}>{section.heading}</h2>
      {section.content.map((paragraph) => <p key={paragraph}>{paragraph}</p>)}
    </section>
  )
}

const COMPONENT_PAGE_HEADINGS: Readonly<Record<ComponentPageSectionKey, string>> = {
  usage: 'Usage',
  examples: 'Examples',
  variantsAndStates: 'Variants and states',
  behavior: 'Behavior',
  contentGuidance: 'Content guidance',
  accessibility: 'Accessibility',
}

function SectionParagraphs({ section }: { section?: CatalogDocumentationSection }) {
  return section ? (
    <>{section.content.map((paragraph) => <p key={paragraph}>{paragraph}</p>)}</>
  ) : null
}

function UsageSection({ document }: { document: CatalogDocument }) {
  const sections = document.sections?.required
  return (
    <section id="usage" className={documentSection}>
      <h2 className={typography({ role: 'section' })}>Usage</h2>
      <h3>Use when</h3>
      <SectionParagraphs section={sections?.whenToUse} />
      <h3>Avoid when</h3>
      <SectionParagraphs section={sections?.whenNotToUse} />
      <h3>Choosing between components</h3>
      <SectionParagraphs section={sections?.choosingBetween} />
      <h3>How it is composed</h3>
      <SectionParagraphs section={sections?.anatomy} />
      <LabeledList heading="Common mistakes" items={document.guidance.commonMistakes} />
    </section>
  )
}

function ExamplesSection({
  document,
  fixtures,
}: {
  document: CatalogDocument
  fixtures: Parameters<typeof ComponentWorkbench>[0]['fixtures']
}) {
  return (
    <section id="examples" className={wideDocumentSection}>
      <h2 className={typography({ role: 'section' })}>Examples</h2>
      <SectionParagraphs section={document.sections?.required.recommendedExample} />
      <SectionParagraphs section={document.sections?.required.implementation} />
      <ComponentWorkbench document={document} fixtures={fixtures} />
    </section>
  )
}

function VariantsAndStatesSection({ document }: { document: CatalogDocument }) {
  const sections = document.sections?.required
  return (
    <section id="variantsAndStates" className={documentSection}>
      <h2 className={typography({ role: 'section' })}>Variants and states</h2>
      <h3>Variants</h3>
      <SectionParagraphs section={sections?.variants} />
      <h3>States and adaptation</h3>
      <SectionParagraphs section={sections?.statesAndAdaptation} />
    </section>
  )
}

function PublicComponentSection({
  keyName,
  document,
  fixtures,
}: {
  keyName: ComponentPageSectionKey
  document: CatalogDocument
  fixtures: Parameters<typeof ComponentWorkbench>[0]['fixtures']
}) {
  if (keyName === 'usage') return <UsageSection document={document} />
  if (keyName === 'examples') {
    return <ExamplesSection document={document} fixtures={fixtures} />
  }
  if (keyName === 'variantsAndStates') {
    return <VariantsAndStatesSection document={document} />
  }

  const sourceKey = keyName === 'contentGuidance' ? 'contentGuidance' : keyName
  const section = document.sections?.required[sourceKey]
  return section ? (
    <ComponentSection id={keyName} section={section} />
  ) : null
}

function ComponentDocumentPage({ document }: { document: CatalogDocument }) {
  const fixtures = catalogRegistry.fixtures.filter((fixture) =>
    document.fixtureIds.includes(fixture.id),
  )
  const pageSections = document.sections?.pageSections ?? []
  const outline: OutlineItem[] = pageSections.map((key) => ({
    id: key,
    label: COMPONENT_PAGE_HEADINGS[key],
  }))

  return (
    <div className="grid grid-cols-document justify-center gap-docs-y max-lg:grid-cols-document-narrow">
      <article className="document-page">
        <Hero document={document} />
        {pageSections.map((key) => (
          <PublicComponentSection
            key={key}
            keyName={key}
            document={document}
            fixtures={fixtures}
          />
        ))}
      </article>
      <TableOfContents items={outline} />
    </div>
  )
}

function FoundationDocumentPage({ document }: { document: CatalogDocument }) {
  return (
    <div className="grid grid-cols-document justify-center gap-docs-y max-lg:grid-cols-document-narrow">
      <article className="document-page">
        <div id="overview"><Hero document={document} /></div>
        <div id="specimen" className={wideDocumentSection}>
          <h2 className={typography({ role: 'section' })}>Examples and usage</h2>
          <FoundationSpecimen document={document} />
        </div>
        <section
          id="guidance"
          className={documentSection}
          aria-labelledby="foundation-guidance-heading"
        >
          <h2 id="foundation-guidance-heading" className={typography({ role: 'section' })}>
            Guidance
          </h2>
          <div className="grid grid-cols-1 gap-7 [&>.document-list-group]:mt-0">
            <LabeledList heading="When to use" items={document.guidance.whenToUse} />
            <LabeledList heading="Avoid when" items={document.guidance.whenNotToUse} />
            <LabeledList heading="Guidance" items={document.guidance.content} />
            <LabeledList heading="Common mistakes" items={document.guidance.commonMistakes} />
          </div>
        </section>

        <section id="accessibility" className={documentSection}>
          <h2 className={typography({ role: 'section' })}>Accessibility</h2>
          <div className="grid grid-cols-1 gap-7 [&>.document-list-group]:mt-0">
            <LabeledList heading="Requirements" items={document.accessibility.requirements} />
            <LabeledList heading="Keyboard" items={document.accessibility.keyboard} />
            <LabeledList heading="Known constraints" items={document.accessibility.knownConstraints} />
          </div>
        </section>
        <section id="contract" className={documentSection}>
          <h2 className={typography({ role: 'section' })}>Public contract</h2>
          <LabeledList heading="Defaults" items={document.api.defaults} />
          <LabeledList heading="Not public API" items={document.api.invalidCombinations} />
        </section>
        <section id="metadata" className={documentSection}>
          <h2 className={typography({ role: 'section' })}>Metadata</h2>
          <Metadata document={document} />
        </section>
      </article>
      <TableOfContents items={FOUNDATION_OUTLINE} />
    </div>
  )
}

function GeneralDocumentPage({ document }: { document: CatalogDocument }) {
  const fixture = catalogRegistry.fixtures.find((candidate) =>
    document.fixtureIds.includes(candidate.id),
  )
  const outline = FOUNDATION_OUTLINE.filter((item) =>
    item.id !== 'specimen',
  )
  const generalOutline = fixture
    ? [
        ...outline.slice(0, 3),
        { id: 'preview', label: 'Preview canvas' },
        ...outline.slice(3),
      ]
    : outline
  const outlineWithUsage = [...generalOutline, { id: 'usage', label: 'Usage' }]

  return (
    <div className="grid grid-cols-document justify-center gap-docs-y max-lg:grid-cols-document-narrow">
      <article className="document-page">
        <div id="overview"><Hero document={document} /></div>
        <section id="guidance" className={documentSection}>
          <h2 className={typography({ role: 'section' })}>Guidance</h2>
          <SectionCopy value={document.guidance} fallback="No additional guidance is registered for this document." />
        </section>
        <section id="accessibility" className={documentSection}>
          <h2 className={typography({ role: 'section' })}>Accessibility</h2>
          <SectionCopy value={document.accessibility} fallback="Accessibility requirements are tracked in the catalogue registry." />
        </section>
        {fixture ? (
          <div id="preview" className={wideDocumentSection}>
            <ComponentWorkbench document={document} fixtures={[fixture]} />
          </div>
        ) : null}
        <GovernanceGuide document={document} />
        <section id="contract" className={documentSection}>
          <h2 className={typography({ role: 'section' })}>Public contract</h2>
          <SectionCopy value={document.api} fallback="This foundation does not expose a component API." />
        </section>
        <section id="metadata" className={documentSection}>
          <h2 className={typography({ role: 'section' })}>Metadata</h2>
          <Metadata document={document} />
        </section>
      </article>
      <TableOfContents items={outlineWithUsage} />
    </div>
  )
}

export function DocumentPage({ document }: { document: CatalogDocument }) {
  if (document.kind === 'component') {
    return <ComponentDocumentPage document={document} />
  }
  if (document.kind === 'foundation') {
    return <FoundationDocumentPage document={document} />
  }
  return <GeneralDocumentPage document={document} />
}
