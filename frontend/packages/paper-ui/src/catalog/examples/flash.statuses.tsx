import { Flash } from 'paper-ui'

export default function Example() {
  return (
    <div className="paper-stack">
      <Flash title="Contest note">Only finished reading counts.</Flash>
      <Flash variant="success" title="Entry saved">
        48 pages added.
      </Flash>
      <Flash
        variant="warning"
        title="Check the date"
        action={<a href="#contest-window">Review dates</a>}
      >
        This entry is outside the contest window.
      </Flash>
      <Flash variant="danger" title="Could not save">
        Review the highlighted fields before trying again.
      </Flash>
      <p id="contest-window">
        Contest window: August 1–31. Choose a reading date inside this window.
      </p>
    </div>
  )
}
