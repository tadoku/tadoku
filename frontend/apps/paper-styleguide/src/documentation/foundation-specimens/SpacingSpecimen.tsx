import ContainerSpacingExample from '../foundation-examples/ContainerSpacingExample'
import ContainerSpacingExampleSource from '../foundation-examples/ContainerSpacingExample.tsx?raw'
import CompactEntryExample from '../foundation-examples/CompactEntryExample'
import CompactEntryExampleSource from '../foundation-examples/CompactEntryExample.tsx?raw'
import { useState } from 'react'
import { Button, cluster, usePaperForm } from 'paper-ui'
import { GuideTable } from '../FoundationGuide'
import { FoundationCanvas, FoundationExample } from './FoundationPreview'

const spaces = [[1, 4, 'Icon/text micro-spacing'], [2, 8, 'Closely related items'], [3, 12, 'Small content groups'], [4, 16, 'Default stack and surface padding'], [6, 24, 'Related sections'], [8, 32, 'Page groups'], [12, 48, 'Major page sections']] as const
function DensityExample({ density }: { density: 'comfortable' | 'compact' }) {
  const methods = usePaperForm({ defaultValues: { work: 'コンビニ人間', pages: '48' }, onSubmit: () => setSaved(true) })
  const [saved, setSaved] = useState(false)
  return <section data-density={density} className="foundation-guide__sample min-w-0 border border-rule bg-paper p-4 text-ink [&>:first-child]:mt-0 [&>:last-child]:mb-0 text-body leading-body [&_form]:grid [&_form]:gap-inline">
    <h3>{density === 'comfortable' ? 'Comfortable' : 'Compact'}</h3>
    <methods.AppForm><methods.Form >
      <methods.AppField name="work">{field => <field.Input label="Work" />}</methods.AppField><methods.AppField name="pages">{field => <field.Input label="Pages" type="number" min={1} />}</methods.AppField>
      <div className={cluster()}><Button type="submit">Save reading</Button><Button variant="outline" onClick={() => {methods.reset(); setSaved(false)}}>Reset</Button></div>
      <p role="status">{saved ? 'Demo entry saved locally.' : 'Japanese · Not submitted to a contest'}</p>
    </methods.Form></methods.AppForm>
  </section>
}
export function SpacingSpecimen() {
  return <>
    <h3>Spacing classes</h3><p>Use <code>p-4</code> for 16px of padding inside a container and <code>gap-4</code> for 16px between flex or grid children. Import <code>paper-ui/theme.css</code> as in <a href="/setup#tailwind">Setup</a>; the spacing utilities come from Paper’s spacing tokens, and steps outside the scale do not compile. The number is a quarter-rem step: 4 × 0.25rem = 1rem.</p><p>The scale is shared in both densities. Choose the smallest interval that makes a relationship clear: 8px within a group, 16px between fields, 24–48px between sections. These are starting points, not a rule that every gap must grow with its container.</p>
    <FoundationCanvas><ul className="grid list-none gap-3 p-0 [&_li]:grid [&_li]:grid-cols-space-scale [&_li]:items-center [&_li]:gap-2 max-xs:[&_li]:grid-cols-space-scale-narrow max-xs:[&_li]:text-sm [&_span]:block [&_span]:h-4 [&_span]:bg-action" aria-label="Spacing scale">{spaces.map(([step, px]) => <li key={step}><code>p-{step}</code><small>{px}px</small><span style={{ inlineSize: `var(--spacing-${step})` }} aria-hidden="true"/></li>)}</ul></FoundationCanvas>
    <GuideTable label="Spacing class reference" columns={['Padding / gap class', 'rem / px at 16px root', 'Starting use']} rows={spaces.map(([step, px, use]) => [`p-${step} / gap-${step}`, `${px / 16}rem / ${px}px`, use])} />
    <GuideTable label="Which spacing utility to use" columns={['Class', 'Effect', 'Use']} rows={[
      ['p-4', '16px padding on every side.', 'Space inside a surface.'], ['px-4 / py-2', '16px horizontal / 8px vertical padding.', 'Adjust container axes independently.'], ['ps-4 / pe-4', '16px on the logical start / end side.', 'Directional layouts that also support right-to-left text.'], ['mt-6 / mb-6 / mx-auto', '24px outside the top / bottom; automatic horizontal margins.', 'Separate blocks or center a bounded container.'], ['flex gap-2 / grid gap-4', '8px / 16px between children.', 'Related actions / fields; a gap needs flex or grid.'], ['gap-x-4 / gap-y-2', '16px between columns / 8px between rows.', 'Wrapped action groups and grids.'], ['p-4 md:p-8', '16px by default; 32px from the md breakpoint.', 'Increase page gutters when the viewport has room.'], ['gap-inline', '10px comfortable; 8px compact.', 'Density-sensitive inline controls.'],
    ]} />
    <p>The seven values above are Paper’s preferred rhythm. The token scale also has the 2px and 4px steps between them, such as <code>p-5</code>, plus named sizes such as <code>max-w-measure</code>; anything else does not compile. Use spacing utilities on composition containers; Button and form controls already own their internal padding.</p>
    <FoundationExample Example={ContainerSpacingExample} label="Container spacing" code={ContainerSpacingExampleSource} />
    <p>Paper ships no stylesheet to write against: <code>stack()</code> and <code>cluster()</code> compose vertical flow and wrapping rows from the same tokens. Add a step to <code>paper-ui/src/tokens</code> only when a value recurs; do not change the global scale for one screen.</p>
    <h3>The same task in two densities</h3><FoundationCanvas><div className="grid grid-cols-guide-cards gap-6"><DensityExample density="comfortable"/><DensityExample density="compact"/></div></FoundationCanvas>
    <GuideTable label="What density actually changes" columns={['Property', 'Comfortable', 'Compact']} rows={[
      ['min-h-control', '2.75rem / 44px minimum', '2.25rem / 36px minimum'], ['gap-inline', '0.625rem / 10px', '0.5rem / 8px'], ['py-field-y', '0.625rem / 10px', '0.375rem / 6px'], ['px-field-x', '0.75rem / 12px', '0.625rem / 10px'], ['Body size / line height', '1rem / 1.625rem', '0.875rem / 1.375rem'], ['Headings, spacing scale, border and rail widths', 'Fixed', 'Fixed'],
    ]}/>
    <p>Comfortable is the default, especially for touch and reading. Compact is an explicit choice for dense administration; it is not a mobile breakpoint. Controls can grow for wrapped text. Compact does not promise a 44px touch target for every text button. Keep icon-only actions at least 44px and use comfortable where touch precision matters.</p>
    <FoundationExample Example={CompactEntryExample} code={CompactEntryExampleSource} />
    <p>Set density once on the app root. A nested boundary changes tokens, but inherited body typography needs <code>{"typography({ role: 'body' })"}</code> on that boundary. Portalled dialogs need the document root’s values; nesting does not move a portal into your local boundary.</p>
  </>
}
