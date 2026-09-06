import { useState } from 'react'
import { Button } from 'paper-ui'
import { PlusIcon, XMarkIcon } from 'paper-ui/icons'

export default function ReadingActions() {
  const [open, setOpen] = useState(false)
  return (
    <div className="paper-stack">
      <div className="paper-cluster">
        <Button
          leadingIcon={<PlusIcon className="paper-icon-default" />}
          onClick={() => setOpen(true)}
        >
          Add reading
        </Button>
        <Button
          variant="ghost"
          aria-label="Close reading entry"
          style={{ minWidth: '2.75rem', minHeight: '2.75rem' }}
          onClick={() => setOpen(false)}
        >
          <XMarkIcon className="paper-icon-default" aria-hidden="true" />
        </Button>
      </div>
      <p role="status">
        {open ? 'Reading entry opened in this demo.' : 'Reading entry closed.'}
      </p>
    </div>
  )
}
