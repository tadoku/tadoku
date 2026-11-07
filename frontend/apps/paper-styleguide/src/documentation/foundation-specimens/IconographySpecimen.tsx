import IconActionsExample from '../foundation-examples/IconActionsExample'
import IconActionsExampleSource from '../foundation-examples/IconActionsExample.tsx?raw'
import { CheckCircleIcon, ExclamationTriangleIcon, iconClassName, MagnifyingGlassIcon, PlusIcon, XMarkIcon } from 'paper-ui/icons'
import { GuideTable } from '../FoundationGuide'
import { FoundationCanvas, FoundationExample } from './FoundationPreview'
import { cluster } from 'paper-ui'

export function IconographySpecimen() {
  const icons = [{name: 'PlusIcon', Icon: PlusIcon, use: 'Add reading'}, {name: 'MagnifyingGlassIcon', Icon: MagnifyingGlassIcon, use: 'Search'}, {name: 'XMarkIcon', Icon: XMarkIcon, use: 'Close'}, {name: 'CheckCircleIcon', Icon: CheckCircleIcon, use: 'Saved'}, {name: 'ExclamationTriangleIcon', Icon: ExclamationTriangleIcon, use: 'Needs attention'}]
  return <><FoundationCanvas><div className="grid grid-cols-icon-tiles gap-4 [&_div]:grid [&_div]:gap-2">{icons.map(({name, Icon, use}) => <div key={name}><Icon className={iconClassName()} aria-hidden="true"/><strong>{use}</strong><code>{name}</code></div>)}</div></FoundationCanvas>
    <h3>Match visual size to its job</h3><FoundationCanvas><div className={cluster()}>{(['compact', 'default', 'prominent', 'emptyState'] as const).map(size => <div key={size}><PlusIcon className={`paper-icon-${size === "emptyState" ? "empty-state" : size}`} aria-hidden="true"/><p>{size}</p></div>)}</div></FoundationCanvas>
    <GuideTable label="Icon size and meaning" columns={['CSS class', 'Pixels', 'Use']} rows={[
      ['iconClassName(\'compact\')', '16', 'Dense metadata and small supporting icons.'], ['iconClassName()', '20', 'Actions and navigation.'], ['iconClassName(\'prominent\')', '24', 'A leading status or important action.'], ['iconClassName(\'emptyState\')', '48', 'Empty-state illustration; not a toolbar action.'],
    ]}/><FoundationExample Example={IconActionsExample} code={IconActionsExampleSource} />
    <p>Size an icon with <code>{'<PlusIcon className={iconClassName()} />'}</code>. The helper returns the size utilities, such as <code>size-icon-sm</code> for compact, merged with any classes you pass as its second argument.</p>
    <p>Outline icons represent actions/navigation; solid icons carry status/confirmation. Icons inherit currentColor. A visible text label is preferred; icon-only controls need an accessible name and a 44px target independent of glyph size. Button hides leading/trailing icons from assistive technology; hide decorative child SVGs explicitly.</p><p>The curated export includes arrows, chevrons, bars, ellipsis, search, edit, add, delete and close outlines; check, information, warning and error solids. Add a new glyph only when none of these communicates the intended action.</p>
  </>
}

