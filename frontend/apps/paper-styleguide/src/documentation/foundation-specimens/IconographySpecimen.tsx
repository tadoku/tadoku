import * as stylex from '@stylexjs/stylex'
import { iconSizes, layout } from 'paper-ui'
import IconActionsExample from '../foundation-examples/IconActionsExample'
import IconActionsExampleSource from '../foundation-examples/IconActionsExample.tsx?raw'
import { CheckCircleIcon, PlusIcon, XMarkIcon, MagnifyingGlassIcon, ExclamationTriangleIcon } from 'paper-ui/icons'
import { GuideTable } from '../FoundationGuide'
import { FoundationCanvas, FoundationExample } from './FoundationPreview'
import { foundationStyles } from '../foundation-guide.styles'
import { descendantStyles } from '../../descendants.styles'

export function IconographySpecimen() {
  const icons = [{name: 'PlusIcon', Icon: PlusIcon, use: 'Add reading'}, {name: 'MagnifyingGlassIcon', Icon: MagnifyingGlassIcon, use: 'Search'}, {name: 'XMarkIcon', Icon: XMarkIcon, use: 'Close'}, {name: 'CheckCircleIcon', Icon: CheckCircleIcon, use: 'Saved'}, {name: 'ExclamationTriangleIcon', Icon: ExclamationTriangleIcon, use: 'Needs attention'}]
  return <><FoundationCanvas><div {...stylex.props(foundationStyles.foundationIconGrid)}>{icons.map(({name, Icon, use}) => <div {...stylex.props(descendantStyles.foundationIconGridU_uDiv)} key={name}><Icon {...stylex.props(iconSizes.default)} aria-hidden="true"/><strong>{use}</strong><code>{name}</code></div>)}</div></FoundationCanvas>
    <h3>Match visual size to its job</h3><FoundationCanvas><div {...stylex.props(layout.cluster)}>{(['compact', 'default', 'prominent', 'emptyState'] as const).map(size => <div key={size}><PlusIcon {...stylex.props(iconSizes[size])} aria-hidden="true"/><p>{size}</p></div>)}</div></FoundationCanvas>
    <GuideTable label="Icon size and meaning" columns={['StyleX export', 'Pixels', 'Use']} rows={[
      ['iconSizes.compact', '16', 'Dense metadata and small supporting icons.'], ['iconSizes.default', '20', 'Actions and navigation.'], ['iconSizes.prominent', '24', 'A leading status or important action.'], ['iconSizes.emptyState', '48', 'Empty-state illustration; not a toolbar action.'],
    ]}/><FoundationExample Example={IconActionsExample} code={IconActionsExampleSource} />
    <p>Spread the size on the icon: <code>{'<PlusIcon {...stylex.props(iconSizes.default)} />'}</code>.</p>
    <p>Outline icons represent actions/navigation; solid icons carry status/confirmation. Icons inherit currentColor. A visible text label is preferred; icon-only controls need an accessible name and a 44px target independent of glyph size. Button hides leading/trailing icons from assistive technology; hide decorative child SVGs explicitly.</p><p>The curated export includes arrows, chevrons, bars, ellipsis, search, edit, add, delete and close outlines; check, information, warning and error solids. Add a new glyph only when none of these communicates the intended action.</p>
  </>
}

