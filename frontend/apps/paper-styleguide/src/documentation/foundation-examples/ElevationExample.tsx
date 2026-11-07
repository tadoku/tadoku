import * as stylex from '@stylexjs/stylex'
import { layout, Surface, typography } from 'paper-ui'

export default function Summary() {
  return (
    <Surface as="article" elevation="flat" xstyle={[layout.stack]}>
      <h2 {...stylex.props(typography.component)}>September reading</h2>
      <p>128 pages across three sessions.</p>
    </Surface>
  )
}
