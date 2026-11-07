import * as stylex from '@stylexjs/stylex'
import { breakpoints, color, density, measure, size, space } from 'paper-ui/tokens.stylex'

/** Small layout utilities the catalogue composes inline. */
export const utilities = stylex.create({
  tw_flex: { display: 'flex' },
  tw_flex_col: { flexDirection: 'column' },
  tw_flex_wrap: { flexWrap: 'wrap' },
  tw_grid: { display: 'grid' },
  tw_grid_cols_1: { gridTemplateColumns: 'repeat(1, minmax(0, 1fr))' },
  tw_lg_grid_cols__minmax_0_2fr__minmax_0_1fr__: {
    gridTemplateColumns: { default: null, [breakpoints.wide]: 'minmax(0, 2fr) minmax(0, 1fr)' },
  },
  tw_items_center: { alignItems: 'center' },
  tw_gap_2: { gap: space[8] },
  tw_gap_4: { gap: space[16] },
  tw_gap_6: { gap: space[24] },
  tw_gap_inline: { gap: density.inlineGap },
  tw_m_0: { margin: 0 },
  tw_mx_auto: { marginInline: 'auto' },
  tw_max_w_6xl: { maxInlineSize: size[1152] },
  tw_max_w_prose: { maxInlineSize: measure.body },
  tw_min_w_0: { minInlineSize: 0 },
  tw_w_full: { width: '100%' },
  tw_p_4: { padding: space[16] },
  tw_p_6: { padding: space[24] },
  tw_md_p_6: { padding: { default: null, [breakpoints.aboveMd]: space[24] } },
  tw_py_4: { paddingBlock: space[16] },
  tw_text_muted: { color: color.textMuted },
  tw_bg_paper: { backgroundColor: color.surfacePaper },
})
