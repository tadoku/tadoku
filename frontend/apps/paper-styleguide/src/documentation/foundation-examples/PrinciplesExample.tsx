import * as stylex from '@stylexjs/stylex'
import { layout, Surface, typography } from 'paper-ui'

export default function ReadingSummary() {
  return (
    <Surface as="article" xstyle={[layout.stack]}>
      <h2 {...stylex.props(typography.component)}>Your reading this week</h2>
      <p {...stylex.props(typography.body)}>Three sessions. 128 pages.</p>
      <p {...stylex.props(typography.metadata)}>Japanese · September 1–5</p>
      <a href="/patterns/logging">View reading</a>
    </Surface>
  )
}
