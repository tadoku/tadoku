import * as stylex from '@stylexjs/stylex'
import { useQuery } from '@tanstack/react-query'
import { BookOpenIcon, LanguageIcon, TrophyIcon } from 'paper-ui/icons'
import { Link, useLoaderData } from '@tanstack/react-router'
import { IconTile, Surface, buttonStyles, linkStyles } from 'paper-ui'
import {
  border,
  breakpoints,
  color,
  fluid,
  font,
  fontSize,
  fontWeight,
  lineHeight,
  measure,
  size,
  space,
  tracking,
  type,
} from 'paper-ui/tokens.stylex'
import type { ReactNode } from 'react'
import { homeQueryOptions, phaseOf, type HomeData, type Phase } from '../lib/home'
import { formatDate, formatPeriod, numberFormat, pointsUnit } from '../lib/format'
import { describeGoal, useStoredGoal } from '../lib/goal'
import { useLinks } from '../lib/links'
import { useParticipation } from '../lib/registration'
import { MemberDashboard } from './MemberDashboard'
import { viewerQueryOptions, type Viewer } from '../lib/session'
import { ContestBoard } from './HomeContestBoard'


const styles = stylex.create({
  ledger: { minInlineSize: 0 },
  hero: {
    gap: { default: space[64], [breakpoints.xl]: space[32] },
    alignItems: 'start',
    display: 'grid',
    gridTemplateColumns: { default: 'minmax(0, 1fr) minmax(0, 1fr)', [breakpoints.md]: 'minmax(0, 1fr)' },
    paddingBlockEnd: space[8],
    paddingBlockStart: { default: space[8], [breakpoints.md]: 0 },
    paddingInlineEnd: space[5],
    paddingInlineStart: 0,
  },
  status: {
    marginInline: 0,
    color: color.statusWarning,
    fontSize: fontSize[14],
    fontWeight: fontWeight.semibold,
    marginBlockEnd: space[16],
    marginBlockStart: 0,
  },
  headline: {
    margin: 0,
    fontFamily: font.display,
    fontSize: {
      default: `clamp(${fontSize[32]}, ${fluid[3.5]}, ${fontSize[44]})`,
      [breakpoints.xl]: fontSize[36],
      [breakpoints.md]: `clamp(${fontSize[28]}, ${fluid[6.75]}, ${fontSize[44]})`,
    },
    letterSpacing: tracking.tightest,
    lineHeight: lineHeight.hero,
    maxInlineSize: { default: measure.headline, [breakpoints.md]: measure.title },
    textWrap: 'balance',
  },
  lead: {
    marginBlock: { default: space[24], [breakpoints.md]: space[20] },
    marginInline: 0,
    color: color.textInk,
    fontSize: { default: fontSize[17], [breakpoints.md]: fontSize[16] },
    lineHeight: lineHeight.open,
    maxInlineSize: measure.intro,
  },
  actions: { alignItems: 'center', columnGap: space[20], display: 'flex', flexWrap: 'wrap', rowGap: space[12] },
  fineprint: {
    marginInline: 0,
    color: color.textMuted,
    fontSize: type.metadataSize,
    lineHeight: lineHeight.relaxed,
    marginBlockEnd: 0,
    marginBlockStart: space[16],
  },
  goal: { marginInline: 0, fontSize: fontSize[15], lineHeight: lineHeight.relaxed, marginBlockEnd: 0, marginBlockStart: space[32], },
  goalTitle: {
    marginInline: 0,
    color: color.textMuted,
    fontFamily: font.body,
    fontSize: type.metadataSize,
    fontWeight: fontWeight.regular,
    marginBlockEnd: space[4],
    marginBlockStart: 0,
  },
  goalText: { margin: 0, color: color.textInk, fontWeight: fontWeight.semibold, },
  goalLink: { fontSize: fontSize[14], fontWeight: fontWeight.regular, marginInlineStart: space[12] },
  how: {
    borderBlockStartColor: color.ruleDefault,
    borderBlockStartStyle: 'solid',
    borderBlockStartWidth: border.static,
    marginBlockStart: { default: space[48], [breakpoints.md]: space[36] },
    paddingBlockStart: { default: space[28], [breakpoints.md]: space[24] },
  },
  howTitle: { marginInline: 0, fontFamily: font.display, fontSize: fontSize[22], marginBlockEnd: space[24], marginBlockStart: 0, },
  howSteps: {
    margin: 0,
    padding: 0,
    gap: { default: space[40], [breakpoints.md]: space[24] },
    listStyle: 'none',
    display: 'grid',
    gridTemplateColumns: { default: 'repeat(3, minmax(0, 1fr))', [breakpoints.md]: 'minmax(0, 1fr)' },
  },
  howStep: { alignContent: 'start', display: 'grid', justifyItems: 'start' },
  howIcon: { marginBlockEnd: space[14] },
  howStepTitle: { marginInline: 0, fontSize: fontSize[16], lineHeight: lineHeight.normal, marginBlockEnd: space[8], marginBlockStart: 0, },
  howStepText: {
    margin: 0,
    color: color.textInk,
    fontSize: fontSize[15],
    lineHeight: lineHeight.airy,
    maxInlineSize: { default: measure.narrow, [breakpoints.md]: measure.summary },
  },
  strip: {
    paddingInline: 0,
    alignItems: 'center',
    borderBlockStartColor: color.ruleDefault,
    borderBlockStartStyle: 'solid',
    borderBlockStartWidth: border.static,
    columnGap: { default: space[32], [breakpoints.xl]: space[24], [breakpoints.md]: space[12] },
    display: 'grid',
    gridTemplateColumns: {
      default: `${size[230]} minmax(0, 1fr) auto`,
      [breakpoints.xl]: `${size[200]} minmax(0, 1fr)`,
      [breakpoints.md]: 'minmax(0, 1fr)',
    },
    marginBlockStart: { default: space[32], [breakpoints.md]: space[28] },
    maxInlineSize: size[768],
    paddingBlockEnd: 0,
    paddingBlockStart: space[24],
    rowGap: { default: space[32], [breakpoints.xl]: space[16], [breakpoints.md]: space[12] },
  },
  stripTitle: { marginInline: 0, fontFamily: font.body, fontSize: fontSize[14], marginBlockEnd: space[4], marginBlockStart: 0, },
  stripText: { margin: 0, color: color.textMuted, fontSize: fontSize[14], lineHeight: lineHeight.airy, maxInlineSize: measure.aside, },
  stripPeriod: { color: color.textInk, fontSize: fontSize[17], fontWeight: fontWeight.semibold },
  stripLink: {
    gridColumn: { default: 3, [breakpoints.xl]: 2, [breakpoints.md]: 'auto' },
    fontSize: fontSize[14],
    justifySelf: { default: 'end', [breakpoints.xl]: 'start' },
  },
})


const guestHeadline = <>Read more.<br /> Listen more.<br /> Keep going together.</>

const promise: Record<Phase, { headline: ReactNode; lead: string }> = {
  open: { headline: guestHeadline, lead: 'Turn time with your languages into a habit. Join a friendly immersion contest, log what you read and listen to, and follow the community’s progress.' },
  live: { headline: guestHeadline, lead: 'Turn time with your languages into a habit. Log what you read and listen to, and follow the community’s progress in the current round.' },
  upcoming: { headline: 'Make time for the next round.', lead: 'Choose your languages now and find something you want to read or listen to. Start the next immersion round with the community.' },
  ended: { headline: 'Keep immersing between rounds.', lead: 'Contests add momentum, but the habit is yours year-round. Track what you read and listen to today, then bring that rhythm into the next official round.' },
  none: { headline: guestHeadline, lead: 'Turn time with your languages into a habit. Log what you read and listen to, and join an immersion contest when one suits you.' },
  unavailable: { headline: 'Your immersion can keep going.', lead: 'The contest board is temporarily unavailable. You can still log what you read and listen to today.' },
  loading: { headline: guestHeadline, lead: 'Turn time with your languages into a habit. Join a friendly immersion contest, log what you read and listen to, and follow the community’s progress.' },
}

export function HomePage() {
  const links = useLinks()
  const { data, dataUpdatedAt, isError, refetch, isFetching } = useQuery(homeQueryOptions())
  // The loader's year matches the dashboard queries it prefetched, even when the contest board fails.
  const { year } = useLoaderData({ from: '/' })
  // The fetch time travels with the dehydrated data, so server and client agree on the contest day.
  const today = new Date(dataUpdatedAt).toISOString().slice(0, 10)
  const phase = phaseOf(data, isError, today)
  const contest = data?.latestContest ?? null
  const joinable = contest && (phase === 'open' || phase === 'upcoming')
  const { data: viewer } = useQuery(viewerQueryOptions())
  const member = viewer?.status === 'signed-in' ? viewer : null

  return (
    <div>
      <section {...stylex.props(styles.hero)} aria-labelledby="home-title">
        {member
          ? <MemberPromise member={member} data={data} phase={phase} />
          : (
            <div>
              {phase === 'unavailable' && <p {...stylex.props(styles.status)}>Contest updates are delayed</p>}
              <h1 id="home-title" {...stylex.props(styles.headline)}>{promise[phase].headline}</h1>
              <p {...stylex.props(styles.lead)}>{promise[phase].lead}</p>
              <div {...stylex.props(styles.actions)}>
                {joinable
                  ? <a href={links.joinContest(contest.id)} {...stylex.props(...buttonStyles())}>Join {contest.title}</a>
                  : <a href={links.logs} {...stylex.props(...buttonStyles())}>Log activity</a>}
                <SetGoalLink />
              </div>
              {joinable && <p {...stylex.props(styles.fineprint)}>Free to join. Registration closes {formatDate(contest.registration_end)}.</p>}
              <DailyGoal />
            </div>
          )}

        <Surface as="section" elevation="showcase" inset="none" xstyle={styles.ledger} aria-labelledby="home-board-title">
          <ContestBoard data={data} phase={phase} today={today} retrying={isFetching} retry={() => void refetch()} />
        </Surface>
      </section>

      {member ? <MemberDashboard viewer={member} year={year} /> : (
        <section {...stylex.props(styles.how)} aria-labelledby="home-how-title">
          <h2 id="home-how-title" {...stylex.props(styles.howTitle)}>How Tadoku works</h2>
          <ol {...stylex.props(styles.howSteps)}>
            <HowStep icon={<LanguageIcon />} title="Choose your language">
              Join an official round or track immersion at your own pace between contests.
            </HowStep>
            <HowStep icon={<BookOpenIcon />} title="Read and listen">
              Books, manga, games, podcasts, video: whatever keeps you engaged.
            </HowStep>
            <HowStep icon={<TrophyIcon />} title="Log your progress">Record what you did and see where it counts.</HowStep>
          </ol>
        </section>
      )}

      {/* While a round is running the card covers it; the next round earns a row once it is scheduled. */}
      {phase !== 'loading' && phase !== 'unavailable' && (data?.nextContest || (phase !== 'open' && phase !== 'live')) && (
        <section {...stylex.props(styles.strip)} aria-labelledby="home-next-title">
          <div>
            <h2 id="home-next-title" {...stylex.props(styles.stripTitle)}>Next official round</h2>
            <p {...stylex.props(styles.stripText, styles.stripPeriod)}>{data?.nextContest ? formatPeriod(data.nextContest) : 'Not announced yet'}</p>
          </div>
          {data?.nextContest && <p {...stylex.props(styles.stripText)}>Registration closes {formatDate(data.nextContest.registration_end)}.</p>}
          <a {...stylex.props(linkStyles.base, styles.stripLink)} href={links.contests}>See all contests</a>
        </section>
      )}

    </div>
  )
}

function HowStep({ icon, title, children }: { icon: ReactNode; title: string; children: ReactNode }) {
  return (
    <li {...stylex.props(styles.howStep)}>
      <IconTile xstyle={styles.howIcon}>{icon}</IconTile>
      <h3 {...stylex.props(styles.howStepTitle)}>{title}</h3>
      <p {...stylex.props(styles.howStepText)}>{children}</p>
    </li>
  )
}

/** The goal is kept in the browser, so guests can set one too. Until one is saved, the hero's actions offer to set it. */
function SetGoalLink() {
  return useStoredGoal() ? null : <Link {...stylex.props(linkStyles.base)} to="/plan">Set a daily goal</Link>
}

function DailyGoal() {
  const goal = useStoredGoal()
  if (!goal) return null

  return (
    <section {...stylex.props(styles.goal)} aria-labelledby="home-goal-title">
      <h2 id="home-goal-title" {...stylex.props(styles.goalTitle)}>Daily goal</h2>
      <p {...stylex.props(styles.goalText)}>{describeGoal(goal)} <Link {...stylex.props(linkStyles.base, styles.goalLink)} to="/plan">Change goal</Link></p>
    </section>
  )
}

const ordinal = new Intl.PluralRules('en', { type: 'ordinal' })
const suffix: Record<string, string> = { one: 'st', two: 'nd', few: 'rd', other: 'th' }
const place = (rank: number) => `${rank}${suffix[ordinal.select(rank)]}`

/** A signed-in learner's hero: where they stand in the current round and the one action that moves it. */
function MemberPromise({ member, data, phase }: {
  member: Extract<Viewer, { status: 'signed-in' }>
  data: HomeData | undefined
  phase: Phase
}) {
  const links = useLinks()
  const contest = data?.latestContest
  const running = contest && (phase === 'open' || phase === 'live' || phase === 'upcoming')
  const entry = running ? data?.standings?.entries.find(row => row.user_id === member.userId) : undefined
  const { registered } = useParticipation(running ? contest.id : '')
  const joinable = running && !registered && (phase === 'open' || phase === 'upcoming')

  let headline: string
  let lead: string
  if (running && entry) {
    headline = `${place(entry.rank)} place`
    lead = `${numberFormat.format(entry.score)} ${pointsUnit(entry.score)} in ${contest.title}.`
  } else if (running && registered) {
    headline = `You’re in ${contest.title}`
    lead = phase === 'upcoming' ? `The round starts ${formatDate(contest.contest_start)}.` : 'Log what you read and listen to; every entry counts towards the round.'
  } else if (joinable) {
    headline = `${contest.title} is open`
    lead = `Registration closes ${formatDate(contest.registration_end)}.`
  } else {
    headline = `Welcome back, ${member.displayName}`
    lead = 'Log what you read and listen to today.'
  }

  return (
    <div>
      {phase === 'unavailable' && <p {...stylex.props(styles.status)}>Contest updates are delayed</p>}
      <h1 id="home-title" {...stylex.props(styles.headline)}>{headline}</h1>
      <p {...stylex.props(styles.lead)}>{lead}</p>
      {/* Log activity is already the header's action; the hero adds Join only while the member can still join. */}
      <div {...stylex.props(styles.actions)}>
        {joinable ? <a href={links.joinContest(contest.id)} {...stylex.props(...buttonStyles())}>Join {contest.title}</a> : null}
        <SetGoalLink />
      </div>
      <DailyGoal />
    </div>
  )
}
