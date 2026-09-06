import PrinciplesExample from './foundation-examples/PrinciplesExample'
import PrinciplesExampleSource from './foundation-examples/PrinciplesExample.tsx?raw'
import TypographyExample from './foundation-examples/TypographyExample'
import TypographyExampleSource from './foundation-examples/TypographyExample.tsx?raw'
import ContainerSpacingExample from './foundation-examples/ContainerSpacingExample'
import ContainerSpacingExampleSource from './foundation-examples/ContainerSpacingExample.tsx?raw'
import CompactEntryExample from './foundation-examples/CompactEntryExample'
import CompactEntryExampleSource from './foundation-examples/CompactEntryExample.tsx?raw'
import LayoutExample from './foundation-examples/LayoutExample'
import LayoutExampleSource from './foundation-examples/LayoutExample.tsx?raw'
import OverflowExample from './foundation-examples/OverflowExample'
import OverflowExampleSource from './foundation-examples/OverflowExample.tsx?raw'
import RailExample from './foundation-examples/RailExample'
import RailExampleSource from './foundation-examples/RailExample.tsx?raw'
import ElevationExample from './foundation-examples/ElevationExample'
import ElevationExampleSource from './foundation-examples/ElevationExample.tsx?raw'
import IconActionsExample from './foundation-examples/IconActionsExample'
import IconActionsExampleSource from './foundation-examples/IconActionsExample.tsx?raw'
import BrandLinkExample from './foundation-examples/BrandLinkExample'
import BrandLinkExampleSource from './foundation-examples/BrandLinkExample.tsx?raw'
import { createContext, useContext, useState, type ComponentType, type CSSProperties, type ReactNode } from 'react'
import { FormProvider, useForm, useWatch } from 'react-hook-form'
import { Button, Flash, Input, Select, Surface } from 'paper-ui'
import cutMeterUrl from 'paper-ui/assets/brand/cut-meter.svg?no-inline'
import cutMeterReversedUrl from 'paper-ui/assets/brand/cut-meter-reversed.svg?no-inline'
import wordmarkUrl from 'paper-ui/assets/brand/wordmark-accent.svg?no-inline'
import wordmarkReversedUrl from 'paper-ui/assets/brand/wordmark-reversed.svg?no-inline'
import type { CatalogDocument } from 'paper-ui/catalog'
import { CheckCircleIcon, PlusIcon, XMarkIcon, MagnifyingGlassIcon, ExclamationTriangleIcon } from 'paper-ui/icons'
import { GuideCode, GuideTable } from './FoundationGuide'

const PreviewSettings = createContext({ theme: 'light', density: 'comfortable' })

function FoundationCanvas({ children }: { children: ReactNode }) {
  const { theme, density } = useContext(PreviewSettings)
  return <div data-theme={theme} data-density={density} className="foundation-guide__preview paper-type-body">{children}</div>
}

function FoundationExample({ Example, code, label = 'Working example' }: { Example: ComponentType<{ theme?: 'light' | 'dark' }>; code: string; label?: string }) {
  const { theme } = useContext(PreviewSettings)
  return <><h3>{label}</h3><FoundationCanvas><Example theme={theme as 'light' | 'dark'} /></FoundationCanvas><GuideCode code={code} label="Source for this example" /></>
}

function PrinciplesSpecimen() {
  return <>
    <FoundationExample Example={PrinciplesExample} code={PrinciplesExampleSource} />
    <GuideTable label="Turn a principle into a decision" columns={['Principle', 'Use it like this', 'Avoid']} rows={[
      ['Reading first', 'Work title and progress lead. Metadata supports them. Use a short, useful next action.', 'Decorative charts or motivational scores competing with the reading task.'],
      ['Quiet structure', 'Use a heading and space for a section; a border only when it marks a real group.', 'A raised card around every paragraph and nested accent rails.'],
      ['Recognizable behavior', 'Links navigate, buttons act, labels stay visible and status includes words.', 'Clickable generic containers, icon-only mystery actions and color-only success.'],
      ['Purposeful complexity', 'Start with native document flow; add a component when its behavior is needed.', 'A new component for every visual arrangement or a menu for two visible actions.'],
    ]} />
    <p>Heading level follows the surrounding document, while its type class controls appearance. The example’s link is application navigation; route it to your own reading page.</p>
  </>
}

const colorRoles = [
  ['Canvas', 'surface-canvas', 'Page background.'],
  ['Paper', 'surface-paper', 'Ordinary content surface.'],
  ['Raised', 'surface-raised', 'Small nested emphasis.'],
  ['Overlay', 'surface-overlay', 'Transient panels.'],
  ['Scrim', 'surface-scrim', 'Translucent backdrop behind a modal; the dialog owns dismissal.'],
  ['Ink', 'text-ink', 'Primary text on neutral surfaces.'],
  ['Muted', 'text-muted', 'Supporting copy, not disabled state.'],
  ['Inverse text', 'text-inverse', 'Light foreground on a suitable dark background; not an automatic contrast guarantee.'],
  ['Link', 'text-link', 'Links with an underline.'],
  ['Subtle rule', 'rule-subtle', 'Low-emphasis division within related content.'],
  ['Rule', 'rule-default', 'Ordinary one-pixel separator.'],
  ['Strong rule', 'rule-strong', 'Floating surface boundary.'],
  ['Field lower edge', 'rule-field-edge', 'The lower border of pale controls.'],
  ['Action lower edge', 'rule-action-edge', 'The lower border of filled actions.'],
  ['Destructive lower edge', 'rule-destructive-action-edge', 'The lower border of destructive actions.'],
  ['Action', 'action-default', 'Filled action background; pair with action-text.'],
  ['Action hover', 'action-hover', 'Hover of filled actions.'],
  ['Action pressed', 'action-active', 'Pressed action feedback.'],
  ['Action soft', 'action-soft', 'Quiet selected background; use ink text.'],
  ['Neutral hover', 'action-neutral-hover', 'Hover background for outline and ghost actions.'],
  ['Action text', 'action-text', 'Foreground paired with filled action backgrounds.'],
  ['Destructive action', 'action-destructive', 'Destructive button background; pair with action-text.'],
  ['Destructive hover', 'action-destructive-hover', 'Hover background for destructive actions.'],
  ['Information', 'status-information', 'Status icon or rail plus an explicit label.'],
  ['Success', 'status-success', 'A successful outcome plus text.'],
  ['Warning', 'status-warning', 'A condition needing attention plus text.'],
  ['Danger', 'status-danger', 'Error text/icon; destructive buttons use action-destructive.'],
  ['Focus', 'focus-ring', 'Keyboard focus, not selection.'],
  ['Focus offset', 'focus-offset', 'Separation between a control and its focus ring.'],
  ['Chart 1', 'chart-1', 'First data series; keep its identity stable between views.'],
  ['Chart 2', 'chart-2', 'Second data series, with a label or legend.'],
  ['Chart 3', 'chart-3', 'Third data series, with a label or legend.'],
  ['Chart 4', 'chart-4', 'Fourth data series, with a label or legend.'],
  ['Chart 5', 'chart-5', 'Fifth data series, with a label or legend.'],
  ['Chart 6', 'chart-6', 'Sixth data series, with a label or legend.'],
  ['Chart 7', 'chart-7', 'Seventh data series, with a label or legend.'],
  ['Chart 8', 'chart-8', 'Eighth data series, with a label or legend.'],
] as const
const colorClasses: Record<string, string> = {
  'surface-canvas': 'bg-canvas', 'surface-paper': 'bg-paper', 'surface-raised': 'bg-raised',
  'surface-overlay': 'bg-overlay', 'surface-scrim': 'bg-scrim',
  'text-ink': 'text-ink', 'text-muted': 'text-muted', 'text-inverse': 'text-inverse', 'text-link': 'text-link',
  'rule-subtle': 'border-rule-subtle', 'rule-default': 'border-rule', 'rule-strong': 'border-rule-strong',
  'rule-field-edge': 'border-field-edge', 'rule-action-edge': 'border-action-edge', 'rule-destructive-action-edge': 'border-destructive-edge',
  'action-default': 'bg-action', 'action-hover': 'hover:bg-action-hover', 'action-active': 'active:bg-action-active',
  'action-soft': 'bg-action-soft', 'action-neutral-hover': 'hover:bg-neutral-hover', 'action-text': 'text-action-text',
  'action-destructive': 'bg-destructive', 'action-destructive-hover': 'hover:bg-destructive-hover',
  'status-information': 'text-information', 'status-success': 'text-success', 'status-warning': 'text-warning', 'status-danger': 'text-danger',
  'focus-ring': 'outline-focus', 'focus-offset': 'ring-offset-focus-offset',
}
function ColorSpecimen() {
  return <>
    <FoundationCanvas><div className="foundation-guide__grid">{(['light', 'dark'] as const).map(theme => <section key={theme} data-theme={theme} className="foundation-guide__sample paper-stack">
      <h3>{theme === 'light' ? 'Light paper' : 'Dark paper'}</h3><p>Read in ink. Keep details muted.</p><p className="paper-text-muted">Japanese · 48 pages</p><div><Button>Save reading</Button></div><Flash variant="success" title="Reading saved">48 pages added to your history.</Flash>
    </section>)}</div></FoundationCanvas>
    <p>Use <code>bg-canvas</code> for the page, <code>bg-paper</code> for content, <code>text-ink</code> for primary text and <code>text-muted</code> for supporting text through <code>paper-ui/tailwind-preset</code>. These classes follow the nearest theme automatically. Use Button and Flash for complete action and status treatments; the reference gives classes first and CSS variables for custom compositions.</p>
    <p>Choose colors by job, not by their current hex value. In dark mode, link/focus violet is lighter than the filled-action violet. Using <code>action-default</code> for paragraph text is therefore incorrect.</p>
    <div className="foundation-reference" role="region" aria-label="Semantic color reference" tabIndex={0}>
      <p className="foundation-reference-hint">Scroll horizontally to see all columns.</p>
      <table className="foundation-color-table"><caption>Semantic color reference</caption>
        <thead><tr><th scope="col">Role</th><th scope="col">Light</th><th scope="col">Dark</th><th scope="col">Tailwind class, variable and use</th></tr></thead>
        <tbody>{colorRoles.map(([name, role, usage]) => <tr key={role}>
          <th scope="row">{name}</th>
          {(['light', 'dark'] as const).map(theme => <td key={theme}><span data-theme={theme} className="foundation-color-swatch" style={{ background: `var(--paper-color-${role})` }} aria-label={`${name} in ${theme} theme`} /></td>)}
          <td><code>{colorClasses[role] ?? `text-${role}`}</code><p>{usage}</p><small><code>{`--paper-color-${role}`}</code></small></td>
        </tr>)}</tbody>
      </table>
    </div>
    <GuideCode label="Semantic colors (application CSS)" code={'.reading-panel {\n  background: var(--paper-color-surface-paper);\n  color: var(--paper-color-text-ink);\n  border: 1px solid var(--paper-color-rule-default);\n}\n.reading-panel a {\n  color: var(--paper-color-text-link);\n  text-decoration: underline;\n}\n.reading-panel__details { color: var(--paper-color-text-muted); }'} />
    <h3>Chart and action pairs</h3><p><code>--paper-color-chart-1</code> through <code>-8</code> identify series, not success or warning. Keep a series’ index stable across views and add a legend, labels or patterns. Use <code>--paper-color-action-text</code> on filled actions; <code>text-inverse</code> alone does not choose a suitable background. Prefer Button and Flash so interaction colors and semantics remain consistent.</p>
  </>
}

const typeRoles = [
  ['display', 'A season of reading', '48 / 52', '700', 'Rare landing-page display; allow wrapping.'], ['page', 'Your reading', '32 / 38', '700', 'One page title.'], ['section', 'Recent activity', '24 / 30', '700', 'Major page sections.'], ['component', 'September Japanese', '18 / 24', '700', 'Component and card headings.'], ['body', 'Make time for a few more pages.', '16 / 26; compact 14 / 22', '400', 'Reading copy and interface explanations.'], ['label', 'Reading language', '12 / 16', '700', 'Short visible labels. No forced uppercase.'], ['metadata', 'Updated September 5', '12 / 18', '400', 'Supporting context; never essential instructions alone.'],
] as const
function TypographySpecimen() {
  return <>
    <FoundationCanvas><ul className="foundation-type-list">{typeRoles.map(([role, text]) => <li key={role}><span className={`paper-type-${role}`}>{text}</span><code>paper-type-{role}</code></li>)}</ul></FoundationCanvas>
    <GuideTable label="Type roles at a 16px root" columns={['CSS class', 'Size / line height (px)', 'Weight', 'Use']} rows={typeRoles.map(([role, , size, weight, use]) => [`paper-type-${role}`, size, weight, use])} />
    <FoundationExample Example={TypographyExample} code={TypographyExampleSource} />
    <p>The full stylesheet includes the bundled fonts: Merriweather ships at 700; Open Sans at 400, 600 and 700. Unshipped italics or weights may be synthesized by the browser. Japanese and other scripts use available fallback fonts. Test realistic multilingual titles rather than assuming every script shares Latin metrics.</p><p>Type classes do not choose heading levels. Keep DOM headings sequential, use body text for instructions and let text wrap. The fixed scale uses rem units and follows browser zoom.</p>
  </>
}

const spaces = [[1, 4, 'Icon/text micro-spacing'], [2, 8, 'Closely related items'], [3, 12, 'Small content groups'], [4, 16, 'Default stack and surface padding'], [6, 24, 'Related sections'], [8, 32, 'Page groups'], [12, 48, 'Major page sections']] as const
function DensityExample({ density }: { density: 'comfortable' | 'compact' }) {
  const methods = useForm({ defaultValues: { work: 'コンビニ人間', pages: '48' } })
  const [saved, setSaved] = useState(false)
  return <section data-density={density} className="foundation-guide__sample foundation-density-example">
    <h3>{density === 'comfortable' ? 'Comfortable' : 'Compact'}</h3>
    <FormProvider {...methods}><form onSubmit={methods.handleSubmit(() => setSaved(true))}>
      <Input name="work" label="Work" /><Input name="pages" label="Pages" type="number" min={1} />
      <div className="paper-cluster"><Button type="submit">Save reading</Button><Button variant="outline" onClick={() => {methods.reset(); setSaved(false)}}>Reset</Button></div>
      <p role="status">{saved ? 'Demo entry saved locally.' : 'Japanese · Not submitted to a contest'}</p>
    </form></FormProvider>
  </section>
}
function SpacingSpecimen() {
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

function LayoutSpecimen() {
  return <>
    <FoundationExample Example={LayoutExample} code={LayoutExampleSource} />
    <GuideTable label="Public composition utilities" columns={['Class', 'Behavior', 'Override']} rows={[
      ['flex flex-col gap-4', 'Vertical flow with a 16px gap.', 'Use gap-6 for 24px.'], ['flex flex-wrap items-center gap-inline', 'Wrapping row with a density-aware inline gap.', 'Use gap-3 for a fixed 12px.'], ['grid grid-cols-1 lg:grid-cols-2 gap-6', 'One column; two from Tailwind’s lg breakpoint.', 'The app chooses columns and breakpoints.'], ['mx-auto w-full max-w-6xl', 'Full width up to 72rem, centered.', 'Add p-4 md:p-8 for responsive padding.'], ['min-w-0', 'Allows a flex/grid child to shrink below its content’s natural width.', 'Use on columns with long titles or wide tables.'], ['paper-stack', 'Standalone CSS: vertical flow, child margin reset and 16px gap.', 'Add gap-6 with Tailwind, or set --paper-stack-gap in CSS.'], ['paper-cluster', 'Standalone CSS: wrapping row with density-aware inline gap.', 'Add gap-3 with Tailwind, or set --paper-cluster-gap in CSS.'], ['paper-measure', 'Maximum 65ch inline size; long words can break.', 'Use max-w-prose with Tailwind or custom CSS for another measure.'],
    ]}/>
    <p>Compose page grids with Tailwind’s responsive utilities and Paper components. Applications own their column proportions, navigation arrangement and handlers. The standalone Paper helpers remain useful when Tailwind is unavailable.</p>
    <GuideCode label="Responsive page grid (application-owned CSS)" code={'.reading-page {\n  inline-size: min(100% - 2rem, 72rem);\n  margin-inline: auto;\n  padding-block: var(--paper-space-8);\n}\n.reading-columns {\n  display: grid;\n  grid-template-columns: minmax(0, 1fr);\n  gap: var(--paper-space-6);\n}\n@media (min-width: 60rem) {\n  .reading-columns { grid-template-columns: minmax(0, 2fr) minmax(0, 1fr); }\n}\n.reading-columns > * { min-inline-size: 0; }'} />
    <h3>Wide data gets a named scroll region</h3><p>Keep the page fluid at 320 CSS pixels. A table may scroll horizontally inside its own labelled, keyboard-focusable region; ordinary text and actions should wrap. Do not hide page overflow to conceal broken widths.</p>
    <FoundationCanvas><div className="foundation-reference" role="region" aria-label="Reading totals; scroll horizontally" tabIndex={0}><table style={{ minWidth: '34rem' }}><caption>Reading totals · horizontal data example</caption><thead><tr><th scope="col">Language</th><th scope="col">September 1</th><th scope="col">September 2</th><th scope="col">September 3</th><th scope="col">Total pages</th></tr></thead><tbody><tr><th scope="row">Japanese</th><td>32</td><td>48</td><td>16</td><td>96</td></tr></tbody></table></div></FoundationCanvas>
    <FoundationExample Example={OverflowExample} label="Overflow wrapper (JSX)" code={OverflowExampleSource} />
  </>
}

function ShapeSpecimen() {
  return <>
    <FoundationCanvas><div className="foundation-guide__grid"><Surface className="paper-stack"><h3 className="paper-type-component">Quiet perimeter</h3><p>One pixel groups related content. No depth or rail is needed.</p></Surface><Surface accent className="paper-stack"><h3 className="paper-type-component">Bordered accent rail</h3><p>The 3px rail covers the leading perimeter and meets the top and bottom edges squarely.</p></Surface></div>
    <h3 className="paper-accent-rail foundation-rail-heading paper-type-section">A borderless section heading</h3><p>The same rail works without a card. Give the heading its own content padding. The rail has no interactive or status meaning.</p>
    <Flash variant="warning" title="Contest ends tomorrow">A status rail belongs to Flash and uses a status color. It is not a recolored decorative accent.</Flash></FoundationCanvas>
    <GuideTable label="Border and rail classes" columns={['CSS class', 'What it supplies']} rows={[
      ['paper-surface / paper-surface-raised', 'One-pixel perimeter, background and text color. Add your own container padding, such as p-4.'],
      ['paper-accent-rail', 'A leading positioned rail. Combine with a surface class for a border join; add p-4 for content space.'],
      ['paper-field-edge', 'One-pixel border with a two-pixel lower field edge. Appearance only; prefer Input/Select for complete controls.'],
      ['paper-action-edge', 'Action-colored border with a two-pixel lower edge. Appearance only; prefer Button for complete actions.'],
    ]} />
    <GuideTable label="Border and rail contract" columns={['Role', 'Value', 'Use']} rows={[
      ['--paper-border-static-width', '1px', 'Static surface perimeter.'], ['--paper-border-field-edge-width', '2px', 'Subtle lower edge on pale controls.'], ['--paper-border-action-edge-width', '2px', 'Stronger lower edge on filled actions.'], ['--paper-accent-rail-width', '3px', 'Leading accent; includes the 1px perimeter on a Surface.'], ['Geometry', 'Square corners', 'No rounded end caps or diagonal border joins.'],
    ]}/>
    <FoundationExample Example={RailExample} code={RailExampleSource} />
    <p><code>Surface accent</code> supplies perimeter, padding and the rail. <code>paper-accent-rail</code> alone supplies only a positioned pseudo-element. Its host owns spacing and semantics. Do not nest rails, clip the host with overflow hidden, reuse its <code>::before</code>, or apply a custom thick border without adapting the recipe.</p><p>Rails use the logical leading edge, so right-to-left containers mirror correctly. Focus is a separate outline around the actual control; adding a rail does not make a Surface clickable.</p>
  </>
}

function ElevationSpecimen() {
  return <><FoundationCanvas><div className="foundation-guide__grid">{(['flat', 'floating', 'showcase'] as const).map(elevation => <Surface key={elevation} elevation={elevation} className="paper-stack"><h3 className="paper-type-component">{elevation}</h3><p>{elevation === 'flat' ? 'A reading summary in normal flow.' : elevation === 'floating' ? 'A transient layer, such as a menu.' : 'A rare visual study or featured demonstration.'}</p></Surface>)}</div></FoundationCanvas>
    <GuideTable label="Elevation roles" columns={['Surface prop / class', 'Offset', 'Contract']} rows={[
      ['flat / paper-elevation-flat', 'none', 'Default. A border supplies separation; no shadow.'], ['floating / paper-elevation-floating', '3px right and down; no blur', 'Transient layer. Stronger perimeter in both themes.'], ['showcase / paper-elevation-showcase', '5px right and down; no blur', 'Rare deliberate emphasis; not a card default.'],
    ]}/><FoundationExample Example={ElevationExample} code={ElevationExampleSource} />
    <p>Elevation controls appearance only. A floating Surface is not a dialog or menu: use Modal, Drawer or ActionMenu for focus management, dismissal and positioning. Shadows take no layout space; leave room at container edges and do not clip them. In forced colors, the border remains the boundary.</p>
  </>
}

function IconographySpecimen() {
  const icons = [{name: 'PlusIcon', Icon: PlusIcon, use: 'Add reading'}, {name: 'MagnifyingGlassIcon', Icon: MagnifyingGlassIcon, use: 'Search'}, {name: 'XMarkIcon', Icon: XMarkIcon, use: 'Close'}, {name: 'CheckCircleIcon', Icon: CheckCircleIcon, use: 'Saved'}, {name: 'ExclamationTriangleIcon', Icon: ExclamationTriangleIcon, use: 'Needs attention'}]
  return <><FoundationCanvas><div className="foundation-icon-grid">{icons.map(({name, Icon, use}) => <div key={name}><Icon className="paper-icon-default" aria-hidden="true"/><strong>{use}</strong><code>{name}</code></div>)}</div></FoundationCanvas>
    <h3>Match visual size to its job</h3><FoundationCanvas><div className="paper-cluster">{(['compact', 'default', 'prominent', 'emptyState'] as const).map(size => <div key={size}><PlusIcon className={`paper-icon-${size === "emptyState" ? "empty-state" : size}`} aria-hidden="true"/><p>{size}</p></div>)}</div></FoundationCanvas>
    <GuideTable label="Icon size and meaning" columns={['CSS class', 'Pixels', 'Use']} rows={[
      ['paper-icon-compact', '16', 'Dense metadata and small supporting icons.'], ['paper-icon-default', '20', 'Actions and navigation.'], ['paper-icon-prominent', '24', 'A leading status or important action.'], ['paper-icon-empty-state', '48', 'Empty-state illustration; not a toolbar action.'],
    ]}/><FoundationExample Example={IconActionsExample} code={IconActionsExampleSource} />
    <p>Use the CSS class directly on an icon: <code>{'<PlusIcon className="paper-icon-default" />'}</code>. The optional <code>iconClassName()</code> helper only returns a class string: <code>{'iconClassName("compact") === "paper-icon-compact"'}</code>. It does not render an icon, register styles or do anything you need for ordinary use.</p>
    <p>Outline icons represent actions/navigation; solid icons carry status/confirmation. Icons inherit currentColor. A visible text label is preferred; icon-only controls need an accessible name and a 44px target independent of glyph size. Button hides leading/trailing icons from assistive technology; hide decorative child SVGs explicitly.</p><p>The curated export includes arrows, chevrons, bars, ellipsis, search, edit, add, delete and close outlines; check, information, warning and error solids. Add a new glyph only when none of these communicates the intended action.</p>
  </>
}

function MotionSpecimen() {
  const [moved, setMoved] = useState(false)
  return <><FoundationCanvas><div className="foundation-motion-demo" data-moved={moved}><div><Button variant="outline" onClick={() => setMoved(!moved)}>{moved ? 'Move markers back' : 'Compare motion'}</Button></div>{([['Quick feedback', 'quick', '120ms'], ['Standard transition', 'standard', '180ms'], ['Deliberate movement', 'deliberate', '240ms']] as const).map(([label, role, duration]) => <div key={role}><strong>{label} · {duration}</strong><div className="foundation-motion-demo__track" aria-hidden="true" style={{ '--demo-duration': `var(--paper-motion-${role})` } as CSSProperties}><span className="foundation-motion-demo__marker"/></div><code>--paper-motion-{role}</code></div>)}</div></FoundationCanvas>
    <p>Activate the button to compare durations. Nothing moves automatically. With reduced motion enabled, markers change position immediately. Motion explains a change that already happened; it must never delay the result or be the only indication of status.</p>
    <GuideCode label="Transition with a reduced-motion override (CSS)" code={'.reading-highlight {\n  background: var(--paper-color-surface-paper);\n  transition: background-color var(--paper-motion-quick) ease-out;\n}\n.reading-highlight[data-selected="true"] {\n  background: var(--paper-color-action-soft);\n}\n@media (prefers-reduced-motion: reduce) {\n  .reading-highlight { transition: none; }\n}'} />
    <p>Public <code>paper-motion-quick</code>, <code>paper-motion-standard</code> and <code>paper-motion-deliberate</code> classes set duration only. They do not select properties, easing or provide a reduced-motion override. When writing custom animation, supply that override yourself. Prefer specific color/opacity/transform transitions over <code>transition: all</code>.</p>
    <GuideTable label="Choosing a duration" columns={['CSS duration class', 'Use', 'Avoid']} rows={[
      ['paper-motion-quick · 120ms', 'Hover/pressed feedback and small emphasis changes.', 'Large spatial movement.'], ['paper-motion-standard · 180ms', 'A small user-triggered reveal or position change.', 'Animating every list item when the page loads.'], ['paper-motion-deliberate · 240ms', 'Rare larger movement that helps orientation.', 'Delaying a confirmation or blocking the next action.'],
    ]}/>
  </>
}

function BrandSpecimen() {
  return <><FoundationCanvas><div className="foundation-guide__grid foundation-brand-guide">{(['light', 'dark'] as const).map(theme => <figure key={theme} data-theme={theme} className="foundation-guide__sample paper-stack"><img src={theme === 'light' ? wordmarkUrl : wordmarkReversedUrl} width={158} height={29} alt="Tadoku"/><img src={theme === 'light' ? cutMeterUrl : cutMeterReversedUrl} width={64} height={64} alt=""/><figcaption>{theme === 'light' ? 'Original wordmark on light paper' : 'Reversed on dark paper'}</figcaption></figure>)}</div></FoundationCanvas>
    <GuideTable label="Choose the packaged artwork" columns={['Asset', 'Use']} rows={[
      ['wordmark-accent.svg / wordmark-reversed.svg', 'Primary identity where the full Tadoku name fits; preserve 158:29 proportions.'], ['cut-meter.svg / cut-meter-reversed.svg', 'Compact identity when the name is already nearby or an accessible label supplies it.'], ['wordmark.svg / cut-meter.svg', 'Monochrome artwork for contexts that require a single ink. The normal wordmark keeps its original violet K.'], ['cut-meter-wordmark.svg', 'Packaged combined lockup; do not rebuild from individually positioned bars.'], ['favicon.svg; favicon-16.svg, -32.svg, -48.svg', 'Small browser identity with the approved small-size geometry.'], ['apple-touch-icon.svg', 'Opaque 180px source; rasterize when the target platform requires PNG.'],
    ]}/>
    <FoundationExample Example={BrandLinkExample} label="Brand link (Vite asset imports)" code={BrandLinkExampleSource} />
    <p>The asset import above assumes Vite’s asset handling and <code>vite/client</code> types. Other bundlers should copy the packaged file to a public asset URL. External SVG images do not inherit the page’s currentColor; choose the reversed asset for dark surfaces instead of applying a CSS filter.</p><p>Keep artwork unstretched and preserve its view box. Supply one accessible name per link: if the link already says “Tadoku home,” the image is decorative. Formal minimum size and clear-space specifications remain a brand decision; the examples show current package dimensions rather than inventing a new standard.</p>
  </>
}

const specimens: Readonly<Record<string, () => JSX.Element>> = {
  'foundation.principles': PrinciplesSpecimen, 'foundation.color': ColorSpecimen,
  'foundation.typography': TypographySpecimen, 'foundation.spacing-and-density': SpacingSpecimen,
  'foundation.layout': LayoutSpecimen, 'foundation.shape-and-borders': ShapeSpecimen,
  'foundation.elevation': ElevationSpecimen, 'foundation.iconography': IconographySpecimen,
  'foundation.motion': MotionSpecimen, 'foundation.brand': BrandSpecimen,
}
export function FoundationSpecimen({ document }: { document: CatalogDocument }) {
  const Specimen = specimens[document.id]
  const methods = useForm({ defaultValues: { theme: 'light', density: 'comfortable' } })
  const theme = useWatch({ control: methods.control, name: 'theme' })
  const density = useWatch({ control: methods.control, name: 'density' })
  return <section className="foundation-guide" aria-label={`${document.name} specimen`} data-foundation-specimen={document.id}>
    <FormProvider {...methods}><div className="foundation-guide__grid" role="group" aria-label="Foundation preview settings">
      <Select name="theme" label="Example theme" options={[{value: 'light', label: 'Light'}, {value: 'dark', label: 'Dark'}]} />
      <Select name="density" label="Example density" options={[{value: 'comfortable', label: 'Comfortable'}, {value: 'compact', label: 'Compact'}]} />
    </div></FormProvider>
    <p className="foundation-guide__settings-note">These controls affect the visual examples below. See <a href="/setup">Setup</a> for imports and application settings.</p>
    <PreviewSettings.Provider value={{ theme, density }}>
      {Specimen ? <Specimen/> : null}
    </PreviewSettings.Provider>
  </section>
}
