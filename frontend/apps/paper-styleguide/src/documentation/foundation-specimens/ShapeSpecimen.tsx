import RailExample from '../foundation-examples/RailExample'
import RailExampleSource from '../foundation-examples/RailExample.tsx?raw'
import { Flash, Surface } from 'paper-ui'
import { GuideTable } from '../FoundationGuide'
import { FoundationCanvas, FoundationExample } from './FoundationPreview'

export function ShapeSpecimen() {
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

