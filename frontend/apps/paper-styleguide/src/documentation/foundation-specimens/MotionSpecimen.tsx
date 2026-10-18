import { useState, type CSSProperties } from 'react'
import { Button } from 'paper-ui'
import { GuideCode, GuideTable } from '../FoundationGuide'
import { FoundationCanvas } from './FoundationPreview'

export function MotionSpecimen() {
  const [moved, setMoved] = useState(false)
  return <><FoundationCanvas><div className="foundation-motion-demo" data-moved={moved}><div><Button variant="outline" onClick={() => setMoved(!moved)}>{moved ? 'Move markers back' : 'Compare motion'}</Button></div>{([['Quick feedback', 'quick', '120ms'], ['Standard transition', 'standard', '180ms'], ['Deliberate movement', 'deliberate', '240ms']] as const).map(([label, role, duration]) => <div key={role}><strong>{label} · {duration}</strong><div className="foundation-motion-demo__track" aria-hidden="true" style={{ '--demo-duration': `var(--paper-motion-${role})` } as CSSProperties}><span className="foundation-motion-demo__marker"/></div><code>--paper-motion-{role}</code></div>)}</div></FoundationCanvas>
    <p>Activate the button to compare durations. Nothing moves automatically. With reduced motion enabled, markers change position immediately. Motion explains a change that already happened; it must never delay the result or be the only indication of status.</p>
    <GuideCode label="Transition with a reduced-motion override (CSS)" code={'.reading-highlight {\n  background: var(--paper-color-surface-paper);\n  transition: background-color var(--paper-motion-quick) ease-out;\n}\n.reading-highlight[data-selected="true"] {\n  background: var(--paper-color-action-soft);\n}\n@media (prefers-reduced-motion: reduce) {\n  .reading-highlight { transition: none; }\n}'} />
    <p>Public <code>paper-motion-quick</code>, <code>paper-motion-standard</code> and <code>paper-motion-deliberate</code> classes set duration only. They do not select properties, easing or provide a reduced-motion override. When writing custom animation, supply that override yourself. Prefer specific color/opacity/transform transitions over <code>transition: all</code>.</p>
    <GuideTable label="Choosing a duration" columns={['CSS duration class', 'Use', 'Avoid']} rows={[
      ['paper-motion-quick · 120ms', 'Hover/pressed feedback and small emphasis changes.', 'Large spatial movement.'], ['paper-motion-standard · 180ms', 'A small user-triggered reveal or position change.', 'Animating every list item when the page loads.'], ['paper-motion-deliberate · 240ms', 'Rare larger movement that helps orientation.', 'Delaying a confirmation or blocking the next action.'],
    ]}/>
  </>
}

