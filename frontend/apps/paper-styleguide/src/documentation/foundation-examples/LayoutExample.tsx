import * as stylex from '@stylexjs/stylex'
import { Button, Surface, typography } from 'paper-ui'
import { utilities } from '../../utilities.styles'

export default function ReadingSection() {
  return (
    <section {...stylex.props(utilities.tw_mx_auto, utilities.tw_grid, utilities.tw_w_full, utilities.tw_max_w_6xl, utilities.tw_grid_cols_1, utilities.tw_gap_6, utilities.tw_lg_grid_cols__minmax_0_2fr__minmax_0_1fr__)}>
      <Surface xstyle={[utilities.tw_min_w_0, utilities.tw_flex, utilities.tw_flex_col, utilities.tw_gap_4, utilities.tw_p_4, utilities.tw_md_p_6]}>
        <h2 {...stylex.props(typography.section, utilities.tw_m_0)}>September reading</h2>
        <p {...stylex.props(utilities.tw_m_0, utilities.tw_max_w_prose)}>Choose a reading session to continue.</p>
        <div {...stylex.props(utilities.tw_flex, utilities.tw_flex_wrap, utilities.tw_items_center, utilities.tw_gap_inline)}>
          <Button variant="outline">All languages</Button>
          <Button variant="outline">September 2026</Button>
          <Button>Log reading</Button>
        </div>
      </Surface>
      <aside {...stylex.props(utilities.tw_min_w_0, utilities.tw_flex, utilities.tw_flex_col, utilities.tw_gap_2, utilities.tw_py_4)}>
        <h2 {...stylex.props(typography.component, utilities.tw_m_0)}>Your reading goal</h2>
        <p {...stylex.props(utilities.tw_m_0)}>Read for 20 minutes on four days this week.</p>
        <p {...stylex.props(utilities.tw_m_0, utilities.tw_text_muted)}>Three sessions complete. One more to go.</p>
      </aside>
    </section>
  )
}
