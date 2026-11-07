import ElevationExample from '../foundation-examples/ElevationExample'
import ElevationExampleSource from '../foundation-examples/ElevationExample.tsx?raw'
import { stack, Surface, typography } from 'paper-ui'
import { GuideTable } from '../FoundationGuide'
import { FoundationCanvas, FoundationExample } from './FoundationPreview'

export function ElevationSpecimen() {
  return <><FoundationCanvas><div className="grid grid-cols-guide-cards gap-6">{(['flat', 'floating', 'showcase'] as const).map(elevation => <Surface key={elevation} elevation={elevation} className={stack()}><h3 className={typography({ role: 'component' })}>{elevation}</h3><p>{elevation === 'flat' ? 'A reading summary in normal flow.' : elevation === 'floating' ? 'A transient layer, such as a menu.' : 'A rare visual study or featured demonstration.'}</p></Surface>)}</div></FoundationCanvas>
    <GuideTable label="Elevation roles" columns={['Surface prop / class', 'Offset', 'Contract']} rows={[
      ['flat · shadow-none', 'none', 'Default. A border supplies separation; no shadow.'], ['floating · shadow-floating', '3px right and down; no blur', 'Transient layer. Stronger perimeter in both themes.'], ['showcase · shadow-showcase', '5px right and down; no blur', 'Rare deliberate emphasis; not a card default.'],
    ]}/><FoundationExample Example={ElevationExample} code={ElevationExampleSource} />
    <p>Elevation controls appearance only. A floating Surface is not a dialog or menu: use Modal, Drawer or ActionMenu for focus management, dismissal and positioning. Shadows take no layout space; leave room at container edges and do not clip them. In forced colors, the border remains the boundary.</p>
  </>
}

