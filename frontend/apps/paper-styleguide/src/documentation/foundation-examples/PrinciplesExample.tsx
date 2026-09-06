import { Surface } from 'paper-ui'

export default function ReadingSummary() {
  return (
    <Surface as="article" className="paper-stack">
      <h2 className="paper-type-component">Your reading this week</h2>
      <p className="paper-type-body">Three sessions. 128 pages.</p>
      <p className="paper-type-metadata">Japanese · September 1–5</p>
      <a href="/patterns/logging">View reading</a>
    </Surface>
  )
}
