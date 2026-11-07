import { stack, Surface, typography } from 'paper-ui'

export default function ReadingSummary() {
  return (
    <Surface as="article" className={stack()}>
      <h2 className={typography({ role: 'component' })}>Your reading this week</h2>
      <p className={typography({ role: 'body' })}>Three sessions. 128 pages.</p>
      <p className={typography({ role: 'metadata' })}>Japanese · September 1–5</p>
      <a href="/patterns/logging">View reading</a>
    </Surface>
  )
}
