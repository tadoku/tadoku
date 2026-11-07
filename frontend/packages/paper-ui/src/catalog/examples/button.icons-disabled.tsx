import { Button, cluster, stack } from 'paper-ui'
import { PlusIcon, XMarkIcon, iconClassName } from 'paper-ui/icons'

export default function Example() {
  return (
    <div className={stack()}>
      <div className={cluster()}>
        <Button leadingIcon={<PlusIcon className={iconClassName()} />}>
          Add reading
        </Button>
        <Button
          variant="ghost"
          aria-label="Close entry"
          className="min-h-target min-w-target"
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
