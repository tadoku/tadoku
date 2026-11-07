import * as stylex from '@stylexjs/stylex'
import { buttonStyles, layout } from 'paper-ui'

export default function Example() {
  return (
    <div {...stylex.props(layout.cluster)}>
      <button type="button" {...stylex.props(...buttonStyles())}>
        Save log
      </button>
      <a href="#logs" {...stylex.props(...buttonStyles({ variant: 'outline' }))}>
        View logs
      </a>
    </div>
  )
}
