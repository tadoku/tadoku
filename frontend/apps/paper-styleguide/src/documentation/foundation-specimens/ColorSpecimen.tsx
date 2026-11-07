import { Button, cn, Flash, stack } from 'paper-ui'
import { GuideCode } from '../FoundationGuide'
import { FoundationCanvas } from './FoundationPreview'

const colorRoles = [
  ['Canvas', 'canvas', 'Page background.'],
  ['Paper', 'paper', 'Ordinary content surface.'],
  ['Raised', 'raised', 'Small nested emphasis.'],
  ['Overlay', 'overlay', 'Transient panels.'],
  ['Scrim', 'scrim', 'Translucent backdrop behind a modal; the dialog owns dismissal.'],
  ['Ink', 'ink', 'Primary text on neutral surfaces.'],
  ['Muted', 'muted', 'Supporting copy, not disabled state.'],
  ['Inverse text', 'inverse', 'Light foreground on a suitable dark background; not an automatic contrast guarantee.'],
  ['Link', 'link', 'Links with an underline.'],
  ['Subtle rule', 'rule-subtle', 'Low-emphasis division within related content.'],
  ['Rule', 'rule', 'Ordinary one-pixel separator.'],
  ['Strong rule', 'rule-strong', 'Floating surface boundary.'],
  ['Field edge', 'field-edge', 'The border of text fields and pickers.'],
  ['Button edge', 'button-edge', 'The border of outline buttons.'],
  ['Button shade', 'button-shade', 'The translucent edge of filled buttons, deeper at the bottom.'],
  ['Action', 'action', 'Filled action background; pair with on-action.'],
  ['Action hover', 'action-hover', 'Hover of filled actions.'],
  ['Action pressed', 'action-active', 'Pressed action feedback.'],
  ['Action soft', 'action-soft', 'Quiet selected background; use ink text.'],
  ['Neutral hover', 'neutral-hover', 'Hover background for outline and ghost actions.'],
  ['Action text', 'on-action', 'Foreground paired with filled action backgrounds.'],
  ['Destructive action', 'destructive', 'Destructive button background; pair with on-action.'],
  ['Destructive hover', 'destructive-hover', 'Hover background for destructive actions.'],
  ['Information', 'information', 'Status icon or rail plus an explicit label.'],
  ['Success', 'success', 'A successful outcome plus text.'],
  ['Warning', 'warning', 'A condition needing attention plus text.'],
  ['Danger', 'danger', 'Error text/icon; destructive buttons use action-destructive.'],
  ['Focus', 'focus', 'Keyboard focus, not selection.'],
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
  canvas: 'bg-canvas', paper: 'bg-paper', raised: 'bg-raised', overlay: 'bg-overlay', scrim: 'bg-scrim',
  'rule-subtle': 'border-rule-subtle', rule: 'border-rule', 'rule-strong': 'border-rule-strong',
  'field-edge': 'border-field-edge', 'button-edge': 'border-button-edge', 'button-shade': 'border-button-shade',
  action: 'bg-action', 'action-hover': 'hover:bg-action-hover', 'action-active': 'active:bg-action-active',
  'action-soft': 'bg-action-soft', 'neutral-hover': 'hover:bg-neutral-hover',
  destructive: 'bg-destructive', 'destructive-hover': 'hover:bg-destructive-hover',
  focus: 'outline-focus', 'focus-offset': 'outline-focus-offset',
}
export function ColorSpecimen() {
  return <>
    <FoundationCanvas><div className="grid grid-cols-guide-cards gap-6">{(['light', 'dark'] as const).map(theme => <section key={theme} data-theme={theme} className={cn(stack(), 'foundation-guide__sample')}>
      <h3>{theme === 'light' ? 'Light paper' : 'Dark paper'}</h3><p>Read in ink. Keep details muted.</p><p className="text-muted">Japanese · 48 pages</p><div><Button>Save reading</Button></div><Flash variant="success" title="Reading saved">48 pages added to your history.</Flash>
    </section>)}</div></FoundationCanvas>
    <p>Use <code>bg-canvas</code> for the page, <code>bg-paper</code> for content, <code>text-ink</code> for primary text and <code>text-muted</code> for supporting text from <code>paper-ui/theme.css</code>. These classes follow the nearest theme automatically. Use Button and Flash for complete action and status treatments; the reference gives utilities first and the token's custom property for computed values, as in a chart.</p>
    <p>Choose colors by job, not by their current hex value. In dark mode, link/focus violet is lighter than the filled-action violet. Using <code>text-action</code> for paragraph text is therefore incorrect.</p>
    <div className="foundation-reference my-6 overflow-x-auto [&_table]:w-full [&_table]:min-w-reference-table [&_table]:border-collapse [&_table]:text-start [&_table]:text-sm [&_caption]:mb-3 [&_caption]:text-start [&_caption]:font-bold [&_:is(th,td)]:border-b [&_:is(th,td)]:border-rule [&_:is(th,td)]:p-3 [&_:is(th,td)]:text-start [&_:is(th,td)]:align-top [&_:is(th,td)]:wrap-anywhere [&_th]:min-w-32 [&_th]:font-semibold max-xs:[&_:is(th,td)]:px-1 max-xs:[&_:is(th,td)]:py-2" role="region" aria-label="Semantic color reference" tabIndex={0}>
      <p className="hidden max-md:sticky max-md:start-0 max-md:block max-md:text-xs max-md:text-muted">Scroll horizontally to see all columns.</p>
      <table className="[&_th:nth-child(2)]:min-w-16 [&_th:nth-child(3)]:min-w-16 [&_td_p]:mb-0 [&_td_p]:mt-1.4 [&_code]:text-fine"><caption>Semantic color reference</caption>
        <thead><tr><th scope="col">Role</th><th scope="col">Light</th><th scope="col">Dark</th><th scope="col">Tailwind class, variable and use</th></tr></thead>
        <tbody>{colorRoles.map(([name, role, usage]) => <tr key={role}>
          <th scope="row">{name}</th>
          {(['light', 'dark'] as const).map(theme => <td key={theme}><span data-theme={theme} className="block h-8 w-12 flex-none border border-rule" style={{ background: `var(--color-${role})` }} aria-label={`${name} in ${theme} theme`} /></td>)}
          <td><code>{colorClasses[role] ?? `text-${role}`}</code><p>{usage}</p><small><code>{`--color-${role}`}</code></small></td>
        </tr>)}</tbody>
      </table>
    </div>
    <GuideCode label="Semantic colors (JSX)" code={'<section className="border border-rule bg-paper text-ink">\n  <a className="text-link underline" href="/logs">Reading log</a>\n  <p className="text-muted">Updated today</p>\n</section>'} />
    <h3>Chart and action pairs</h3><p><code>chart-1</code> through <code>chart-8</code> identify series, not success or warning. Keep a series’ index stable across views and add a legend, labels or patterns. Use <code>text-on-action</code> on filled actions; <code>text-inverse</code> alone does not choose a suitable background. Prefer Button and Flash so interaction colors and semantics remain consistent.</p>
  </>
}

