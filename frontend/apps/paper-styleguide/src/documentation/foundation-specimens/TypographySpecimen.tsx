import TypographyExample from '../foundation-examples/TypographyExample'
import TypographyExampleSource from '../foundation-examples/TypographyExample.tsx?raw'
import { GuideTable } from '../FoundationGuide'
import { FoundationCanvas, FoundationExample } from './FoundationPreview'

const typeRoles = [
  ['display', 'A season of reading', '48 / 52', '700', 'Rare landing-page display; allow wrapping.'], ['page', 'Your reading', '32 / 38', '700', 'One page title.'], ['section', 'Recent activity', '24 / 30', '700', 'Major page sections.'], ['component', 'September Japanese', '18 / 24', '700', 'Component and card headings.'], ['body', 'Make time for a few more pages.', '16 / 26; compact 14 / 22', '400', 'Reading copy and interface explanations.'], ['label', 'Reading language', '12 / 16', '700', 'Short visible labels. No forced uppercase.'], ['metadata', 'Updated September 5', '12 / 18', '400', 'Supporting context; never essential instructions alone.'],
] as const
export function TypographySpecimen() {
  return <>
    <FoundationCanvas><ul className="foundation-type-list">{typeRoles.map(([role, text]) => <li key={role}><span className={`paper-type-${role}`}>{text}</span><code>paper-type-{role}</code></li>)}</ul></FoundationCanvas>
    <GuideTable label="Type roles at a 16px root" columns={['CSS class', 'Size / line height (px)', 'Weight', 'Use']} rows={typeRoles.map(([role, , size, weight, use]) => [`paper-type-${role}`, size, weight, use])} />
    <FoundationExample Example={TypographyExample} code={TypographyExampleSource} />
    <p>The full stylesheet includes the bundled fonts: Merriweather ships at 700; Open Sans at 400, 600 and 700. Unshipped italics or weights may be synthesized by the browser. Japanese and other scripts use available fallback fonts. Test realistic multilingual titles rather than assuming every script shares Latin metrics.</p><p>Type classes do not choose heading levels. Keep DOM headings sequential, use body text for instructions and let text wrap. The fixed scale uses rem units and follows browser zoom.</p>
  </>
}

