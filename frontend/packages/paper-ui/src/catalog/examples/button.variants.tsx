import { Button, buttonClassName, cluster } from 'paper-ui'

export default function Example() {
  return (
    <div className={cluster()}>
      <Button>Save log</Button>
      <Button variant="outline">Cancel</Button>
      <Button variant="ghost">More options</Button>
      <Button variant="link">Clear filters</Button>
      <Button variant="destructive">Delete log</Button>
      <Button variant="destructive-ghost">Remove</Button>
      <a className={buttonClassName({ variant: 'outline' })} href="#logs">
        View logs
      </a>
    </div>
  )
}
