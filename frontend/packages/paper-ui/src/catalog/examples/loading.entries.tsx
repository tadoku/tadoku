import * as stylex from '@stylexjs/stylex'
import { Loading, layout } from 'paper-ui'

export default function Example() {
  return (
    <div {...stylex.props(layout.stack)}>
      <p>Loading reading entries. Your saved entries will appear here.</p>
      <div {...stylex.props(layout.cluster)}>
        <span {...stylex.props(layout.stack)}>
          <Loading
            label="Loading reading entries, small example"
            size="small"
          />
          <span>Small</span>
        </span>
        <span {...stylex.props(layout.stack)}>
          <Loading label="Loading reading entries, default example" />
          <span>Default</span>
        </span>
        <span {...stylex.props(layout.stack)}>
          <Loading
            label="Loading reading entries, large example"
            size="large"
          />
          <span>Large</span>
        </span>
      </div>
    </div>
  )
}
