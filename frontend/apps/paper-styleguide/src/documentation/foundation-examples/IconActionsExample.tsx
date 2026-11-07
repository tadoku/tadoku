import { useState } from 'react'
import { Button, cluster, stack } from 'paper-ui'
import { iconClassName, PlusIcon, XMarkIcon } from 'paper-ui/icons'

export default function ReadingActions() {
  const [open, setOpen] = useState(false)
  return (
    <div className={stack()}>
      <div className={cluster()}>
        <Button
          leadingIcon={<PlusIcon className={iconClassName()} />}
          onClick={() => setOpen(true)}
        >
          Add reading
        </Button>
        <Button
          variant="ghost"
          aria-label="Close reading entry"
          className="min-h-target min-w-target"
          onClick={() => setOpen(false)}
        >
          <XMarkIcon className={iconClassName()} aria-hidden="true" />
        </Button>
      </div>
      <p role="status">
        {open ? 'Reading entry opened in this demo.' : 'Reading entry closed.'}
      </p>
    </div>
  )
}
