import LayoutExample from '../foundation-examples/LayoutExample'
import LayoutExampleSource from '../foundation-examples/LayoutExample.tsx?raw'
import OverflowExample from '../foundation-examples/OverflowExample'
import OverflowExampleSource from '../foundation-examples/OverflowExample.tsx?raw'
import { GuideCode, GuideTable } from '../FoundationGuide'
import { FoundationCanvas, FoundationExample } from './FoundationPreview'

export function LayoutSpecimen() {
  return <>
    <h3>Start with one shared page frame</h3><p>Give the header, main content and footer the same maximum width and responsive gutters. Align their content edges first, then create narrower reading and form columns inside that frame. A Navbar’s internal padding also counts toward its content edge.</p>
    <FoundationExample Example={LayoutExample} code={LayoutExampleSource} />
    <GuideTable label="Public composition utilities" columns={['Class', 'Behavior', 'Override']} rows={[
      ['flex flex-col gap-4', 'Vertical flow with a 16px gap.', 'Use gap-6 for 24px.'], ['flex flex-wrap items-center gap-inline', 'Wrapping row with a density-aware inline gap.', 'Use gap-3 for a fixed 12px.'], ['grid grid-cols-1 lg:grid-cols-2 gap-6', 'One column; two from Tailwind’s lg breakpoint.', 'The app chooses columns and breakpoints.'], ['mx-auto w-full max-w-6xl', 'Full width up to 72rem, centered.', 'Add p-4 md:p-8 for responsive padding.'], ['min-w-0', 'Allows a flex/grid child to shrink below its content’s natural width.', 'Use on columns with long titles or wide tables.'], ['paper-stack', 'Standalone CSS: vertical flow, child margin reset and 16px gap.', 'Add gap-6 with Tailwind, or set --paper-stack-gap in CSS.'], ['paper-cluster', 'Standalone CSS: wrapping row with density-aware inline gap.', 'Add gap-3 with Tailwind, or set --paper-cluster-gap in CSS.'], ['paper-measure', 'Maximum 65ch inline size; long words can break.', 'Use max-w-prose with Tailwind or custom CSS for another measure.'],
    ]}/>
    <p>Compose page grids with Tailwind’s responsive utilities and Paper components. Applications own their column proportions, navigation arrangement and handlers. The standalone Paper helpers remain useful when Tailwind is unavailable.</p>
    <h3>Let the layout choose field widths</h3><p>Paper fields fill their parent. Use <code>w-full max-w-2xl</code> on a form and <code>grid gap-4 sm:grid-cols-2</code> on a related pair of fields. A field outside that pair spans the same total width. Use a smaller column for short values; controls should not impose an unrelated maximum width.</p><p>See the <a href="/components/forms/input">Input example</a> for a working form with a full-width title and aligned date and amount fields.</p>
    <GuideCode label="Responsive page grid (application-owned CSS)" code={'.reading-page {\n  inline-size: min(100% - 2rem, 72rem);\n  margin-inline: auto;\n  padding-block: var(--paper-space-8);\n}\n.reading-columns {\n  display: grid;\n  grid-template-columns: minmax(0, 1fr);\n  gap: var(--paper-space-6);\n}\n@media (min-width: 60rem) {\n  .reading-columns { grid-template-columns: minmax(0, 2fr) minmax(0, 1fr); }\n}\n.reading-columns > * { min-inline-size: 0; }'} />
    <h3>Wide data gets a named scroll region</h3><p>Keep the page fluid at 320 CSS pixels. A table may scroll horizontally inside its own labelled, keyboard-focusable region; ordinary text and actions should wrap. Do not hide page overflow to conceal broken widths.</p>
    <FoundationCanvas><div className="foundation-reference" role="region" aria-label="Reading totals; scroll horizontally" tabIndex={0}><table style={{ minWidth: '34rem' }}><caption>Reading totals · horizontal data example</caption><thead><tr><th scope="col">Language</th><th scope="col">September 1</th><th scope="col">September 2</th><th scope="col">September 3</th><th scope="col">Total pages</th></tr></thead><tbody><tr><th scope="row">Japanese</th><td>32</td><td>48</td><td>16</td><td>96</td></tr></tbody></table></div></FoundationCanvas>
    <FoundationExample Example={OverflowExample} label="Overflow wrapper (JSX)" code={OverflowExampleSource} />
  </>
}

