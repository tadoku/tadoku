import * as stylex from '@stylexjs/stylex'
import { fontSize, lineHeight, space } from 'paper-ui/tokens.stylex'

/** Small layout utilities the playground composes inline. */
export const utilities = stylex.create({
  tw_flex: { display: 'flex' },
  tw_flex_wrap: { flexWrap: 'wrap' },
  tw_grid: { display: 'grid' },
  tw_items_center: { alignItems: 'center' },
  tw_justify_start: { justifyContent: 'flex-start' },
  tw_gap_2: { gap: space[8] },
  tw_gap_3: { gap: space[12] },
  tw_gap_4: { gap: space[16] },
  tw_gap_x_6: { columnGap: space[24] },
  tw_gap_y_2: { rowGap: space[8] },
  tw_mb_4: { marginBlockEnd: space[16] },
  tw_mt_6: { marginBlockStart: space[24] },
  tw_mt_8: { marginBlockStart: space[32] },
  tw_my_6: { marginBlock: space[24] },
  tw_p_0: { padding: 0 },
  tw_min_w_0: { minInlineSize: 0 },
  tw_shrink_0: { flexShrink: 0 },
  tw_text_sm: { fontSize: fontSize[14], lineHeight: lineHeight[20] },
  tw_whitespace_nowrap: { whiteSpace: 'nowrap' },
  tw_whitespace_pre_wrap: { whiteSpace: 'pre-wrap' },
  tw_rotate_180: { transform: 'rotate(180deg)' },
})
