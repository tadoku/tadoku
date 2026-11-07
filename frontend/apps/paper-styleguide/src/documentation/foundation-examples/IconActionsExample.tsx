import * as stylex from '@stylexjs/stylex'
import { useState } from 'react'
import { Button, iconSizes, layout } from 'paper-ui'
import { PlusIcon, XMarkIcon } from 'paper-ui/icons'
import { size } from 'paper-ui/tokens.stylex'

const styles = stylex.create({ iconAction: { minBlockSize: size[44], minInlineSize: size[44] } })

export default function ReadingActions() {
  const [open, setOpen] = useState(false)
  return (
    <div {...stylex.props(layout.stack)}>
      <div {...stylex.props(layout.cluster)}>
        <Button
          leadingIcon={<PlusIcon {...stylex.props(iconSizes.default)} />}
          onClick={() => setOpen(true)}
        >
          Add reading
        </Button>
        <Button
          variant="ghost"
          aria-label="Close reading entry"
          xstyle={styles.iconAction}
          onClick={() => setOpen(false)}
        >
          <XMarkIcon {...stylex.props(iconSizes.default)} aria-hidden="true" />
        </Button>
      </div>
      <p role="status">
        {open ? 'Reading entry opened in this demo.' : 'Reading entry closed.'}
      </p>
    </div>
  )
}
