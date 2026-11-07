import { accentRail, cn, stack, Surface, typography } from 'paper-ui'

export default function FeaturedReading() {
  return (
    <section className={stack()}>
      <h2
        className={cn(accentRail(), typography({ role: 'section' }), 'ps-4')}
      >
        Your next chapter
      </h2>
      <Surface as="article" accent className={stack()}>
        <h3 className={typography({ role: 'component' })}>コンビニ人間</h3>
        <p>48 pages read. Continue at your own pace.</p>
      </Surface>
    </section>
  )
}
