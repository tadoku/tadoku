import { Button } from 'paper-ui'
import { PlusIcon, XMarkIcon, iconClassName } from 'paper-ui/icons'

export default function Example() {
  return (
    <div className="paper-stack">
      <div className="paper-cluster">
        <Button leadingIcon={<PlusIcon className={iconClassName()} />}>
          Add reading
        </Button>
        <Button
          variant="ghost"
          aria-label="Close entry"
          style={{ minWidth: '2.75rem', minHeight: '2.75rem' }}
        >
          <XMarkIcon className={iconClassName()} aria-hidden="true" />
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
