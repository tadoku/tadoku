import * as stylex from '@stylexjs/stylex'
import LayoutExample from '../foundation-examples/LayoutExample'
import LayoutExampleSource from '../foundation-examples/LayoutExample.tsx?raw'
import OverflowExample from '../foundation-examples/OverflowExample'
import OverflowExampleSource from '../foundation-examples/OverflowExample.tsx?raw'
import { GuideCode, GuideTable } from '../FoundationGuide'
import { FoundationCanvas, FoundationExample } from './FoundationPreview'
import { foundationStyles } from '../foundation-guide.styles'
import { descendantStyles } from '../../descendants.styles'
import { size } from 'paper-ui/tokens.stylex'

const wide = stylex.create({ table: { minInlineSize: size[576] } })

export function LayoutSpecimen() {
  return <>
    <h3>Start with one shared page frame</h3><p>Give the header, main content and footer the same maximum width and responsive gutters. Align their content edges first, then create narrower reading and form columns inside that frame. A Navbar’s internal padding also counts toward its content edge.</p>
    <FoundationExample Example={LayoutExample} code={LayoutExampleSource} />
    <GuideTable label="Public composition recipes" columns={['StyleX export', 'Behavior', 'Override']} rows={[
      ['layout.stack', 'Vertical flex flow with a 16px gap.', 'Follow it with your own style, such as { gap: space[24] }.'], ['layout.cluster', 'Wrapping row with a density-aware inline gap.', 'Follow it with { gap: space[12] } for a fixed gap.'], ['layout.measure', 'Maximum 65ch inline size; long words can break.', 'Use another measure token, such as measure.prose.'], ['layout.flow', 'Removes block margins and lets a flex or grid child shrink.', 'Use on columns with long titles or wide tables.'],
    ]}/>
    <p>Compose page grids in your own <code>stylex.create</code> styles with tokens and the <code>breakpoints</code> constants. Applications own their column proportions, navigation arrangement and handlers.</p>
    <h3>Let the layout choose field widths</h3><p>Paper fields fill their parent. Bound a form with a measure or size token and put a related pair of fields in a two-column grid from a breakpoint. A field outside that pair spans the same total width. Use a smaller column for short values; controls should not impose an unrelated maximum width.</p><p>See the <a href="/components/forms/input">Input example</a> for a working form with a full-width title and aligned date and amount fields.</p>
    <GuideCode label="Responsive page grid (StyleX)" code={'const styles = stylex.create({\n  page: {\n    marginInline: "auto",\n    maxInlineSize: page.contentWidth,\n    paddingBlock: space[32],\n  },\n  columns: {\n    display: "grid",\n    gap: space[24],\n    gridTemplateColumns: {\n      default: "minmax(0, 1fr)",\n      [breakpoints.lg]: "minmax(0, 2fr) minmax(0, 1fr)",\n    },\n  },\n});'} />
    <h3>Wide data gets a named scroll region</h3><p>Keep the page fluid at 320 CSS pixels. A table may scroll horizontally inside its own labelled, keyboard-focusable region; ordinary text and actions should wrap. Do not hide page overflow to conceal broken widths.</p>
    <FoundationCanvas><div {...stylex.props(foundationStyles.foundationReference)} role="region" aria-label="Reading totals; scroll horizontally" tabIndex={0}><table {...stylex.props(descendantStyles.foundationReferenceU_uTable, wide.table)}><caption {...stylex.props(descendantStyles.foundationReferenceU_uCaption)}>Reading totals · horizontal data example</caption><thead><tr><th {...stylex.props(descendantStyles.foundationReferenceU_uTh)} scope="col">Language</th><th {...stylex.props(descendantStyles.foundationReferenceU_uTh)} scope="col">September 1</th><th {...stylex.props(descendantStyles.foundationReferenceU_uTh)} scope="col">September 2</th><th {...stylex.props(descendantStyles.foundationReferenceU_uTh)} scope="col">September 3</th><th {...stylex.props(descendantStyles.foundationReferenceU_uTh)} scope="col">Total pages</th></tr></thead><tbody><tr><th {...stylex.props(descendantStyles.foundationReferenceU_uTh)} scope="row">Japanese</th><td {...stylex.props(descendantStyles.foundationReferenceU_uTd)}>32</td><td {...stylex.props(descendantStyles.foundationReferenceU_uTd)}>48</td><td {...stylex.props(descendantStyles.foundationReferenceU_uTd)}>16</td><td {...stylex.props(descendantStyles.foundationReferenceU_uTd)}>96</td></tr></tbody></table></div></FoundationCanvas>
    <FoundationExample Example={OverflowExample} label="Overflow wrapper (JSX)" code={OverflowExampleSource} />
  </>
}

