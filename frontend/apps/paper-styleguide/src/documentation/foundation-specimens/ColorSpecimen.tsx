import { Button, Flash } from 'paper-ui'
import { GuideCode } from '../FoundationGuide'
import { FoundationCanvas } from './FoundationPreview'

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
export function ColorSpecimen() {
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

