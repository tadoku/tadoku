import * as stylex from '@stylexjs/stylex'
import { layout, typography } from 'paper-ui'
export default function ReadingPage() {
  return (
    <article {...stylex.props(layout.stack, layout.measure)}>
      <h2 {...stylex.props(typography.page)}>Your reading</h2>
      <p {...stylex.props(typography.body)}>Small sessions add up.</p>
      <section {...stylex.props(layout.stack)}>
        <h3 {...stylex.props(typography.section)}>Recent activity</h3>
        <p {...stylex.props(typography.body)}>You read 48 pages of コンビニ人間.</p>
        <p {...stylex.props(typography.metadata)}>Japanese · September 5</p>
      </section>
    </article>
  )
}
