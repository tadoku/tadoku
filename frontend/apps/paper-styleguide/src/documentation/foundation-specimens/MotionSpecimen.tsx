import * as stylex from '@stylexjs/stylex'
import { useState } from 'react'
import { Button, motionStyles } from 'paper-ui'
import { GuideCode, GuideTable } from '../FoundationGuide'
import { FoundationCanvas } from './FoundationPreview'
import { foundationStyles } from '../foundation-guide.styles'

export function MotionSpecimen() {
  const [moved, setMoved] = useState(false)
  return <><FoundationCanvas><div {...stylex.props(foundationStyles.foundationMotionDemo)}><div><Button variant="outline" onClick={() => setMoved(!moved)}>{moved ? 'Move markers back' : 'Compare motion'}</Button></div>{([['Quick feedback', 'quick', '120ms'], ['Standard transition', 'standard', '180ms'], ['Deliberate movement', 'deliberate', '240ms']] as const).map(([label, role, duration]) => <div key={role}><strong>{label} · {duration}</strong><div {...stylex.props(foundationStyles.foundationMotionDemo__track)} aria-hidden="true"><span {...stylex.props(foundationStyles.foundationMotionDemo__marker, motionStyles[role], moved && foundationStyles.foundationMotionDemo__markerMoved)}/></div><code>motionStyles.{role}</code></div>)}</div></FoundationCanvas>
    <p>Activate the button to compare durations. Nothing moves automatically. With reduced motion enabled, markers change position immediately. Motion explains a change that already happened; it must never delay the result or be the only indication of status.</p>
    <GuideCode label="Transition with a reduced-motion override (StyleX)" code={'const styles = stylex.create({\n  highlight: {\n    backgroundColor: color.surfacePaper,\n    transitionProperty: "background-color",\n    transitionTimingFunction: "ease-out",\n  },\n  selected: { backgroundColor: color.actionSoft },\n});\n\n<div {...stylex.props(styles.highlight, motionStyles.quick, selected && styles.selected)} />'} />
    <p><code>motionStyles.quick</code>, <code>motionStyles.standard</code> and <code>motionStyles.deliberate</code> set the duration and drop it to zero under reduced motion. They do not select properties or easing. For a duration in your own style, use the <code>motion</code> tokens and supply the reduced-motion override yourself. Prefer specific color/opacity/transform transitions over <code>transition: all</code>.</p>
    <GuideTable label="Choosing a duration" columns={['StyleX export', 'Use', 'Avoid']} rows={[
      ['motionStyles.quick · 120ms', 'Hover/pressed feedback and small emphasis changes.', 'Large spatial movement.'], ['motionStyles.standard · 180ms', 'A small user-triggered reveal or position change.', 'Animating every list item when the page loads.'], ['motionStyles.deliberate · 240ms', 'Rare larger movement that helps orientation.', 'Delaying a confirmation or blocking the next action.'],
    ]}/>
  </>
}

