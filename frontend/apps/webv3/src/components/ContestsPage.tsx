import * as stylex from '@stylexjs/stylex'
import { useQuery } from '@tanstack/react-query'
import { PlusIcon } from 'paper-ui/icons'
import { Link } from '@tanstack/react-router'
import { Button, buttonStyles, EmptyState, iconSizes, linkStyles, PageHeader, Pagination, StatusTag, Tabbar, typography } from 'paper-ui'
import { border, breakpoints, color, font, fontSize, fontWeight, lineHeight, space, type } from 'paper-ui/tokens.stylex'
import {
  accountSinceQuery,
  canCreateContestQuery,
  contestScope,
  contestsHref,
  contestsPageSize,
  contestsQuery,
  viewHref,
  type ContestCollection,
  type ContestList,
  type ContestListScope,
  type ContestView,
} from '../lib/contests'
import { formatDate, formatPeriod, numberFormat } from '../lib/format'
import { useLinks } from '../lib/links'
import { useLoginUrl, viewerKey, viewerQueryOptions } from '../lib/session'
import { renderLink } from './router-link'
import { YearStepper } from './YearStepper'
import { standingsLayout } from './page-styles'

const styles = stylex.create({
  tabs: { marginBlockStart: space[8] },
  muted: { color: color.textMuted, fontSize: type.metadataSize },
  note: {
    marginInline: 0,
    borderBlockStartColor: color.ruleDefault,
    borderBlockStartStyle: 'solid',
    borderBlockStartWidth: border.static,
    marginBlockEnd: 0,
    marginBlockStart: space[24],
    paddingBlockStart: space[20],
  },
  /** The Joined year sits under the tabs, its chevron aligned with the list's left edge. */
  year: { marginBlockEnd: 0, marginBlockStart: space[16], marginInlineEnd: 0, marginInlineStart: `calc(-1 * ${space[12]})` },
  paginationTop: { display: 'flex', justifyContent: 'flex-end', marginBlockEnd: space[4], marginBlockStart: space[8] },
  wrap: { flexWrap: 'wrap' },
  phoneWrap: { flexWrap: { default: null, [breakpoints.sm]: 'wrap' } },
  list: { padding: 0, listStyle: 'none', marginInline: 0, marginBlockEnd: 0, marginBlockStart: space[8], },
  row: {
    paddingBlock: space[16],
    paddingInline: 0,
    borderBlockStartColor: color.ruleDefault,
    borderBlockStartStyle: 'solid',
    borderBlockStartWidth: { default: border.static, ':first-child': 0 },
  },
  body: { minInlineSize: 0, overflowWrap: 'anywhere' },
  /** The tag stays on the title's first line; a long title wraps beside it rather than around it. */
  title: {
    alignItems: 'start',
    columnGap: space[8],
    display: 'grid',
    gridTemplateColumns: 'minmax(0, max-content) auto',
    justifyContent: 'start',
    rowGap: space[4],
  },
  /** On a phone two tags would squeeze the title, so it takes the row and the tags sit beneath it. */
  titleWithTags: { gridTemplateColumns: { default: 'minmax(0, max-content) auto', [breakpoints.phone]: 'minmax(0, 1fr)' } },
  heading: { margin: 0, fontFamily: font.body, fontSize: fontSize[16], fontWeight: fontWeight.semibold, lineHeight: lineHeight.normal, },
  titleLink: { color: { default: 'inherit', ':focus-visible': color.textLink, ':hover': color.textLink } },
  tags: { columnGap: space[6], display: 'flex', flexWrap: 'wrap', marginBlockStart: space[2], rowGap: space[4] },
  tagsBelow: { marginBlockStart: { default: space[2], [breakpoints.phone]: 0 } },
  detail: { marginInline: 0, color: color.textMuted, fontSize: fontSize[14], lineHeight: lineHeight.relaxed, marginBlockEnd: 0, marginBlockStart: space[6], },
  pagination: {
    alignItems: 'center',
    borderBlockStartColor: color.ruleDefault,
    borderBlockStartStyle: 'solid',
    borderBlockStartWidth: border.static,
    columnGap: space[16],
    display: 'flex',
    flexWrap: 'wrap',
    justifyContent: 'space-between',
    paddingBlockStart: space[12],
    rowGap: space[8],
  },
  paginationSummary: { margin: 0, color: color.textMuted, fontSize: type.metadataSize, },
})

const tabs: { id: ContestCollection; label: string }[] = [
  { id: 'official', label: 'Official' },
  { id: 'community', label: 'Community' },
  { id: 'joined', label: 'Joined' },
  { id: 'organizing', label: 'Organizing' },
]

const thisYear = () => new Date().getUTCFullYear()

/**
 * One page of the official or community contests, the contests the viewer joined in a year, or the ones they
 * organize, with the collection tabs and create action.
 */
export function ContestsPage({ view, page }: { view: ContestView; page: number }) {
  const { collection } = view
  const links = useLinks()
  const viewerQuery = useQuery(viewerQueryOptions())
  const viewer = viewerQuery.data
  const signedIn = viewer?.status === 'signed-in'
  const canCreate = useQuery({ ...canCreateContestQuery(viewerKey(viewer)), enabled: signedIn }).data === true
  const scope = contestScope(view, viewer)

  return (
    <div>
      <PageHeader
        xstyle={standingsLayout.header}
        title="Contests"
        actions={canCreate ? <Link {...stylex.props(...buttonStyles({ variant: 'outline' }))} to={links.newContest}><PlusIcon {...stylex.props(iconSizes.compact)} aria-hidden="true" />Create contest</Link> : null}
      />
      <Tabbar
        label="Contest collections"
        renderLink={renderLink}
        xstyle={styles.tabs}
        links={tabs.map(tab => ({
          ...tab,
          href: tab.id === 'joined' ? contestsHref.joined(thisYear()) : contestsHref[tab.id](),
          current: tab.id === collection,
        }))}
      />
      {view.collection === 'joined' && scope?.kind === 'participant' ? <JoinedYear userId={scope.userId} year={view.year} /> : null}
      {scope ? (
        <Contests scope={scope} view={view} page={page} />
      ) : viewer?.status === 'anonymous' ? (
        <SignInPrompt joined={collection === 'joined'} />
      ) : (
        <EmptyState
          headingLevel={2}
          title="Account service unavailable"
          description="We could not check your session. Try again."
          actions={<Button variant="outline" loading={viewerQuery.isFetching} onClick={() => void viewerQuery.refetch()}>Try again</Button>}
        />
      )}
      <p {...stylex.props(styles.muted, styles.note)}>Contest dates are calendar days in UTC.</p>
    </div>
  )
}

/** The Joined list's year, stepping between the year the account joined Tadoku and this year. */
function JoinedYear({ userId, year }: { userId: string; year: number }) {
  const since = useQuery(accountSinceQuery(userId)).data
  const current = thisYear()
  return (
    <div {...stylex.props(styles.year)}>
      <YearStepper year={year} first={Math.min(since ?? current, current)} current={current} href={contestsHref.joined} label="Year" />
    </div>
  )
}

function SignInPrompt({ joined }: { joined: boolean }) {
  return (
    <EmptyState
      headingLevel={2}
      title={joined ? 'Sign in to see the contests you joined' : 'Sign in to see the contests you organize'}
      description="You can still browse official rounds and community contests."
      actions={<a {...stylex.props(...buttonStyles())} href={useLoginUrl()}>Log in</a>}
    />
  )
}

function Contests({ scope, view, page }: { scope: ContestListScope; view: ContestView; page: number }) {
  const { data: viewer } = useQuery(viewerQueryOptions())
  const list = useQuery(contestsQuery(scope, page, viewerKey(viewer)))
  // The route loader read the list; its read time decides each contest's phase on the server and after hydration.
  const today = new Date(list.dataUpdatedAt).toISOString().slice(0, 10)
  const result = list.data

  if (result && !result.ok) {
    return (
      <EmptyState
        headingLevel={2}
        title="Contests could not be loaded"
        description="Try loading this list again in a moment."
        actions={<Button variant="outline" loading={list.isFetching} onClick={() => void list.refetch()}>Try again</Button>}
      />
    )
  }
  if (!result) return <p {...stylex.props(styles.muted)} role="status">Loading contests…</p>

  const { contests, total } = result.data
  // A session cookie can belong to a guest, whom the configuration endpoint also answers; Webv2 names restrictions
  // for signed-in viewers only.
  const names = viewer?.status === 'signed-in' ? result.data.names : null
  const pageHref = (next: number) => viewHref(view, next)
  if (contests.length === 0) {
    return total > 0 ? (
      <EmptyState
        headingLevel={2}
        title="This page is past the end of the list"
        description={`There are ${numberFormat.format(total)} contests.`}
        actions={<Link {...stylex.props(linkStyles.base)} to={pageHref(1)}>Go to the first page</Link>}
      />
    ) : view.collection === 'organizing' ? (
      <EmptyState
        headingLevel={2}
        title="You don’t organize any contests yet"
        actions={<Link {...stylex.props(...buttonStyles({ variant: 'outline' }))} to={contestsHref.community()}>Browse community contests</Link>}
      />
    ) : view.collection === 'joined' ? (
      <EmptyState
        headingLevel={2}
        title={`You didn’t join any contests in ${view.year}`}
        actions={<Link {...stylex.props(...buttonStyles({ variant: 'outline' }))} to={contestsHref.official()}>Browse official contests</Link>}
      />
    ) : <EmptyState headingLevel={2} title="No contests found" description="Contests appear here once they’re scheduled." />
  }

  const first = (page - 1) * contestsPageSize + 1
  const totalPages = Math.max(page, Math.ceil(total / contestsPageSize))
  return (
    <>
      {totalPages > 1 ? (
        <div {...stylex.props(styles.paginationTop)}>
          <Pagination totalPages={totalPages} currentPage={page} getHref={pageHref} renderLink={renderLink} label="Contest pages, top" listXstyle={styles.wrap} />
        </div>
      ) : null}
      <ul {...stylex.props(styles.list)} aria-label={view.collection === 'joined' ? `Contests joined in ${view.year}` : `${tabs.find(tab => tab.id === view.collection)?.label} contests`}>
        {contests.map(contest => <ContestRow key={contest.id} contest={contest} names={names} today={today} />)}
      </ul>
      {totalPages > 1 ? (
        <div {...stylex.props(styles.pagination)} data-contests-pagination="">
          <p {...stylex.props(styles.paginationSummary)}>{numberFormat.format(first)}–{numberFormat.format(first + contests.length - 1)} of {numberFormat.format(total)} contests</p>
          <Pagination totalPages={totalPages} currentPage={page} getHref={pageHref} renderLink={renderLink} label="Contest pages" listXstyle={styles.phoneWrap} />
        </div>
      ) : null}
    </>
  )
}

function ContestRow({ contest, names, today }: {
  contest: ContestList['contests'][number]
  names: ContestList['names']
  /** The UTC calendar day the list was read. */
  today: string
}) {
  const links = useLinks()
  const started = today >= contest.contest_start
  const ended = today > contest.contest_end
  const round = ended ? <StatusTag>Ended</StatusTag> : started ? <StatusTag tone="success">Live now</StatusTag> : <StatusTag tone="information">Upcoming</StatusTag>
  const registration = !ended && today <= contest.registration_end ? `Registration closes ${formatDate(contest.registration_end)}` : null
  // Only restrictions earn a mention; an unrestricted language or activity list is the default.
  const languages = contest.language_code_allow_list?.length
    ? contest.language_code_allow_list.map(code => names?.languages[code] ?? code).join(', ')
    : null
  const activities = contest.activity_type_id_allow_list.length
    ? contest.activity_type_id_allow_list.map(id => names?.activities[id] ?? String(id)).join(', ')
    : null
  // Official rounds are open to every language and activity in practice, and the Official list never names them.
  const restrictions = names && !contest.official
  const meta = [registration, restrictions ? languages : null, restrictions ? activities : null].filter(Boolean).join(' · ')

  return (
    <li {...stylex.props(styles.row)} data-contest-row="">
      <div {...stylex.props(styles.body)}>
        <div {...stylex.props(styles.title, contest.private && styles.titleWithTags)}>
          <h2 {...stylex.props(typography.component, styles.heading)}>
            <a href={links.contest(contest.id)} {...stylex.props(linkStyles.base, linkStyles.quiet, styles.titleLink)}>{contest.title}</a>
          </h2>
          <span {...stylex.props(styles.tags, contest.private && styles.tagsBelow)}>{round}{contest.private ? <StatusTag>Unlisted</StatusTag> : null}</span>
        </div>
        <p {...stylex.props(styles.detail)}>{formatPeriod(contest)}</p>
        {meta ? <p {...stylex.props(styles.detail)} data-contest-allows="">{meta}</p> : null}
      </div>
    </li>
  )
}
