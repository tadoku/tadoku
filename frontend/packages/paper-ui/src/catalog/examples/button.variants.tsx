import * as stylex from '@stylexjs/stylex'
import { Button, buttonStyles, layout } from 'paper-ui'

export default function Example() {
  return (
    <div {...stylex.props(layout.cluster)}>
      <Button>Save log</Button>
      <Button variant="outline">Cancel</Button>
      <Button variant="ghost">More options</Button>
      <Button variant="link">Clear filters</Button>
      <Button variant="destructive">Delete log</Button>
      <a {...stylex.props(...buttonStyles({ variant: 'outline' }))} href="#logs">
        View logs
      </a>
    </div>
  )
}
