import ElevationExample from '../foundation-examples/ElevationExample'
import ElevationExampleSource from '../foundation-examples/ElevationExample.tsx?raw'
import { Surface } from 'paper-ui'
import { GuideTable } from '../FoundationGuide'
import { FoundationCanvas, FoundationExample } from './FoundationPreview'

export function ElevationSpecimen() {
  return <><FoundationCanvas><div className="foundation-guide__grid">{(['flat', 'floating', 'showcase'] as const).map(elevation => <Surface key={elevation} elevation={elevation} className="paper-stack"><h3 className="paper-type-component">{elevation}</h3><p>{elevation === 'flat' ? 'A reading summary in normal flow.' : elevation === 'floating' ? 'A transient layer, such as a menu.' : 'A rare visual study or featured demonstration.'}</p></Surface>)}</div></FoundationCanvas>
    <GuideTable label="Elevation roles" columns={['Surface prop / class', 'Offset', 'Contract']} rows={[
      ['flat / paper-elevation-flat', 'none', 'Default. A border supplies separation; no shadow.'], ['floating / paper-elevation-floating', '3px right and down; no blur', 'Transient layer. Stronger perimeter in both themes.'], ['showcase / paper-elevation-showcase', '5px right and down; no blur', 'Rare deliberate emphasis; not a card default.'],
    ]}/><FoundationExample Example={ElevationExample} code={ElevationExampleSource} />
    <p>Elevation controls appearance only. A floating Surface is not a dialog or menu: use Modal, Drawer or ActionMenu for focus management, dismissal and positioning. Shadows take no layout space; leave room at container edges and do not clip them. In forced colors, the border remains the boundary.</p>
  </>
}

