import { Surface } from 'paper-ui'

export default function FeaturedReading() {
  return (
    <section className="paper-stack">
      <h2
        className="paper-accent-rail paper-type-section"
        style={{ paddingInlineStart: 'var(--paper-space-4)' }}
      >
        Your next chapter
      </h2>
      <Surface as="article" accent className="paper-stack">
        <h3 className="paper-type-component">コンビニ人間</h3>
        <p>48 pages read. Continue at your own pace.</p>
      </Surface>
    </section>
  )
}
