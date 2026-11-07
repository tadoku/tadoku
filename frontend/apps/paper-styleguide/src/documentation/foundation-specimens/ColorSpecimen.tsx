import * as stylex from '@stylexjs/stylex'
import { Button, designTokens, Flash, layout, themes, typography } from 'paper-ui'
import { color } from 'paper-ui/tokens.stylex'
import { GuideCode } from '../FoundationGuide'
import { FoundationCanvas } from './FoundationPreview'
import { foundationStyles } from '../foundation-guide.styles'
import { descendantStyles } from '../../descendants.styles'

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
const tokenName = (role: string) => role.replace(/-(\w)/g, (_, letter: string) => letter.toUpperCase()) as keyof typeof designTokens.colorRoles.light
const swatch = stylex.create({ fill: (value: string) => ({ backgroundColor: value }) })
export function ColorSpecimen() {
  return <>
    <FoundationCanvas><div {...stylex.props(foundationStyles.foundationGuide__grid)}>{(['light', 'dark'] as const).map(theme => <section key={theme} data-theme={theme} {...stylex.props(themes[theme], foundationStyles.foundationGuide__sample, layout.stack)}>
      <h3 {...stylex.props(descendantStyles.foundationGuideU_uSampleU_uAny)}>{theme === 'light' ? 'Light paper' : 'Dark paper'}</h3><p {...stylex.props(descendantStyles.foundationGuideU_uSampleU_uAny)}>Read in ink. Keep details muted.</p><p {...stylex.props(typography.muted, descendantStyles.foundationGuideU_uSampleU_uAny)}>Japanese · 48 pages</p><div {...stylex.props(descendantStyles.foundationGuideU_uSampleU_uAny)}><Button>Save reading</Button></div><Flash variant="success" title="Reading saved">48 pages added to your history.</Flash>
    </section>)}</div></FoundationCanvas>
    <p>Use <code>color.surfaceCanvas</code> for the page, <code>color.surfacePaper</code> for content, <code>color.textInk</code> for primary text and <code>color.textMuted</code> for supporting text, imported from <code>paper-ui/tokens.stylex</code>. Colour tokens follow the nearest theme automatically. Use Button and Flash for complete action and status treatments.</p>
    <p>Choose colors by job, not by their current hex value. In dark mode, link/focus violet is lighter than the filled-action violet. Using <code>action-default</code> for paragraph text is therefore incorrect.</p>
    <div {...stylex.props(foundationStyles.foundationReference)} role="region" aria-label="Semantic color reference" tabIndex={0}>
      <p {...stylex.props(foundationStyles.foundationReferenceHint)}>Scroll horizontally to see all columns.</p>
      <table {...stylex.props(foundationStyles.foundationColorTable, descendantStyles.foundationReferenceU_uTable)}><caption {...stylex.props(descendantStyles.foundationReferenceU_uCaption)}>Semantic color reference</caption>
        <thead><tr><th {...stylex.props(descendantStyles.foundationReferenceU_uTh, descendantStyles.foundationColorTableU_uTh)} scope="col">Role</th><th {...stylex.props(descendantStyles.foundationReferenceU_uTh, descendantStyles.foundationColorTableU_uTh)} scope="col">Light</th><th {...stylex.props(descendantStyles.foundationReferenceU_uTh, descendantStyles.foundationColorTableU_uTh)} scope="col">Dark</th><th {...stylex.props(descendantStyles.foundationReferenceU_uTh, descendantStyles.foundationColorTableU_uTh)} scope="col">Token and use</th></tr></thead>
        <tbody>{colorRoles.map(([name, role, usage]) => <tr key={role}>
          <th {...stylex.props(descendantStyles.foundationReferenceU_uTh, descendantStyles.foundationColorTableU_uTh)} scope="row">{name}</th>
          {(['light', 'dark'] as const).map(theme => <td {...stylex.props(descendantStyles.foundationReferenceU_uTd)} key={theme}><span data-theme={theme} {...stylex.props(themes[theme], foundationStyles.foundationColorSwatch, swatch.fill(color[tokenName(role)]))} aria-label={`${name} in ${theme} theme`} /></td>)}
          <td {...stylex.props(descendantStyles.foundationReferenceU_uTd)}><code {...stylex.props(descendantStyles.foundationColorTableU_uCode)}>{`color.${tokenName(role)}`}</code><p {...stylex.props(descendantStyles.foundationColorTableU_uTdU_uP)}>{usage}</p></td>
        </tr>)}</tbody>
      </table>
    </div>
    <GuideCode label="Semantic colors (StyleX)" code={'const styles = stylex.create({\n  panel: {\n    backgroundColor: color.surfacePaper,\n    borderColor: color.ruleDefault,\n    borderStyle: "solid",\n    borderWidth: border.static,\n    color: color.textInk,\n  },\n  details: { color: color.textMuted },\n});'} />
    <h3>Chart and action pairs</h3><p><code>color.chart1</code> through <code>color.chart8</code> identify series, not success or warning. Keep a series’ index stable across views and add a legend, labels or patterns. Use <code>color.actionText</code> on filled actions; <code>color.textInverse</code> alone does not choose a suitable background. Prefer Button and Flash so interaction colors and semantics remain consistent.</p>
  </>
}

