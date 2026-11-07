import * as stylex from '@stylexjs/stylex'
import { Surface, typography } from 'paper-ui'
import { utilities } from '../../utilities.styles'

export default function ReadingSummary() {
  return (
    <Surface xstyle={[utilities.tw_flex, utilities.tw_flex_col, utilities.tw_gap_4, utilities.tw_p_6]}>
      <h2 {...stylex.props(typography.component, utilities.tw_m_0)}>September reading</h2>
      <p {...stylex.props(utilities.tw_m_0)}>24px inside the surface. 16px between these items.</p>
    </Surface>
  )
}
