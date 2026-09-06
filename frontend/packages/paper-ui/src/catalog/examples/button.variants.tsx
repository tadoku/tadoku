import { Button, buttonClassName } from 'paper-ui'

export default function Example() {
  return (
    <div className="paper-cluster">
      <Button>Save log</Button>
      <Button variant="outline">Cancel</Button>
      <Button variant="ghost">More options</Button>
      <Button variant="link">Clear filters</Button>
      <Button variant="destructive">Delete log</Button>
      <a className={buttonClassName({ variant: 'outline' })} href="#logs">
        View logs
      </a>
    </div>
  )
}
