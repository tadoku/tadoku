import PrinciplesExample from '../foundation-examples/PrinciplesExample'
import PrinciplesExampleSource from '../foundation-examples/PrinciplesExample.tsx?raw'
import { GuideTable } from '../FoundationGuide'
import { FoundationExample } from './FoundationPreview'

export function PrinciplesSpecimen() {
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
