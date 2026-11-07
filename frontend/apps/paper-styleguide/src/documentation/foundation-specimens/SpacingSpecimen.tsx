import * as stylex from '@stylexjs/stylex'
import ContainerSpacingExample from '../foundation-examples/ContainerSpacingExample'
import ContainerSpacingExampleSource from '../foundation-examples/ContainerSpacingExample.tsx?raw'
import CompactEntryExample from '../foundation-examples/CompactEntryExample'
import CompactEntryExampleSource from '../foundation-examples/CompactEntryExample.tsx?raw'
import { useState } from 'react'
import { Button, densities, designTokens, layout, usePaperForm } from 'paper-ui'
import { space } from 'paper-ui/tokens.stylex'

const scale = stylex.create({ bar: (value: string) => ({ inlineSize: value }) })
import { GuideTable } from '../FoundationGuide'
import { FoundationCanvas, FoundationExample } from './FoundationPreview'
import { foundationStyles } from '../foundation-guide.styles'
import { descendantStyles } from '../../descendants.styles'

const spaces = [[1, 4, 'Icon/text micro-spacing'], [2, 8, 'Closely related items'], [3, 12, 'Small content groups'], [4, 16, 'Default stack and surface padding'], [6, 24, 'Related sections'], [8, 32, 'Page groups'], [12, 48, 'Major page sections']] as const
function DensityExample({ density }: { density: 'comfortable' | 'compact' }) {
  const methods = usePaperForm({ defaultValues: { work: 'コンビニ人間', pages: '48' }, onSubmit: () => setSaved(true) })
  const [saved, setSaved] = useState(false)
  return <section data-density={density} {...stylex.props(densities[density], foundationStyles.foundationGuide__sample, foundationStyles.foundationDensityExample)}>
    <h3 {...stylex.props(descendantStyles.foundationGuideU_uSampleU_uAny)}>{density === 'comfortable' ? 'Comfortable' : 'Compact'}</h3>
    <methods.AppForm {...stylex.props(descendantStyles.foundationGuideU_uSampleU_uAny)}><methods.Form >
      <methods.AppField name="work">{field => <field.Input label="Work" />}</methods.AppField><methods.AppField name="pages">{field => <field.Input label="Pages" type="number" min={1} />}</methods.AppField>
      <div {...stylex.props(layout.cluster)}><Button type="submit">Save reading</Button><Button variant="outline" onClick={() => {methods.reset(); setSaved(false)}}>Reset</Button></div>
      <p role="status">{saved ? 'Demo entry saved locally.' : 'Japanese · Not submitted to a contest'}</p>
    </methods.Form></methods.AppForm>
  </section>
}
export function SpacingSpecimen() {
  return <>
    <h3>Spacing tokens</h3><p>Use <code>space[16]</code> from <code>paper-ui/tokens.stylex</code> for 16px of padding inside a container and as the <code>gap</code> between flex or grid children. Space tokens are keyed by their pixel value at a 16px root and resolve to rem.</p><p>The scale is shared in both densities. Choose the smallest interval that makes a relationship clear: 8px within a group, 16px between fields, 24–48px between sections. These are starting points, not a rule that every gap must grow with its container.</p>
    <FoundationCanvas><ul {...stylex.props(foundationStyles.foundationSpaceScale)} aria-label="Spacing scale">{spaces.map(([step, px]) => <li {...stylex.props(descendantStyles.foundationSpaceScaleU_uLi)} key={step}><code>space[{px}]</code><small>{px}px</small><span {...stylex.props(descendantStyles.foundationSpaceScaleU_uSpan, scale.bar(space[px]))} aria-hidden="true"/></li>)}</ul></FoundationCanvas>
    <GuideTable label="Spacing token reference" columns={['Token', 'rem / px at 16px root', 'Starting use']} rows={spaces.map(([, px, use]) => [`space[${px}]`, `${px / 16}rem / ${px}px`, use])} />
    <GuideTable label="Which spacing property to use" columns={['StyleX property', 'Effect', 'Use']} rows={[
      ['padding: space[16]', '16px padding on every side.', 'Space inside a surface.'], ['paddingInline / paddingBlock', 'Horizontal / vertical padding.', 'Adjust container axes independently.'], ['paddingInlineStart / End', 'Padding on the logical start / end side.', 'Directional layouts that also support right-to-left text.'], ['marginBlockStart: space[24]', '24px before a block.', 'Separate blocks; prefer gap inside flex and grid.'], ['gap: space[8] / space[16]', '8px / 16px between children.', 'Related actions / fields; a gap needs flex or grid.'], ['columnGap / rowGap', 'Different column and row spacing.', 'Wrapped action groups and grids.'], ['{ default: space[16], [breakpoints.md]: space[32] }', 'A value per breakpoint.', 'Increase page gutters when the viewport has room.'], ['density.inlineGap', '10px comfortable; 8px compact.', 'Density-sensitive inline controls.'],
    ]} />
    <p>The seven values above are Paper’s preferred rhythm; the full scale lists every space step Paper and its apps use. Add a step to the token source rather than writing a raw length, which the lint rule rejects. Button and form controls already own their internal padding.</p>
    <FoundationExample Example={ContainerSpacingExample} label="Container spacing" code={ContainerSpacingExampleSource} />
    <h3>The same task in two densities</h3><FoundationCanvas><div {...stylex.props(foundationStyles.foundationGuide__grid)}><DensityExample density="comfortable"/><DensityExample density="compact"/></div></FoundationCanvas>
    <GuideTable label="What density actually changes" columns={['Property', 'Comfortable', 'Compact']} rows={[
      ...Object.keys(designTokens.density.comfortable).map(name => [`density.${name}`, designTokens.density.comfortable[name], designTokens.density.compact[name]]),
      ['Headings, spacing scale, border and rail widths', 'Fixed', 'Fixed'],
    ]}/>
    <p>Comfortable is the default, especially for touch and reading. Compact is an explicit choice for dense administration; it is not a mobile breakpoint. Controls can grow for wrapped text. Compact does not promise a 44px touch target for every text button. Keep icon-only actions at least 44px and use comfortable where touch precision matters.</p>
    <FoundationExample Example={CompactEntryExample} code={CompactEntryExampleSource} />
    <p>Set density once on the app root with <code>applyDensity</code>. A nested boundary takes <code>densities.compact</code> through <code>stylex.props</code>; inherited body typography also needs <code>typography.body</code> on that boundary. Portalled dialogs need the document root’s values; nesting does not move a portal into your local boundary.</p>
  </>
}
