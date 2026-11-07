import * as stylex from '@stylexjs/stylex'
import RailExample from '../foundation-examples/RailExample'
import RailExampleSource from '../foundation-examples/RailExample.tsx?raw'
import { accentRail, designTokens, Flash, layout, Surface, typography } from 'paper-ui'
import { GuideTable } from '../FoundationGuide'
import { FoundationCanvas, FoundationExample } from './FoundationPreview'
import { foundationStyles } from '../foundation-guide.styles'

export function ShapeSpecimen() {
  return <>
    <FoundationCanvas><div {...stylex.props(foundationStyles.foundationGuide__grid)}><Surface xstyle={[layout.stack]}><h3 {...stylex.props(typography.component)}>Quiet perimeter</h3><p>One pixel groups related content. No depth or rail is needed.</p></Surface><Surface accent xstyle={[layout.stack]}><h3 {...stylex.props(typography.component)}>Bordered accent rail</h3><p>The 3px rail covers the leading perimeter and meets the top and bottom edges squarely.</p></Surface></div>
    <h3 {...stylex.props(accentRail.base, foundationStyles.foundationRailHeading, typography.section)}>A borderless section heading</h3><p>The same rail works without a card. Give the heading its own content padding. The rail has no interactive or status meaning.</p>
    <Flash variant="warning" title="Contest ends tomorrow">A status rail belongs to Flash and uses a status color. It is not a recolored decorative accent.</Flash></FoundationCanvas>
    <GuideTable label="Border and rail recipes" columns={['StyleX export', 'What it supplies']} rows={[
      ['surfaces.paper / surfaces.raised', 'One-pixel perimeter, background and text color. Add your own container padding, such as space[16].'],
      ['accentRail.base', 'A leading positioned rail. Add accentRail.onBorder on a bordered host for the border join; the host owns content space.'],
      ['edges.field', 'One-pixel border with a two-pixel lower field edge. Appearance only; prefer Input/Select for complete controls.'],
      ['edges.action', 'Action-colored border with a two-pixel lower edge. Appearance only; prefer Button for complete actions.'],
    ]} />
    <GuideTable label="Border and rail contract" columns={['Role', 'Value', 'Use']} rows={[
      ['border.static', designTokens.border.static, 'Static surface perimeter.'], ['border.fieldEdge', designTokens.border.fieldEdge, 'Subtle lower edge on pale controls.'], ['border.actionEdge', designTokens.border.actionEdge, 'Stronger lower edge on filled actions.'], ['border.accentRail', designTokens.border.accentRail, 'Leading accent; includes the 1px perimeter on a Surface.'], ['Geometry', 'Square corners', 'No rounded end caps or diagonal border joins.'],
    ]}/>
    <FoundationExample Example={RailExample} code={RailExampleSource} />
    <p><code>Surface accent</code> supplies perimeter, padding and the rail. <code>accentRail.base</code> alone supplies only a positioned pseudo-element. Its host owns spacing and semantics. Do not nest rails, clip the host with overflow hidden, reuse its <code>::before</code>, or apply a custom thick border without adapting the recipe.</p><p>Rails use the logical leading edge, so right-to-left containers mirror correctly. Focus is a separate outline around the actual control; adding a rail does not make a Surface clickable.</p>
  </>
}

