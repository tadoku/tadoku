import * as stylex from '@stylexjs/stylex'
import { a11y, accentRail, linkStyles, typography } from 'paper-ui'
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
import { shellStyles } from '../styles/shell.styles'
import { descendantStyles } from '../descendants.styles'

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
    <div {...stylex.props(shellStyles.documentListGroup, descendantStyles.documentListGroup)}>
      <h3 {...stylex.props(descendantStyles.documentListGroupU_uH3)}>{heading}</h3>
      {items.length > 0 ? (
        <ul {...stylex.props(descendantStyles.documentListGroupU_uUl)}>
          {items.map((item) => <li {...stylex.props(descendantStyles.documentListGroupU_uLi)} key={item}>{item}</li>)}
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
    <dl {...stylex.props(shellStyles.metadataList)}>
      <div {...stylex.props(descendantStyles.metadataListU_uDiv)}>
        <dt {...stylex.props(descendantStyles.metadataListU_uDt)}>Canonical route</dt>
        <dd {...stylex.props(descendantStyles.metadataListU_uDd)}><code>{document.route}</code></dd>
      </div>
      <div {...stylex.props(descendantStyles.metadataListU_uDiv)}>
        <dt {...stylex.props(descendantStyles.metadataListU_uDt)}>Source</dt>
        <dd {...stylex.props(descendantStyles.metadataListU_uDd)}>
          <a {...stylex.props(linkStyles.base, a11y.focusRing)} href={sourceHref(document.sourcePath)}>
            <code>{document.sourcePath}</code>
          </a>
        </dd>
      </div>
      <div {...stylex.props(descendantStyles.metadataListU_uDiv)}>
        <dt {...stylex.props(descendantStyles.metadataListU_uDt)}>Dependencies</dt>
        <dd {...stylex.props(descendantStyles.metadataListU_uDd)}>{[...document.dependencies.documents, ...document.dependencies.packages].join(', ') || 'None'}</dd>
      </div>
    </dl>
  )
}

function Hero({ document }: { document: CatalogDocument }) {
  return (
    <header {...stylex.props(shellStyles.documentHero, accentRail.base)}>
      <p {...stylex.props(shellStyles.eyebrow)}>{document.kind} · {document.category}</p>
      <h1 {...stylex.props(typography.page, descendantStyles.documentHeroU_uH1)}>{document.name}</h1>
      <p {...stylex.props(shellStyles.documentSummary)}>{document.summary}</p>
      <dl {...stylex.props(shellStyles.metadataStrip)}>
        <div><dt {...stylex.props(descendantStyles.metadataStripU_uDt)}>Status</dt><dd {...stylex.props(descendantStyles.metadataStripU_uDd)}>{document.lifecycle}</dd></div>
        <div><dt {...stylex.props(descendantStyles.metadataStripU_uDt)}>Version</dt><dd {...stylex.props(descendantStyles.metadataStripU_uDd)}>{document.packageVersion}</dd></div>
        <div><dt {...stylex.props(descendantStyles.metadataStripU_uDt)}>Reviewed</dt><dd {...stylex.props(descendantStyles.metadataStripU_uDd)}>{document.reviewDate}</dd></div>
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
    <section id={id} {...stylex.props(shellStyles.documentSection, descendantStyles.documentSection)}>
      <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>{section.heading}</h2>
      {section.content.map((paragraph) => <p {...stylex.props(descendantStyles.documentSectionU_uP)} key={paragraph}>{paragraph}</p>)}
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
    <section id="usage" {...stylex.props(shellStyles.documentSection, descendantStyles.documentSection)}>
      <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>Usage</h2>
      <h3 {...stylex.props(descendantStyles.documentSectionU_uH3)}>Use when</h3>
      <SectionParagraphs section={sections?.whenToUse} />
      <h3 {...stylex.props(descendantStyles.documentSectionU_uH3)}>Avoid when</h3>
      <SectionParagraphs section={sections?.whenNotToUse} />
      <h3 {...stylex.props(descendantStyles.documentSectionU_uH3)}>Choosing between components</h3>
      <SectionParagraphs section={sections?.choosingBetween} />
      <h3 {...stylex.props(descendantStyles.documentSectionU_uH3)}>How it is composed</h3>
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
    <section id="examples" {...stylex.props(shellStyles.documentSection, shellStyles.documentSection_wide, descendantStyles.documentSection)}>
      <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>Examples</h2>
      <SectionParagraphs section={document.sections?.required.recommendedExample} />
      <SectionParagraphs section={document.sections?.required.implementation} />
      <ComponentWorkbench document={document} fixtures={fixtures} />
    </section>
  )
}

function VariantsAndStatesSection({ document }: { document: CatalogDocument }) {
  const sections = document.sections?.required
  return (
    <section id="variantsAndStates" {...stylex.props(shellStyles.documentSection, descendantStyles.documentSection)}>
      <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>Variants and states</h2>
      <h3 {...stylex.props(descendantStyles.documentSectionU_uH3)}>Variants</h3>
      <SectionParagraphs section={sections?.variants} />
      <h3 {...stylex.props(descendantStyles.documentSectionU_uH3)}>States and adaptation</h3>
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
    <div {...stylex.props(shellStyles.documentLayout)}>
      <article>
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
    <div {...stylex.props(shellStyles.documentLayout)}>
      <article>
        <div id="overview"><Hero document={document} /></div>
        <div id="specimen" {...stylex.props(shellStyles.documentSection, shellStyles.documentSection_wide, descendantStyles.documentSection)}>
          <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>Examples and usage</h2>
          <FoundationSpecimen document={document} />
        </div>
        <section
          id="guidance"
          {...stylex.props(shellStyles.documentSection, descendantStyles.documentSection)}
          aria-labelledby="foundation-guidance-heading"
        >
          <h2 id="foundation-guidance-heading" {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>
            Guidance
          </h2>
          <div {...stylex.props(shellStyles.documentListGrid)}>
            <LabeledList heading="When to use" items={document.guidance.whenToUse} />
            <LabeledList heading="Avoid when" items={document.guidance.whenNotToUse} />
            <LabeledList heading="Guidance" items={document.guidance.content} />
            <LabeledList heading="Common mistakes" items={document.guidance.commonMistakes} />
          </div>
        </section>

        <section id="accessibility" {...stylex.props(shellStyles.documentSection, descendantStyles.documentSection)}>
          <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>Accessibility</h2>
          <div {...stylex.props(shellStyles.documentListGrid)}>
            <LabeledList heading="Requirements" items={document.accessibility.requirements} />
            <LabeledList heading="Keyboard" items={document.accessibility.keyboard} />
            <LabeledList heading="Known constraints" items={document.accessibility.knownConstraints} />
          </div>
        </section>
        <section id="contract" {...stylex.props(shellStyles.documentSection, descendantStyles.documentSection)}>
          <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>Public contract</h2>
          <LabeledList heading="Defaults" items={document.api.defaults} />
          <LabeledList heading="Not public API" items={document.api.invalidCombinations} />
        </section>
        <section id="metadata" {...stylex.props(shellStyles.documentSection, descendantStyles.documentSection)}>
          <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>Metadata</h2>
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
    <div {...stylex.props(shellStyles.documentLayout)}>
      <article>
        <div id="overview"><Hero document={document} /></div>
        <section id="guidance" {...stylex.props(shellStyles.documentSection, descendantStyles.documentSection)}>
          <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>Guidance</h2>
          <SectionCopy value={document.guidance} fallback="No additional guidance is registered for this document." />
        </section>
        <section id="accessibility" {...stylex.props(shellStyles.documentSection, descendantStyles.documentSection)}>
          <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>Accessibility</h2>
          <SectionCopy value={document.accessibility} fallback="Accessibility requirements are tracked in the catalogue registry." />
        </section>
        {fixture ? (
          <div id="preview" {...stylex.props(shellStyles.documentSection, shellStyles.documentSection_wide, descendantStyles.documentSection)}>
            <ComponentWorkbench document={document} fixtures={[fixture]} />
          </div>
        ) : null}
        <GovernanceGuide document={document} />
        <section id="contract" {...stylex.props(shellStyles.documentSection, descendantStyles.documentSection)}>
          <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>Public contract</h2>
          <SectionCopy value={document.api} fallback="This foundation does not expose a component API." />
        </section>
        <section id="metadata" {...stylex.props(shellStyles.documentSection, descendantStyles.documentSection)}>
          <h2 {...stylex.props(typography.section, descendantStyles.documentSectionU_uH2)}>Metadata</h2>
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
