import * as stylex from '@stylexjs/stylex'
import { linkStyles } from 'paper-ui'
import { color, fontSize, fontWeight, lineHeight, radius, size, space, tracking } from 'paper-ui/tokens.stylex'
import { standingsLayout } from './page-styles'


const styles = stylex.create({
  rank: { boxSizing: "border-box", color: color.textMuted },
  podium: {
    borderRadius: radius.round,
    paddingInline: space[4],
    placeItems: 'center',
    blockSize: size[28],
    display: 'inline-grid',
    fontVariantNumeric: 'tabular-nums',
    fontWeight: fontWeight.bold,
    minInlineSize: size[28],
  },
  gold: { backgroundColor: color.podiumGoldFill, color: color.podiumGoldInk },
  silver: { backgroundColor: color.podiumSilverFill, color: color.podiumSilverInk },
  bronze: { backgroundColor: color.podiumBronzeFill, color: color.podiumBronzeInk },
  person: { alignItems: 'baseline', columnGap: space[8], display: 'inline-flex', flexWrap: 'wrap', rowGap: space[4] },
  personLink: {
    color: { default: color.textInk, ':hover': color.textLink },
    overflowWrap: 'anywhere',
    textDecorationColor: { default: color.ruleDefault, ':hover': 'currentColor' },
  },
  /** The viewer's row: an accent-soft band and a filled You pill, not a rail and a boxed tag. */
  you: {
    borderRadius: radius.pill,
    paddingBlock: space[1],
    paddingInline: space[8],
    backgroundColor: color.actionDefault,
    color: color.actionText,
    display: 'inline-block',
    fontSize: fontSize[12],
    fontWeight: fontWeight.bold,
    letterSpacing: tracking.wide,
    lineHeight: lineHeight.normal,
    verticalAlign: space[1],
  },
})

const podiums = [styles.gold, styles.silver, styles.bronze]

/** A rank, on a gold, silver or bronze disc for the first three places. */
export function Rank({ rank, tie = false }: { rank: number; tie?: boolean }) {
  return (
    <span {...stylex.props(styles.rank, rank <= 3 && [styles.podium, podiums[rank - 1]])} data-podium={rank <= 3 ? rank : undefined}>
      {tie ? <abbr title="Tied rank" {...stylex.props(standingsLayout.plainAbbr)}>T</abbr> : null}{rank}
    </span>
  )
}

/** A participant's name linked to their profile, with the viewer's You pill; the link covers its table row. */
export function Participant({ name, href, you, rowLink }: { name: string; href: string; you: boolean; rowLink: stylex.StyleXStyles }) {
  return (
    <span {...stylex.props(styles.person)}>
      <a {...stylex.props(linkStyles.base, styles.personLink, rowLink)} href={href} data-participant="">{name}</a>
      {you ? <span {...stylex.props(styles.you)}>You</span> : null}
    </span>
  )
}

