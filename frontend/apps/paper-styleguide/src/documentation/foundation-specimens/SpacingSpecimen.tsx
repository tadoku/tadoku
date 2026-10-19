import ContainerSpacingExample from '../foundation-examples/ContainerSpacingExample'
import ContainerSpacingExampleSource from '../foundation-examples/ContainerSpacingExample.tsx?raw'
import CompactEntryExample from '../foundation-examples/CompactEntryExample'
import CompactEntryExampleSource from '../foundation-examples/CompactEntryExample.tsx?raw'
import { useState } from 'react'
import { usePaperForm, Button } from 'paper-ui'
import { GuideTable } from '../FoundationGuide'
import { FoundationCanvas, FoundationExample } from './FoundationPreview'

const spaces = [[1, 4, 'Icon/text micro-spacing'], [2, 8, 'Closely related items'], [3, 12, 'Small content groups'], [4, 16, 'Default stack and surface padding'], [6, 24, 'Related sections'], [8, 32, 'Page groups'], [12, 48, 'Major page sections']] as const
function DensityExample({ density }: { density: 'comfortable' | 'compact' }) {
  const methods = usePaperForm({ defaultValues: { work: 'コンビニ人間', pages: '48' }, onSubmit: () => setSaved(true) })
  const [saved, setSaved] = useState(false)
  return <section data-density={density} className="foundation-guide__sample foundation-density-example">
    <h3>{density === 'comfortable' ? 'Comfortable' : 'Compact'}</h3>
    <methods.AppForm><methods.Form >
      <methods.AppField name="work">{field => <field.Input label="Work" />}</methods.AppField><methods.AppField name="pages">{field => <field.Input label="Pages" type="number" min={1} />}</methods.AppField>
      <div className="paper-cluster"><Button type="submit">Save reading</Button><Button variant="outline" onClick={() => {methods.reset(); setSaved(false)}}>Reset</Button></div>
      <p role="status">{saved ? 'Demo entry saved locally.' : 'Japanese · Not submitted to a contest'}</p>
    </methods.Form></methods.AppForm>
  </section>
}
export function SpacingSpecimen() {
  return <>
    <h3>Spacing classes</h3><p>Use <code>p-4</code> for 16px of padding inside a container and <code>gap-4</code> for 16px between flex or grid children. Add <code>paper-ui/tailwind-preset</code> in <a href="/setup#tailwind">Setup</a> once; these are ordinary Tailwind utilities mapped to Paper’s values. The number is a quarter-rem step: 4 × 0.25rem = 1rem.</p><p>The scale is shared in both densities. Choose the smallest interval that makes a relationship clear: 8px within a group, 16px between fields, 24–48px between sections. These are starting points, not a rule that every gap must grow with its container.</p>
    <FoundationCanvas><ul className="foundation-space-scale" aria-label="Spacing scale">{spaces.map(([step, px]) => <li key={step}><code>p-{step}</code><small>{px}px</small><span style={{ inlineSize: `var(--paper-space-${step})` }} aria-hidden="true"/></li>)}</ul></FoundationCanvas>
    <GuideTable label="Spacing class reference" columns={['Padding / gap class', 'rem / px at 16px root', 'Starting use']} rows={spaces.map(([step, px, use]) => [`p-${step} / gap-${step}`, `${px / 16}rem / ${px}px`, use])} />
    <GuideTable label="Which spacing utility to use" columns={['Class', 'Effect', 'Use']} rows={[
      ['p-4', '16px padding on every side.', 'Space inside a surface.'], ['px-4 / py-2', '16px horizontal / 8px vertical padding.', 'Adjust container axes independently.'], ['ps-4 / pe-4', '16px on the logical start / end side.', 'Directional layouts that also support right-to-left text.'], ['mt-6 / mb-6 / mx-auto', '24px outside the top / bottom; automatic horizontal margins.', 'Separate blocks or center a bounded container.'], ['flex gap-2 / grid gap-4', '8px / 16px between children.', 'Related actions / fields; a gap needs flex or grid.'], ['gap-x-4 / gap-y-2', '16px between columns / 8px between rows.', 'Wrapped action groups and grids.'], ['p-4 md:p-8', '16px by default; 32px from Tailwind’s md breakpoint.', 'Increase page gutters when the viewport has room.'], ['gap-inline', '10px comfortable; 8px compact.', 'Density-sensitive inline controls.'],
    ]} />
    <p>The seven values above are Paper’s preferred rhythm. The preset extends Tailwind’s scale, so values such as <code>p-5</code>, <code>p-10</code>, <code>w-full</code> and <code>max-w-6xl</code> still work. Use spacing utilities on composition containers; Button and form controls already own their internal padding.</p>
    <FoundationExample Example={ContainerSpacingExample} label="Container spacing" code={ContainerSpacingExampleSource} />
    <p>For custom CSS without Tailwind, <code>padding: var(--paper-space-4)</code> gives the same 16px as <code>p-4</code>. The <code>paper-stack</code> and <code>paper-cluster</code> helpers also remain available from the standalone stylesheet. Do not change the global scale for one screen.</p>
    <h3>The same task in two densities</h3><FoundationCanvas><div className="foundation-guide__grid"><DensityExample density="comfortable"/><DensityExample density="compact"/></div></FoundationCanvas>
    <GuideTable label="What density actually changes" columns={['Property', 'Comfortable', 'Compact']} rows={[
      ['--paper-control-height', '2.75rem / 44px minimum', '2.25rem / 36px minimum'], ['--paper-inline-gap', '0.625rem / 10px', '0.5rem / 8px'], ['--paper-field-padding-block', '0.625rem / 10px', '0.375rem / 6px'], ['--paper-field-padding-inline', '0.75rem / 12px', '0.625rem / 10px'], ['Body size / line height', '1rem / 1.625rem', '0.875rem / 1.375rem'], ['Headings, spacing scale, border and rail widths', 'Fixed', 'Fixed'],
    ]}/>
    <p>Comfortable is the default, especially for touch and reading. Compact is an explicit choice for dense administration; it is not a mobile breakpoint. Controls can grow for wrapped text. Compact does not promise a 44px touch target for every text button. Keep icon-only actions at least 44px and use comfortable where touch precision matters.</p>
    <FoundationExample Example={CompactEntryExample} code={CompactEntryExampleSource} />
    <p>Set density once on the app root. A nested boundary changes tokens, but inherited body typography needs <code>paper-type-body</code> on that boundary. Portalled dialogs need the document root’s values; nesting does not move a portal into your local boundary.</p>
  </>
}
