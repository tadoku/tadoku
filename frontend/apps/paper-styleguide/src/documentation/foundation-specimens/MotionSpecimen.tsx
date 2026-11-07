import { useState } from 'react'
import { Button } from 'paper-ui'
import { GuideCode, GuideTable } from '../FoundationGuide'
import { FoundationCanvas } from './FoundationPreview'

const durations = { quick: 'duration-quick', standard: 'duration-standard', deliberate: 'duration-deliberate' } as const

export function MotionSpecimen() {
  const [moved, setMoved] = useState(false)
  return <><FoundationCanvas><div className="grid gap-4" data-moved={moved}><div><Button variant="outline" onClick={() => setMoved(!moved)}>{moved ? 'Move markers back' : 'Compare motion'}</Button></div>{([['Quick feedback', 'quick', '120ms'], ['Standard transition', 'standard', '180ms'], ['Deliberate movement', 'deliberate', '240ms']] as const).map(([label, role, duration]) => <div key={role}><strong>{label} · {duration}</strong><div className="relative h-8 border border-rule bg-canvas" aria-hidden="true" ><span className={`absolute inset-y-1 start-1 w-5 bg-action transition-inset ease-out motion-reduce:transition-none [[data-moved=true]_&]:start-motion-end ${durations[role]}`}/></div><code>duration-{role}</code></div>)}</div></FoundationCanvas>
    <p>Activate the button to compare durations. Nothing moves automatically. With reduced motion enabled, markers change position immediately. Motion explains a change that already happened; it must never delay the result or be the only indication of status.</p>
    <GuideCode label="Transition with a reduced-motion override (JSX)" code={'<li\n  data-selected={selected}\n  className="bg-paper transition-colors duration-quick ease-out data-[selected=true]:bg-action-soft motion-reduce:transition-none"\n>'} />
    <p>The <code>duration-quick</code>, <code>duration-standard</code> and <code>duration-deliberate</code> utilities set duration only. They do not select properties, easing or provide a reduced-motion override. Add <code>motion-reduce:transition-none</code> yourself. Prefer specific color/opacity/transform transitions over <code>transition: all</code>.</p>
    <GuideTable label="Choosing a duration" columns={['Duration utility', 'Use', 'Avoid']} rows={[
      ['duration-quick · 120ms', 'Hover/pressed feedback and small emphasis changes.', 'Large spatial movement.'], ['duration-standard · 180ms', 'A small user-triggered reveal or position change.', 'Animating every list item when the page loads.'], ['duration-deliberate · 240ms', 'Rare larger movement that helps orientation.', 'Delaying a confirmation or blocking the next action.'],
    ]}/>
  </>
}

