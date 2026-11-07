import * as stylex from '@stylexjs/stylex'
import { Button, iconSizes, layout } from 'paper-ui'
import { size } from 'paper-ui/tokens.stylex'
import { PlusIcon, XMarkIcon } from 'paper-ui/icons'

const styles = stylex.create({
  iconOnly: { minBlockSize: size[44], minInlineSize: size[44] },
})

export default function Example() {
  return (
    <div {...stylex.props(layout.stack)}>
      <div {...stylex.props(layout.cluster)}>
        <Button leadingIcon={<PlusIcon {...stylex.props(iconSizes.default)} />}>
          Add reading
        </Button>
        <Button
          variant="ghost"
          aria-label="Close entry"
          xstyle={styles.iconOnly}
        >
          <XMarkIcon {...stylex.props(iconSizes.default)} aria-hidden="true" />
        </Button>
        <Button disabled aria-describedby="archive-reason">
          Submit to contest
        </Button>
      </div>
      <p id="archive-reason">
        This contest is archived and no longer accepts entries.
      </p>
    </div>
  )
}
