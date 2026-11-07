import { useStore } from '@tanstack/react-form'
import { Link, useLocation } from 'react-router-dom'
import { Button, buttonClassName, EmptyState, PageHeader, SectionHeading, StatusTag, Surface, Tabbar, typography, usePaperForm } from 'paper-ui'
import { contestStatus, type SampleContest } from '../data'
import { contestEnd, contestStart, formatDateRange, formatDateTime } from '../dates'
import { usePlayground, useScenario } from '../state'
import { ContestFacts, ContestPrimaryAction, ContestStatus } from './contestParts'
import { contestPath, registrationIsOpen, useContestDate } from './contestRound'

const scopeLabels = { official: 'Official', community: 'Community', participating: 'Participating', managed: 'Managed by me' }
type ContestCollection = keyof typeof scopeLabels

function ContestFeature({ contest, asOf }: { contest: SampleContest; asOf: string }) {
  const { viewer, joinedContests } = usePlayground()
  return <Surface as="article" className="grid grid-cols-main-aside p-0 max-tablet:grid-cols-single">
    <div className="p-7 max-md:p-5 [&_p]:mb-3 [&_p]:mt-0 [&_p]:max-w-measure-feature [&>:last-child]:mt-6"><SectionHeading level={2} className="[&_h2_a]:text-inherit [&_h2_a]:no-underline [&_h2_a:hover]:underline" title={<Link to={contestPath(contest, asOf)}>{contest.title}</Link>} status={<ContestStatus contest={contest} asOf={asOf} />} /><p>{contest.description}</p><p className="text-muted">{contest.languages.join(', ')} · {contest.activities.join(' and ')}</p>
      {viewer !== 'guest' && joinedContests.includes(contest.id) ? <p><strong>You’re registered.</strong></p> : null}
      <div className="flex flex-wrap gap-3"><ContestPrimaryAction contest={contest} asOf={asOf} prominent /><Link className={buttonClassName({ variant: 'outline' })} to={contestPath(contest, asOf, '/leaderboard')}>View standings</Link></div>
    </div>
    <div className="border-s border-rule p-7 max-tablet:border-s-0 max-tablet:border-t max-md:p-5 max-tablet:[&_.contest-facts]:grid-cols-thirds max-md:[&_.contest-facts]:grid-cols-pair max-md:[&_.contest-facts]:gap-4 max-md:[&_.contest-facts>:last-child]:col-span-full"><ContestFacts contest={contest} asOf={asOf} /><Link className="text-link" to="/guide">Read the contest guide</Link></div>
  </Surface>
}
function ContestRow({ contest, asOf, managed = false }: { contest: SampleContest; asOf: string; managed?: boolean }) {
  const { viewer, userId, joinedContests } = usePlayground()
  const date = new Date(contestStart(contest.start))
  const canManage = ['organizer', 'admin'].includes(viewer) && contest.ownerId === userId
  return <article className="contest-row grid grid-cols-dated-row gap-x-6 border-t border-rule py-6 max-md:grid-cols-dated-row-narrow max-md:gap-x-4 max-md:py-5">
    <div className="pt-0.5 text-center [&_span]:block [&_span]:text-xs [&_span]:leading-140 [&_span]:text-muted [&_strong]:block [&_strong]:font-display [&_strong]:text-figure [&_strong]:leading-150" aria-hidden="true"><span>{new Intl.DateTimeFormat(undefined, { month: 'short', year: 'numeric' }).format(date)}</span><strong>{new Intl.DateTimeFormat(undefined, { day: '2-digit' }).format(date)}</strong></div>
    <div className="min-w-0 [&_h3]:m-0 [&_h3_a]:text-inherit [&_h3_a]:no-underline [&_h3_a:hover]:underline [&_p]:mb-0 [&_p]:mt-1.5 [&_p]:text-sm [&_p]:leading-160"><div className="flex flex-wrap items-center gap-2"><h3 className={typography({ role: 'component' })}><Link className="text-link" to={contestPath(contest, asOf)}>{contest.title}</Link></h3>{contest.unlisted ? <StatusTag>Unlisted</StatusTag> : null}</div><p>{formatDateRange(contestStart(contest.start), contestEnd(contest.end))}</p><p className="text-muted">{contest.languages.join(', ')} · {contest.activities.join(' and ')}</p></div>
    <div className="flex flex-col items-end gap-3 max-md:col-start-2 max-md:mt-3 max-md:flex-row max-md:flex-wrap max-md:items-center"><ContestStatus contest={contest} asOf={asOf} />{managed ? <Link className={buttonClassName({ variant: 'outline' })} to={contestPath(contest, asOf, canManage ? '/edit' : '')}>{canManage ? 'Manage contest' : 'View contest'}</Link> : <ContestPrimaryAction contest={contest} asOf={asOf} />}</div>
    <details className="col-start-2 -col-end-1 mt-2.5 min-w-0 text-sm max-md:col-end-auto [&_summary]:w-fit [&_summary]:cursor-pointer [&_summary]:text-link [&>div]:max-w-measure-prose [&>div]:pt-1 [&_p]:my-3"><summary>Contest details</summary><div><p>{contest.description}</p><p><strong>{registrationIsOpen(contest, asOf) ? `Registration until ${formatDateTime(contestEnd(contest.registrationDeadline))}.` : 'Registration closed.'}</strong>{viewer !== 'guest' && joinedContests.includes(contest.id) ? ' You are registered.' : ''}</p>{contest.unlisted ? <p>Anyone with the link can view this contest. It does not appear in public discovery.</p> : null}<div className="flex flex-wrap gap-x-6 gap-y-2"><Link className="text-link" to={contestPath(contest, asOf, '/leaderboard')}>{contestStatus(contest, asOf) === 'ended' ? 'View final standings' : 'View standings'}</Link><Link className="text-link" to="/guide/scoring">How contest scoring works</Link>{canManage && !managed ? <Link className="text-link" to={contestPath(contest, asOf, '/edit')}>Manage contest</Link> : null}</div></div></details>
  </article>
}

export function ContestsPage() {
  const location = useLocation()
  const { contests, viewer, userId, joinedContests } = usePlayground()
  const [scenario, setScenario] = useScenario('contests')
  const asOf = useContestDate()
  const collection = location.pathname.split('/')[2]
  const scope: ContestCollection = collection === 'mine' ? 'managed' : collection === 'community' || collection === 'participating' || collection === 'managed' ? collection : 'official'
  const methods = usePaperForm({ defaultValues: { query: '', period: 'all' } })
  const { query = '', period = 'all' } = useStore(methods.store, state => state.values)
  const signedIn = viewer !== 'guest'
  const blocked = !signedIn && (scope === 'participating' || scope === 'managed')
  const managedContests = signedIn && scenario !== 'empty' ? contests.filter(contest => contest.ownerId === userId || contest.moderatorIds?.includes(userId)) : []
  const collections: ContestCollection[] = ['official', 'community', ...(signedIn ? ['participating' as const] : []), ...(managedContests.length ? ['managed' as const] : [])]
  const isError = scenario === 'unavailable'
  const canCreate = ['organizer', 'admin'].includes(viewer)
  const collectionRows = scope === 'managed' ? managedContests : contests.filter(contest => scope === 'participating' ? signedIn && joinedContests.includes(contest.id) : !contest.unlisted && contest.scope === scope)
  const rows = collectionRows.filter(contest => (period === 'all' || contestStatus(contest, asOf) === period) && `${contest.title} ${contest.languages.join(' ')}`.toLocaleLowerCase().includes(query.trim().toLocaleLowerCase()))
    .sort((left, right) => left.start.localeCompare(right.start) || left.title.localeCompare(right.title))
  const feature = scope === 'official' && !query.trim() && period === 'all' ? rows.find(contest => contestStatus(contest, asOf) === 'live') ?? rows.find(contest => contestStatus(contest, asOf) === 'upcoming') : undefined
  const descriptions = { official: 'Official rounds bring the Tadoku community together throughout the year.', community: 'Public challenges organized by Tadoku members. Check the languages and activities before joining.', participating: 'Contests you’ve joined, including past rounds.', managed: 'Contests you organize or moderate, including unlisted contests.' }
  const emptyCollection = collectionRows.length === 0 && !query && period === 'all' && (scope === 'participating' || scope === 'managed')
  const clearFilters = () => { methods.reset(); document.getElementsByName('query')[0]?.focus() }

  return <>
    <PageHeader title="Contests" lead={descriptions[scope]} actions={canCreate ? <Link className={buttonClassName()} to={`/contests/new?asOf=${asOf}`}>Create contest</Link> : null} />
    <div className="max-xs:[&_ul]:grid max-xs:[&_ul]:grid-cols-halves max-xs:[&_a]:px-2"><Tabbar label="Contest collections" links={collections.map(id => ({ id, label: scopeLabels[id], href: `/contests/${id}${location.search}`, current: scope === id, onSelect: () => methods.reset() }))} renderLink={({ href, ...props }) => <Link to={href} {...props} />} /></div>
    {blocked ? <EmptyState headingLevel={2} title={scope === 'participating' ? 'Sign in to see the contests you’ve joined.' : 'Sign in to see the contests you manage.'} description="You can still browse official rounds and public community challenges." actions={<Link className={buttonClassName()} to={`/sign-in?next=${encodeURIComponent(`/contests/${scope}`)}`}>Sign in</Link>} /> : isError ? <EmptyState role="alert" headingLevel={2} title="Contests could not be loaded." description="Try loading this collection again." actions={<Button onClick={() => setScenario('member')}>Try again</Button>} /> : <>
      <div className="contest-toolbar flex items-end justify-between gap-5 py-6 max-md:py-5"><methods.AppForm><form className="flex flex-1 flex-wrap items-end gap-4 [&>.paper-field]:grow-0 [&>.paper-field]:basis-filter [&>.paper-field:first-child]:max-w-filter-search-max [&>.paper-field:first-child]:grow [&>.paper-field:first-child]:basis-filter-search max-md:gap-3 max-md:[&>.paper-field]:grow max-md:[&>.paper-field]:basis-filter-narrow max-md:[&>.paper-field:first-child]:max-w-none max-md:[&>.paper-field:first-child]:basis-full max-md:[&>.paper-button]:self-end" onSubmit={event => event.preventDefault()}><methods.AppField name="query">{field => <field.Input label="Find a contest" type="search" placeholder="Search title or language" />}</methods.AppField><methods.AppField name="period">{field => <field.Select label="Dates" options={[{ value: 'all', label: 'All dates' }, { value: 'live', label: 'Live now' }, { value: 'upcoming', label: 'Upcoming' }, { value: 'ended', label: 'Past contests' }]} />}</methods.AppField>{query || period !== 'all' ? <Button variant="ghost" onClick={clearFilters}>Clear filters</Button> : null}</form></methods.AppForm>
      <p className="mb-2.5 mt-0 shrink-0 text-fine max-tablet:sr-only text-muted" role="status">{rows.length} {rows.length === 1 ? 'contest' : 'contests'}{query.trim() || period !== 'all' ? ' match your filters' : ''}</p></div>
      {rows.length === 0 ? <EmptyState headingLevel={2} title={emptyCollection ? scope === 'participating' ? 'You haven’t joined a contest yet.' : 'You don’t manage any contests yet.' : 'No contests match these filters.'} description={emptyCollection ? scope === 'participating' ? 'Join an official round or community challenge to see it here.' : 'Contests appear here when you organize them or are assigned as a moderator.' : 'Try another title or language, or include all dates.'} actions={emptyCollection ? <Link className={buttonClassName({ variant: 'outline' })} to="/contests/official">Explore official contests</Link> : <Button variant="outline" onClick={clearFilters}>Clear filters</Button>} /> : <>
        {feature ? <ContestFeature contest={feature} asOf={asOf} /> : null}
        {(['live', 'upcoming', 'ended'] as const).map(status => {
          const group = rows.filter(contest => contest.id !== feature?.id && contestStatus(contest, asOf) === status)
          if (status === 'ended') group.reverse()
          return group.length > 0 ? <section className="mt-9 [.contest-toolbar+&]:mt-2" key={status} aria-label={{ live: 'Happening now', upcoming: 'Coming up', ended: 'Past contests' }[status]}><SectionHeading level={2} title={{ live: 'Happening now', upcoming: 'Coming up', ended: 'Past contests' }[status]} />{group.map(contest => <ContestRow key={contest.id} contest={contest} asOf={asOf} managed={scope === 'managed'} />)}</section> : null
        })}
      </>}
    </>}
    <p className="mb-0 mt-3 border-t border-rule pt-5 text-fine text-muted">Contest times are shown in your local time zone.</p>
  </>
}
