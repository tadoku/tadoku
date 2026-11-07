import * as stylex from '@stylexjs/stylex'
import { size } from 'paper-ui/tokens.stylex'
import { Surface, layout } from 'paper-ui'

const styles = stylex.create({
  column: { maxInlineSize: size[512] },
  flush: { margin: 0 },
})

export default function Example() {
  return (
    <div {...stylex.props(layout.stack, styles.column)}>
      {(['flat', 'floating', 'showcase'] as const).map(elevation => (
        <Surface
          key={elevation}
          as="article"
          elevation={elevation}
          {...stylex.props(layout.stack)}
        >
          <h3 {...stylex.props(styles.flush)}>
            {elevation === 'flat'
              ? 'Flat'
              : elevation === 'floating'
              ? 'Floating'
              : 'Showcase'}{' '}
            summary
          </h3>
          <p {...stylex.props(styles.flush)}>
            August Japanese: 1,240 pages across 18 entries.
          </p>
        </Surface>
      ))}
      <Surface as="article" accent {...stylex.props(layout.stack)}>
        <h3 {...stylex.props(styles.flush)}>Flat with accent rail</h3>
        <p {...stylex.props(styles.flush)}>
          The same reading summary with an independent accent.
        </p>
      </Surface>
    </div>
  )
}
