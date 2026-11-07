import * as stylex from '@stylexjs/stylex'
import { accentRail, layout, Surface, typography } from 'paper-ui'
import { space } from 'paper-ui/tokens.stylex'

const styles = stylex.create({ heading: { paddingInlineStart: space[16] } })

export default function FeaturedReading() {
  return (
    <section {...stylex.props(layout.stack)}>
      <h2
        {...stylex.props(accentRail.base, typography.section, styles.heading)}
      >
        Your next chapter
      </h2>
      <Surface as="article" accent xstyle={[layout.stack]}>
        <h3 {...stylex.props(typography.component)}>コンビニ人間</h3>
        <p>48 pages read. Continue at your own pace.</p>
      </Surface>
    </section>
  )
}
