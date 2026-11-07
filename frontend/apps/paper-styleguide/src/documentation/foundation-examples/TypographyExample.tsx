import { cn, measure, stack, typography } from 'paper-ui'
export default function ReadingPage() {
  return (
    <article className={cn(stack(), measure())}>
      <h2 className={typography({ role: 'page' })}>Your reading</h2>
      <p className={typography({ role: 'body' })}>Small sessions add up.</p>
      <section className={stack()}>
        <h3 className={typography({ role: 'section' })}>Recent activity</h3>
        <p className={typography({ role: 'body' })}>You read 48 pages of コンビニ人間.</p>
        <p className={typography({ role: 'metadata' })}>Japanese · September 5</p>
      </section>
    </article>
  )
}
