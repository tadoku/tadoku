import { useQuery } from '@tanstack/react-query'
import { useLoaderData } from '@tanstack/react-router'
import { Button, EmptyState, Loading, SectionHeading, StatGrid, Surface, Table, buttonClassName } from 'paper-ui'
import type { ReactNode } from 'react'
import { nextContestQuery, useHomeData, type HomeData } from '../lib/home'
import { contestEntryQuery } from '../lib/leaderboard'
import { Person, Rank, currentRow, personLink } from './StandingsParts'
import { formatDate, formatPeriod, numberFormat, scoreFormat } from '../lib/format'
import { useLinks } from '../lib/links'
import { useParticipation } from '../lib/registration'
import { MemberActivity, MemberDashboard } from './MemberDashboard'
import { viewerQueryOptions } from '../lib/session'
import { AppLink } from './router-link'

const dayNumber = (from: string, to: string) => Math.round((Date.parse(to) - Date.parse(from)) / 86_400_000) + 1

type Phase = 'open' | 'live' | 'upcoming' | 'ended' | 'none' | 'unavailable' | 'loading'

function phaseOf(data: HomeData | undefined, failed: boolean, today: string): Phase {
  if (failed) return 'unavailable'
  if (!data) return 'loading'
  const contest = data.latestContest
  if (!contest) return 'none'
  if (today < contest.contest_start) return 'upcoming'
  if (today > contest.contest_end) return 'ended'
  return today <= contest.registration_end ? 'open' : 'live'
}

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

const hero = [
  'grid grid-cols-halves items-start gap-16 py-2 max-lg:gap-8',
  'max-tablet:grid-cols-single max-tablet:gap-8 max-tablet:pt-0',
].join(' ')
const status = 'mb-4 mt-0 text-sm font-semibold text-warning'
const headline = 'm-0 max-w-measure-hero text-balance font-display text-hero leading-120 tracking-hero'
const lead = 'my-6 max-w-measure-sm text-lead leading-175 text-ink max-tablet:my-5 max-tablet:text-base'
const actions = 'flex flex-wrap items-center gap-x-5 gap-y-3'
const step = 'grid content-start'
const stepTitle = 'mb-1 mt-0 text-base leading-150'
const stepText = 'm-0 max-w-measure-how text-ui leading-170 text-ink'
const strip = [
  'mt-8 grid max-w-record grid-cols-strip items-center gap-8 border-t border-rule pt-6',
  'max-lg:grid-cols-strip-narrow max-lg:gap-x-6 max-lg:gap-y-4',
  'max-tablet:mt-7 max-tablet:grid-cols-single max-tablet:gap-3',
].join(' ')
const stripText = 'm-0 max-w-measure-strip leading-170'
const stripLink = [
  'col-start-3 justify-self-end text-sm text-link hover:text-link-hover max-lg:col-start-2 max-lg:justify-self-start max-tablet:col-auto',
].join(' ')
const ledgerHeader = 'm-0 border-b border-rule px-ledger-inset py-5.5 max-xs:py-4.5'
// The card frames the table, so the table's own top and bottom rules would double its edges.
export const cardTable = 'border-y-0 [&_tbody_tr:last-child>*]:border-b-0'
const message = [
  'px-ledger-inset py-6 [&_h3]:mb-3 [&_h3]:mt-0 [&_h3]:text-lead [&_h3]:leading-150',
  '[&>p]:mb-5 [&>p]:mt-0 [&>p]:max-w-measure-message [&>p]:leading-175 [&>p:last-child]:mb-0',
].join(' ')

export function HomePage() {
  const links = useLinks()
  const { data, dataUpdatedAt, isError, refetch, isFetching } = useHomeData()
  const next = useQuery(nextContestQuery()).data
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
    <div className="home-page">
      <section className={hero} aria-labelledby="home-title">
        {member
          ? (
            <div className="min-w-0">
              <MemberActivity viewer={member} year={year} />
            </div>
          )
          : (
            <div>
              {phase === 'unavailable' && <p className={status}>Contest updates are delayed</p>}
              <h1 id="home-title" className={headline}>{promise[phase].headline}</h1>
              <p className={lead}>{promise[phase].lead}</p>
              <div className={actions}>
                {joinable
                  ? <AppLink href={links.joinContest(contest.id)} className={buttonClassName()}>Join {contest.title}</AppLink>
                  : <AppLink href={links.logs} className={buttonClassName()}>Log activity</AppLink>}
              </div>
              {joinable && <p className="mb-0 mt-4 text-metadata leading-160 text-muted">Free to join. Registration closes {formatDate(contest.registration_end)}.</p>}
            </div>
          )}

        <Surface as="section" elevation="showcase" className="me-1.25 min-w-0 p-0" aria-labelledby="home-board-title">
          <ContestBoard data={data} phase={phase} today={today} retrying={isFetching} retry={() => void refetch()} />
        </Surface>
        {member ? <MemberDashboard viewer={member} year={year} boardContestId={contest?.id} className="col-span-full mt-8" /> : null}
      </section>

      {member ? null : (
        <section className="mt-12 border-t border-rule pt-7 max-tablet:mt-9 max-tablet:pt-6" aria-labelledby="home-how-title">
          <h2 id="home-how-title" className="mb-6 mt-0 font-display text-heading">How Tadoku works</h2>
          <ol className="m-0 grid list-none grid-cols-thirds gap-10 p-0 max-tablet:grid-cols-single max-tablet:gap-6">
            <li className={step}>
              <h3 className={stepTitle}>Choose your language</h3><p className={stepText}>Join an official round or track immersion at your own pace between contests.</p>
            </li>
            <li className={step}>
              <h3 className={stepTitle}>Read and listen</h3><p className={stepText}>Books, manga, games, podcasts, video: whatever keeps you engaged.</p>
            </li>
            <li className={step}>
              <h3 className={stepTitle}>Log your progress</h3><p className={stepText}>Record what you did and see where it counts.</p>
            </li>
          </ol>
        </section>
      )}

      {/* While a round is running the card covers it; the next round earns a row once it is scheduled. */}
      {phase !== 'loading' && next && ((next.ok && next.data) || (phase !== 'open' && phase !== 'live')) && (
        <section className={strip} aria-labelledby="home-next-title">
          <div>
            <h2 id="home-next-title" className="mb-1 mt-0 font-body text-sm">Next official round</h2>
            <p className={`${stripText} text-lead font-semibold text-ink`}>{next.ok ? (next.data ? formatPeriod(next.data) : 'Not announced yet') : 'Temporarily unavailable'}</p>
          </div>
          {next.ok && next.data && <p className={`${stripText} text-sm text-muted`}>Registration closes {formatDate(next.data.registration_end)}.</p>}
          <AppLink className={stripLink} href={links.contests}>See all contests</AppLink>
        </section>
      )}

    </div>
  )
}


/** The viewer's place on the round's board at any rank; undefined when they aren't on it or it can't be read. */
function useViewerEntry(contestId: string | undefined, userId: string | undefined) {
  const entry = useQuery({ ...contestEntryQuery(contestId ?? '', userId ?? ''), enabled: Boolean(contestId && userId) }).data
  return entry?.ok ? entry.data ?? undefined : undefined
}

function ContestBoard({ data, phase, today, retrying, retry }: {
  data: HomeData | undefined
  phase: Phase
  today: string
  retrying: boolean
  retry: () => void
}) {
  const links = useLinks()
  const { data: viewer } = useQuery(viewerQueryOptions())
  const viewerId = viewer?.status === 'signed-in' ? viewer.userId : undefined
  const viewerEntry = useViewerEntry(data?.latestContest?.id, viewerId)
  // Guests get Join in the hero; a signed-in learner who hasn't joined gets it on the board.
  const { registration } = useParticipation(phase === 'open' || phase === 'upcoming' ? data?.latestContest?.id : undefined)
  if (phase === 'loading') {
    return <>
      <SectionHeading level={2} id="home-board-title" className={ledgerHeader} title="Current contest" />
      <Loading className={message} label="Loading the current contest" />
    </>
  }

  const unavailable = <>
    <SectionHeading level={2} id="home-board-title" className={ledgerHeader} title={data?.latestContest?.title ?? 'Current contest'} description="Standings temporarily unavailable" />
    <div className={message} role="status">
      <h3>The board could not be loaded</h3>
      <p>Try again to see the current standings. You can still log your activity in the meantime.</p>
      <Button variant="outline" loading={retrying} onClick={retry}>Retry standings</Button>
    </div>
  </>
  if (phase === 'unavailable' || !data) return unavailable

  const contest = data.latestContest
  if (!contest) {
    return <>
      <SectionHeading level={2} id="home-board-title" className={ledgerHeader} title="Official contests" />
      <EmptyState
        className={message}
        variant="plain"
        headingLevel={3}
        title="No official round right now"
        description="Keep logging at your own pace. The next round will appear here once it’s scheduled."
        actions={<AppLink href={links.contests} className="text-link hover:text-link-hover">Browse contests</AppLink>}
      />
    </>
  }

  const length = dayNumber(contest.contest_start, contest.contest_end)
  const join = registration?.kind === 'not-registered' ? (
    <div className="flex flex-wrap items-center justify-between gap-x-5 gap-y-3 border-b border-rule px-ledger-inset py-4">
      <p className="m-0 text-metadata leading-160 text-muted">Registration closes {formatDate(contest.registration_end)}.</p>
      <AppLink href={links.joinContest(contest.id)} className={buttonClassName()}>Join contest</AppLink>
    </div>
  ) : null
  const header = (description: string, action: ReactNode) => <>
    <SectionHeading level={2} id="home-board-title" className={ledgerHeader} title={contest.title} description={description} actions={action} />
    {join}
  </>

  if (phase === 'upcoming') {
    const activities = contest.allowed_activities.map(({ name }, index) => index ? name.toLowerCase() : name).join(' & ')
    return <>
      {header(formatPeriod(contest), <AppLink href={links.contest(contest.id)} className="text-link hover:text-link-hover">Contest details</AppLink>)}
      <StatGrid className="px-ledger-inset py-4.5" stats={[{ label: 'Length', value: `${length} days` }, ...(activities ? [{ label: 'Activities', value: activities }] : [])]} />
      <div className={message}>
        <p className="grid gap-1">
          <span className="text-metadata text-muted">Starts</span>
          <strong className="font-display text-3xl leading-130">{formatDate(contest.contest_start)}</strong>
        </p>
      </div>
    </>
  }

  const { summary, standings } = data
  if (!summary || !standings) return unavailable
  const ended = phase === 'ended'
  const you = viewerEntry && !standings.entries.some(row => row.user_id === viewerEntry.user_id) ? viewerEntry : null
  const rows = you ? [...standings.entries, you] : standings.entries
  return <>
    {header(
      `${formatPeriod(contest)} · ${ended ? 'Final results' : `Day ${dayNumber(contest.contest_start, today)} of ${length}`}`,
      <AppLink href={links.contest(contest.id)} className="text-link hover:text-link-hover">{ended ? 'See final board' : 'Full leaderboard'}</AppLink>,
    )}
    <StatGrid className="px-ledger-inset py-4.5" stats={[
      { label: 'Participants', value: numberFormat.format(summary.participant_count) },
      { label: 'Languages', value: numberFormat.format(summary.language_count) },
      { label: 'Points logged', value: scoreFormat.format(summary.total_score) },
    ]} />
    {standings.entries.length ? (
      <Table
        caption={`${ended ? 'Final' : 'Current'} top ${standings.entries.length} in ${contest.title}`}
        captionVisibility="screen-reader"
        minWidth="0"
        linkedRows
        columns={[
          { id: 'rank', header: 'Rank', width: '4.5rem', headerClassName: 'ps-ledger-inset', cellClassName: 'ps-ledger-inset text-muted', cell: row => <Rank rank={row.rank} /> },
          {
            id: 'person',
            header: 'Participant',
            rowHeader: true,
            cell: row => (
              <Person you={row.user_id === viewerId}>
                <AppLink className={`home-person-link ${personLink}`} href={links.profile(row.user_id)}>{row.user_display_name}</AppLink>
              </Person>
            ),
          },
          { id: 'score', header: 'Score', align: 'end', headerClassName: 'pe-ledger-inset', cellClassName: 'pe-ledger-inset', cell: row => scoreFormat.format(row.score) },
        ]}
        rows={rows}
        getRowKey={row => row.user_id}
        tableClassName={cardTable}
        getRowProps={row => (row.user_id === viewerId ? { 'aria-current': 'true', className: currentRow } : {})}
      />
    ) : (
      <EmptyState
        className={message}
        variant="plain"
        headingLevel={3}
        title="No scores yet"
        description="Standings appear here once participants log reading or listening in this round."
      />
    )}
  </>
}
