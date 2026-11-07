import * as stylex from '@stylexjs/stylex'
import { size } from 'paper-ui/tokens.stylex'

const styles = stylex.create({
  region: { maxInlineSize: '100%', overflowX: 'auto' },
  table: { minInlineSize: size[576] },
})

export default function ReadingTotals() {
  return (
    <div
      role="region"
      aria-label="Reading totals"
      tabIndex={0}
      {...stylex.props(styles.region)}
    >
      <table {...stylex.props(styles.table)}>
        <caption>Reading totals</caption>
        <thead>
          <tr>
            <th scope="col">Language</th>
            <th scope="col">Pages</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <th scope="row">Japanese</th>
            <td>96</td>
          </tr>
        </tbody>
      </table>
    </div>
  )
}
