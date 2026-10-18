import IconActionsExample from '../foundation-examples/IconActionsExample'
import IconActionsExampleSource from '../foundation-examples/IconActionsExample.tsx?raw'
import { CheckCircleIcon, PlusIcon, XMarkIcon, MagnifyingGlassIcon, ExclamationTriangleIcon } from 'paper-ui/icons'
import { GuideTable } from '../FoundationGuide'
import { FoundationCanvas, FoundationExample } from './FoundationPreview'

export function IconographySpecimen() {
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

