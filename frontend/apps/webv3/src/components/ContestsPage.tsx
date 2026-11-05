import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { Button, EmptyState, PageHeader, Pagination, StatusTag, Tabbar, buttonClassName } from 'paper-ui'
import type { ReactNode } from 'react'
import {
  canCreateContestQuery,
  contestScope,
  contestsHref,
  contestsPageSize,
  contestsQuery,
  type ContestCollection,
  type ContestList,
  type ContestListScope,
} from '../lib/contests'
import { formatPeriod, months, numberFormat } from '../lib/format'
import { useLinks } from '../lib/links'
import { useLoginUrl, viewerKey, viewerQueryOptions } from '../lib/session'
import { renderLink } from './router-link'

const tabs: { id: ContestCollection; label: string }[] = [
  { id: 'official', label: 'Official' },
  { id: 'community', label: 'Community' },
  { id: 'mine', label: 'My contests' },
]

const leads: Record<ContestCollection, string> = {
  official: 'Official rounds bring the Tadoku community together throughout the year.',
  community: 'Challenges organized by Tadoku members. Check the languages and activities before joining.',
  mine: 'Community contests you organize, including unlisted ones.',
}

/** One page of the official, community or viewer's own contests, with the collection tabs and create action. */
export function ContestsPage({ collection, page }: { collection: ContestCollection; page: number }) {
  const links = useLinks()
  const viewerQuery = useQuery(viewerQueryOptions())
  const viewer = viewerQuery.data
  const signedIn = viewer?.status === 'signed-in'
  const canCreate = useQuery({ ...canCreateContestQuery(viewerKey(viewer)), enabled: signedIn }).data === true
  const scope = contestScope(collection, viewer)

  return (
    <div className="contests-page">
      <PageHeader
        title="Contests"
        lead={leads[collection]}
        actions={canCreate ? <Link className={buttonClassName()} to={links.newContest}>Create contest</Link> : null}
      />
      <Tabbar
        label="Contest collections"
        renderLink={renderLink}
        links={tabs.map(tab => ({ ...tab, href: contestsHref[tab.id](), current: tab.id === collection }))}
      />
      {scope ? (
        <Contests scope={scope} collection={collection} page={page} />
      ) : viewer?.status === 'anonymous' ? (
        <SignInPrompt />
      ) : (
        <EmptyState
          headingLevel={2}
          title="Account service unavailable"
          description="We could not check your session. Try again."
          actions={<Button variant="outline" loading={viewerQuery.isFetching} onClick={() => void viewerQuery.refetch()}>Try again</Button>}
        />
      )}
      <p className="contests-note">Contest dates are calendar days in UTC.</p>
    </div>
  )
}

function SignInPrompt() {
  return (
    <EmptyState
      headingLevel={2}
      title="Sign in to see the contests you organize"
      description="You can still browse official rounds and community contests."
      actions={<a className={buttonClassName()} href={useLoginUrl()}>Log in</a>}
    />
  )
}

function Contests({ scope, collection, page }: { scope: ContestListScope; collection: ContestCollection; page: number }) {
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
  if (!result) return <p className="contests-status" role="status">Loading contests…</p>

  const { contests, total } = result.data
  // A session cookie can belong to a guest, whom the configuration endpoint also answers; Webv2 names restrictions
  // for signed-in viewers only.
  const names = viewer?.status === 'signed-in' ? result.data.names : null
  const pageHref = contestsHref[collection]
  if (contests.length === 0) {
    return total > 0 ? (
      <EmptyState
        headingLevel={2}
        title="This page is past the end of the list"
        description={`There are ${numberFormat.format(total)} contests.`}
        actions={<Link className="text-link" to={pageHref(1)}>Go to the first page</Link>}
      />
    ) : collection === 'mine' ? (
      <EmptyState
        headingLevel={2}
        title="You don’t organize any contests yet"
        description="Contests you create appear here."
        actions={<Link className={buttonClassName({ variant: 'outline' })} to={contestsHref.community()}>Browse community contests</Link>}
      />
    ) : <EmptyState headingLevel={2} title="No contests found" description="Contests appear here once they’re scheduled." />
  }

  const first = (page - 1) * contestsPageSize + 1
  const totalPages = Math.max(page, Math.ceil(total / contestsPageSize))
  return (
    <>
      <ul className="contest-list" aria-label={`${tabs.find(tab => tab.id === collection)?.label} contests`}>
        {contests.map(contest => <ContestRow key={contest.id} contest={contest} names={names} today={today} />)}
      </ul>
      <div className="contests-pagination">
        <p>{numberFormat.format(first)}–{numberFormat.format(first + contests.length - 1)} of {numberFormat.format(total)} contests</p>
        {totalPages > 1
          ? <Pagination totalPages={totalPages} currentPage={page} getHref={pageHref} renderLink={renderLink} label="Contest pages" />
          : null}
      </div>
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
  const [year, month, day] = contest.contest_start.split('-')
  const started = today >= contest.contest_start
  const ended = today > contest.contest_end
  const round = ended ? <StatusTag>Ended</StatusTag> : started ? <StatusTag tone="success">Live now</StatusTag> : <StatusTag tone="information">Upcoming</StatusTag>
  let registration: ReactNode = null
  if (!ended) {
    registration = today <= contest.registration_end
      ? <StatusTag tone="information">Registration open</StatusTag>
      : <StatusTag>Registration closed</StatusTag>
  }
  const languages = contest.language_code_allow_list?.length
    ? contest.language_code_allow_list.map(code => names?.languages[code] ?? code).join(', ')
    : 'All languages'
  const activities = contest.activity_type_id_allow_list.length
    ? contest.activity_type_id_allow_list.map(id => names?.activities[id] ?? String(id)).join(', ')
    : 'All activities'

  return (
    <li className="contest-row">
      <div className="contest-date" aria-hidden="true">
        <span>{months[Number(month) - 1]} {year}</span>
        <strong>{Number(day)}</strong>
      </div>
      <div className="contest-row__body">
        <div className="contest-row__title">
          <h2 className="paper-type-component"><a href={links.contest(contest.id)}>{contest.title}</a></h2>
          {contest.private ? <StatusTag>Unlisted</StatusTag> : null}
        </div>
        <p>{formatPeriod(contest)}</p>
        {names ? <p className="contest-row__allows">{languages} · {activities}</p> : null}
      </div>
      <div className="contest-row__status">{round}{registration}</div>
    </li>
  )
}
