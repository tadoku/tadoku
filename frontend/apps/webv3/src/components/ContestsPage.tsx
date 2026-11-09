import { useQuery } from '@tanstack/react-query'
import { PlusIcon, iconClassName } from 'paper-ui/icons'
import { Link } from '@tanstack/react-router'
import { Button, buttonClassName, EmptyState, listFooter, PageHeader, Pagination, StatusTag, Tabbar, typography } from 'paper-ui'
import {
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
import { standingsHeader } from './standings-styles'

const tabs: { id: ContestCollection; label: string }[] = [
  { id: 'official', label: 'Official' },
  { id: 'community', label: 'Community' },
  { id: 'joined', label: 'Joined' },
  { id: 'organizing', label: 'Organizing' },
]

/**
 * One page of the official or community contests, the contests the viewer joined, or the ones they organize, with the
 * collection tabs and create action.
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
    <div className="contests-page">
      <PageHeader
        className={standingsHeader}
        title="Contests"
        actions={(
          <>
            {canCreate ? <Link className={buttonClassName({ variant: 'outline' })} to={links.newContest}><PlusIcon className={iconClassName('compact')} aria-hidden="true" />Create contest</Link> : null}
          </>
        )}
      />
      <Tabbar
        className="mt-2"
        label="Contest collections"
        renderLink={renderLink}
        links={tabs.filter(tab => viewer?.status !== 'anonymous' || tab.id === 'official' || tab.id === 'community').map(tab => ({
          ...tab,
          href: contestsHref[tab.id](),
          current: tab.id === collection,
        }))}
      />
      {scope ? (
        <Contests scope={scope} view={view} page={page} canCreate={canCreate} />
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
    </div>
  )
}

function SignInPrompt({ joined }: { joined: boolean }) {
  return (
    <EmptyState
      headingLevel={2}
      title={joined ? 'Sign in to see the contests you joined' : 'Sign in to see the contests you organize'}
      description="You can still browse official rounds and community contests."
      actions={<a className={buttonClassName()} href={useLoginUrl()}>Log in</a>}
    />
  )
}

function Contests({ scope, view, page, canCreate }: { scope: ContestListScope; view: ContestView; page: number; canCreate: boolean }) {
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
  if (!result) return <p className="text-metadata text-muted" role="status">Loading contests…</p>

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
        actions={<Link className="text-link hover:text-link-hover" to={pageHref(1)}>Go to the first page</Link>}
      />
    ) : view.collection === 'organizing' ? (
      <EmptyState
        headingLevel={2}
        title="You haven’t organized a contest yet"
        description={canCreate ? undefined : 'Accounts can create contests from one month after joining. Until then, join one that’s running.'}
        actions={canCreate
          ? <Link className={buttonClassName()} to="/contests/new">Create contest</Link>
          : <Link className={buttonClassName({ variant: 'outline' })} to={contestsHref.community()}>Browse community contests</Link>}
      />
    ) : view.collection === 'joined' ? (
      <EmptyState
        headingLevel={2}
        title="You haven’t joined a contest yet"
        actions={<Link className={buttonClassName({ variant: 'outline' })} to={contestsHref.official()}>Browse official contests</Link>}
      />
    ) : <EmptyState headingLevel={2} title="No contests found" description="Contests appear here once they’re scheduled." />
  }

  const first = (page - 1) * contestsPageSize + 1
  const totalPages = Math.max(page, Math.ceil(total / contestsPageSize))
  return (
    <>
      <ul className="mb-0 mt-2 list-none p-0" aria-label={view.collection === 'joined' ? 'Contests you joined' : `${tabs.find(tab => tab.id === view.collection)?.label} contests`}>
        {contests.map(contest => <ContestRow key={contest.id} contest={contest} names={names} today={today} />)}
      </ul>
      {totalPages > 1 ? (
        <div className={listFooter().root({ className: 'contests-pagination border-t border-rule pt-3' })}>
          <p className={listFooter().summary()}>{numberFormat.format(first)}–{numberFormat.format(first + contests.length - 1)} of {numberFormat.format(total)} contests</p>
          <Pagination totalPages={totalPages} currentPage={page} getHref={pageHref} renderLink={renderLink} label="Contest pages" classNames={{ list: 'max-md:flex-wrap' }} />
        </div>
      ) : null}
    </>
  )
}

const rowText = 'mb-0 mt-1.5 text-sm leading-160 text-muted'

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
  const meta = [registration, restrictions ? languages : null, restrictions ? activities : null].filter(Boolean)

  return (
    <li className="contest-row border-t border-rule py-4 first:border-t-0">
      <div className="contest-row__body min-w-0 wrap-anywhere lg:grid lg:grid-cols-content-end lg:items-baseline lg:gap-x-8">
        <div
          className="contest-row__title grid grid-cols-title-tags items-start justify-start gap-x-2 gap-y-1 max-sm:grid-cols-single"
        >
          <h2 className={typography({ role: 'component', className: 'm-0 font-body text-base leading-150 font-semibold' })}>
            <a className="text-inherit no-underline hover:text-link hover:underline max-lg:underline max-lg:decoration-rule max-lg:underline-offset-link" href={links.contest(contest.id)}>
              {contest.title}
            </a>
          </h2>
          <span className="contest-row__tags mt-0.5 flex flex-wrap gap-x-1.5 gap-y-1 max-sm:mt-0">
            {round}{contest.private ? <StatusTag>Unlisted</StatusTag> : null}
          </span>
        </div>
        <div className="lg:text-end">
          <p className={`${rowText} lg:mt-0`}>{formatPeriod(contest)}</p>
          {meta.length ? (
            <p className={`${rowText} mt-0 flex flex-wrap gap-x-3 lg:justify-end`}>
              {meta.map((part, index) => <span key={index} className="contest-row__allows whitespace-nowrap">{part}</span>)}
            </p>
          ) : null}
        </div>
      </div>
    </li>
  )
}
