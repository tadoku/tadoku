import { useStore } from '@tanstack/react-form'
import { useEffect, useId, useRef, useState } from 'react'
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom'
import { AppliedFilters, Button, buttonClassName, DescriptionList, EmptyState, PageHeader, Pagination, SectionHeading, StatusTag, Tabbar, Table, type TableColumn, typography, usePaperForm } from 'paper-ui'
import { ChevronDownIcon, iconClassName } from 'paper-ui/icons'
import { contestStatus, formatDate, formatNumber, initialLogs, leaderboardPeople, supportedLanguages, users, type SampleLog } from '../data'
import { contestEnd, contestStart, formatDateRange, formatDateTime } from '../dates'
import { usePlayground, useScenario } from '../state'

interface Standing { userId: string; name: string; score: number; rank: number; tied: boolean }
interface BoardFilters { language: string; activity: string; query: string; year: string }
const pageSize = 6

function scoreMatches(log: SampleLog, language: string, activity: string) {
  return (language === 'all' || log.language === language) && (activity === 'all' || log.activity === activity)
}

export function LeaderboardPage() {
  const { contestId, year = '2026' } = useParams()
  const location = useLocation()
  const navigate = useNavigate()
  const { viewer, userId, joinedContests, contests, logs } = usePlayground()
  const [scenario, setScenario] = useScenario('leaderboard')
  const scope = location.pathname.includes('/yearly/') ? 'year' : location.pathname.endsWith('/all-time') ? 'all' : 'contest'
  const params = new URLSearchParams(location.search)
  const scenarioDate = scenario === 'upcoming' ? '2026-08-28' : scenario === 'ended' ? '2026-10-01' : ['closed', 'closed-participant'].includes(scenario) ? '2026-09-24' : '2026-09-05'
  const asOf = /^\d{4}-\d{2}-\d{2}$/.test(params.get('asOf') ?? '') ? params.get('asOf')! : scenarioDate
  const contest = contests.find(item => item.id === (contestId ?? 'round5'))
  const status = contest ? contestStatus(contest, asOf) : 'ended'
  const joined = Boolean(contest && viewer !== 'guest' && joinedContests.includes(contest.id))
  const registrationOpen = Boolean(contest && status !== 'ended' && asOf <= contest.registrationDeadline)
  const isError = scenario === 'error'
  const isEmpty = scenario === 'empty' || (scope === 'contest' && status === 'upcoming') || (scope === 'year' && !['2026', '2025', '2024'].includes(year))
  const methods = usePaperForm({ defaultValues: ({ language: 'all', activity: 'all', query: '', year } as BoardFilters) })
  const { language = 'all', activity = 'all', query = '' } = useStore(methods.store, state => state.values)
  const [filterExpansion, setFilterExpansion] = useState({ scope, open: scope === 'year' })
  if (filterExpansion.scope !== scope) setFilterExpansion({ scope, open: scope === 'year' })
  const filtersOpen = filterExpansion.open
  const showFilters = scope === 'year' || (!isError && !isEmpty)
  const filtersId = useId()
  const filtersToggle = useRef<HTMLButtonElement>(null)
  const appliedFilters = [language !== 'all' ? language : '', activity !== 'all' ? activity : '', query.trim() ? `Name: ${query.trim()}` : ''].filter(Boolean)
  const clearFilters = () => {
    methods.reset({ language: 'all', activity: 'all', query: '', year }, { keepDefaultValues: true })
    filtersToggle.current?.focus()
  }
  const [pageState, setPageState] = useState({ key: '', page: 1 })
  const [focusRequest, setFocusRequest] = useState(0)
  const selfRowId = `standing-${userId}`
  useEffect(() => { if (focusRequest) document.getElementById(selfRowId)?.focus() }, [focusRequest, selfRowId])

  const boardHref = (path: string) => `${path}${location.search}`
  const contestHref = (suffix = '') => `/contests/${contest?.id ?? 'round5'}${suffix}?asOf=${asOf}`
  const registrationHref = contestHref('/registration')
  const signInHref = `/sign-in?next=${encodeURIComponent(registrationHref)}`
  const factor = scope === 'all' ? 61 : scope === 'year' ? ({ '2026': 8, '2025': 12, '2024': 10 }[year] ?? 0) : 1
  const useRoundFixture = scope !== 'contest' || contest?.id === 'round5'
  const scores = new Map<string, number>()
  if (useRoundFixture) {
    const slots = [0, 1, 2, 3].filter(index => (language === 'all' || language === (index < 2 ? 'Japanese' : 'Spanish')) && (activity === 'all' || activity === (index % 2 === 0 ? 'Reading' : 'Listening')))
    leaderboardPeople.forEach(person => scores.set(person.userId, slots.reduce((sum, index) => sum + person.scores[index], 0) * factor))
  }
  // The board fixtures include initial entries. Apply local changes as deltas, not a second copy of their scores.
  const collectLogScores = (entries: SampleLog[], direction: number) => entries.forEach(log => {
    if (!scoreMatches(log, language, activity)) return
    if (scope === 'year' && !log.date.startsWith(year)) return
    if (scope === 'contest' && log.date > asOf) return
    const matching = log.submissions.filter(submission => scope === 'contest' ? submission.contestId === contest?.id : contests.some(item => item.id === submission.contestId && item.scope === 'official'))
    const score = matching.reduce((sum, submission) => sum + submission.score, 0)
    if (score) scores.set(log.userId, (scores.get(log.userId) ?? 0) + direction * score)
  })
  if (useRoundFixture) collectLogScores(initialLogs, -1)
  collectLogScores(logs, 1)
  const ranked = Array.from(scores, ([id, score]) => ({ userId: id, name: users.find(person => person.id === id)?.name ?? id, score: Math.max(0, Math.round(score * 10) / 10), rank: 0, tied: false }))
    .filter(row => scope === 'contest' || row.score > 0)
    .sort((left, right) => right.score - left.score || left.name.localeCompare(right.name))
  ranked.forEach((row, index) => {
    row.rank = index > 0 && row.score === ranked[index - 1].score ? ranked[index - 1].rank : index + 1
    row.tied = (index > 0 && row.score === ranked[index - 1].score) || (index < ranked.length - 1 && row.score === ranked[index + 1].score)
  })
  const visible = ranked.filter(row => row.name.toLocaleLowerCase().includes(query.trim().toLocaleLowerCase()))
  const filterKey = `${scope}:${contest?.id}:${year}:${scenario}:${language}:${activity}:${query}`
  const totalPages = Math.max(1, Math.ceil(visible.length / pageSize))
  const page = pageState.key === filterKey ? Math.min(pageState.page, totalPages) : 1
  const self = ranked.find(row => row.userId === userId)
  const showSelf = viewer !== 'guest' && (scope !== 'contest' || joined) && !isError && !isEmpty && self
  const findMyRow = () => {
    const index = ranked.findIndex(row => row.userId === userId)
    methods.setFieldValue('query', '')
    setPageState({ key: `${scope}:${contest?.id}:${year}:${scenario}:${language}:${activity}:`, page: Math.floor(index / pageSize) + 1 })
    setFocusRequest(value => value + 1)
  }
  const columns: TableColumn<Standing>[] = [
    { id: 'rank', header: 'Rank', width: '4.5rem', cell: row => <span className="text-muted [&_abbr]:me-0.5 [&_abbr]:no-underline">{row.tied ? <abbr title="Tied rank">T</abbr> : null}{row.rank}</span> },
    { id: 'participant', header: 'Participant', rowHeader: true, cell: row => <div className="flex flex-wrap items-center gap-2"><Link className="text-link" to={`/users/${row.userId}${scope === 'contest' ? `?contest=${contest?.id}` : scope === 'year' ? `?year=${year}` : ''}`}>{row.name}</Link>{row.userId === userId && viewer !== 'guest' && (scope !== 'contest' || joined) ? <span className="text-2xs font-regular text-muted">You</span> : null}</div> },
    { id: 'score', header: 'Score', align: 'end', width: '7rem', cell: row => formatNumber(row.score) },
  ]
  const recent = logs.filter(log => log.date <= asOf && log.submissions.some(submission => submission.contestId === contest?.id)).sort((left, right) => right.date.localeCompare(left.date)).slice(0, 3)
  const title = scope === 'contest' ? contest?.title ?? 'Contest not found' : 'Official standings'
  const moderators = (contest?.moderatorIds ?? []).map(id => users.find(person => person.id === id)).filter(person => person !== undefined)

  const roundAction = scope === 'contest' && contest ? status === 'live' && joined ? <Link className={buttonClassName({ variant: 'outline' })} to={`/logs/new?contest=${contest.id}&asOf=${asOf}`}>Log activity</Link> : registrationOpen ? <Link className={buttonClassName()} to={viewer === 'guest' ? signInHref : registrationHref}>{viewer === 'guest' ? 'Sign in to join' : joined ? 'Update registration' : 'Join this round'}</Link> : <Link className={buttonClassName({ variant: 'outline' })} to="/contests/official">Browse contests</Link> : null

  return <>
    <PageHeader className="standings-header mb-5" title="Leaderboard" actions={roundAction} />
    <Tabbar label="Leaderboard period" links={[
      { id: 'contest', label: 'Latest official', href: boardHref('/leaderboard/latest'), current: scope === 'contest' && !contestId },
      ...(contestId ? [{ id: 'selected', label: contest?.title ?? 'Selected contest', href: boardHref(`/contests/${contestId}/leaderboard`), current: true }] : []),
      { id: 'year', label: 'Yearly', href: boardHref('/leaderboard/yearly/2026'), current: scope === 'year' },
      { id: 'all', label: 'All time', href: boardHref('/leaderboard/all-time'), current: scope === 'all' },
    ]} renderLink={({ href, ...props }) => <Link to={href} {...props} />} />
    {!contest && scope === 'contest' ? <EmptyState headingLevel={2} title="This contest is not available." description="Check the link or find another round." actions={<Link className={buttonClassName()} to="/contests/official">Browse contests</Link>} /> : <>
      <div className="mt-5 grid grid-cols-workspace items-start gap-8 max-lg:grid-cols-workspace-narrow max-lg:gap-6 max-tablet:grid-cols-single">
        <div className="min-w-0">
        <section className="min-w-0 border border-rule bg-paper" aria-label="Standings">
          <SectionHeading
            level={2}
            className="m-0 gap-y-0 p-5 max-md:gap-x-2 max-md:px-4 max-md:pb-4 max-md:pt-5 [&_h2]:text-heading max-md:[&_h2]:text-lg [&>div+div]:ms-auto max-md:[&_div:has(>h2)]:gap-x-2 max-md:[&_.paper-button]:px-1"
            title={title}
            status={scope === 'contest' && contest ? <StatusTag tone={status === 'live' ? 'success' : status === 'upcoming' ? 'information' : 'neutral'}>{status === 'ended' ? 'Complete' : status === 'upcoming' ? 'Upcoming' : 'Live now'}</StatusTag> : null}
            description={scope === 'contest' && contest ? formatDateRange(contestStart(contest.start), contestEnd(contest.end)) : null}
            actions={showFilters ? <Button ref={filtersToggle} variant="ghost" aria-expanded={filtersOpen} aria-controls={filtersId} onClick={() => setFilterExpansion({ scope, open: !filtersOpen })} trailingIcon={<ChevronDownIcon className={iconClassName('compact', filtersOpen ? 'rotate-flip' : undefined)} />}>Filters{appliedFilters.length ? ` (${appliedFilters.length})` : ''}</Button> : null}
          />
          {showFilters ? <><div id={filtersId} hidden={!filtersOpen}><methods.AppForm><form className={`grid grid-cols-standings-filters gap-3 px-5 pb-5 pt-0 max-lg:grid-cols-pair max-lg:[&>.paper-field:nth-child(3)]:col-span-full max-md:px-4 max-md:pb-4 max-xs:grid-cols-single max-xs:[&>.paper-field:nth-child(3)]:col-auto ${scope === 'year' ? ' grid-cols-standings-filters-yearly max-lg:[&>.paper-field:nth-child(3)]:col-auto' : ''}`} onSubmit={event => event.preventDefault()}>
            {scope === 'year' ? <methods.AppField name="year">{field => <field.Select label="Year" value={year} options={['2026', '2025', '2024'].map(value => ({ value, label: value }))} onInput={event => navigate(`/leaderboard/yearly/${event.currentTarget.value}${location.search}`)} />}</methods.AppField> : null}
            <methods.AppField name="language">{field => <field.Select label="Language" options={[{ value: 'all', label: 'All languages' }, ...supportedLanguages.map(value => ({ value, label: value }))]} />}</methods.AppField>
            <methods.AppField name="activity">{field => <field.Select label="Activity" options={[{ value: 'all', label: 'All activities' }, { value: 'Reading', label: 'Reading' }, { value: 'Listening', label: 'Listening' }]} />}</methods.AppField>
            <methods.AppField name="query">{field => <field.Input label="Find a participant" type="search" placeholder="Search by name" />}</methods.AppField>
          </form></methods.AppForm></div><AppliedFilters className="px-5 pb-3 pt-0 max-md:px-4" filters={appliedFilters} onClear={clearFilters} /></> : null}
          {showSelf ? <div className="flex items-center justify-between gap-3 border-t border-rule py-1 pe-3 ps-5 [&_p]:m-0 [&_p]:flex [&_p]:flex-wrap [&_p]:gap-x-4 [&_p]:gap-y-0.5 [&_p]:text-sm [&_strong]:whitespace-nowrap [&_span]:text-muted [&_.paper-button]:shrink-0 max-md:py-1.5 max-md:pe-2 max-md:ps-4 max-md:[&_p]:block max-md:[&_p]:text-fine max-md:[&_span]:mt-0.5 max-md:[&_span]:block"><p><strong>Your rank: {self.tied ? 'T' : '#'}{self.rank}</strong><span>{formatNumber(self.score)} points</span></p><Button variant="ghost" onClick={findMyRow}>Jump to me</Button></div> : null}
          {isError || isEmpty || visible.length === 0 ? <EmptyState
            variant="plain"
            className="border-t border-rule px-5 max-md:px-4"
            aria-live="polite"
            headingLevel={3}
            title={isError ? 'Standings could not be loaded' : isEmpty || ranked.length === 0 ? scope === 'contest' ? 'The first page is yours to write' : 'No official scores for this period' : 'No participants match that name'}
            description={isError ? 'Try loading the standings again. The available contest details are still shown alongside.' : isEmpty || ranked.length === 0 ? scope === 'contest' ? status === 'upcoming' ? 'Scores begin when the round starts. Join now, then return after your first eligible activity.' : 'No participants are listed yet. Join the round and log your first activity.' : 'Scores will appear here when eligible activity is logged in this period.' : 'Try another display name. Searching does not change anyone’s rank.'}
            actions={isError ? <Button onClick={() => setScenario('participant')}>Try again</Button> : isEmpty || ranked.length === 0 ? <Link className={buttonClassName({ variant: 'outline' })} to="/contests/official">Browse contests</Link> : <Button variant="outline" onClick={() => { methods.setFieldValue('query', ''); if (filtersOpen) document.getElementsByName('query')[0]?.focus(); else filtersToggle.current?.focus() }}>Clear search</Button>}
          /> : <>
            <Table captionVisibility="screen-reader" caption={`${language === 'all' ? 'All languages' : language} · ${activity === 'all' ? 'All activities' : activity}${query.trim() ? ' · Name search applied' : ''}`} columns={columns} rows={visible.slice((page - 1) * pageSize, page * pageSize)} getRowKey={row => row.userId} minWidth="17rem" tableClassName="standings-table tabular-nums [&_td]:px-5 [&_th]:px-5 max-md:[&_td]:px-4 max-md:[&_th]:px-4 [&_tr[aria-current=true]]:bg-action-soft [&_tr:focus]:outline-ring [&_tr:focus]:outline-focus [&_tr:focus]:-outline-offset-ring" getRowProps={row => ({ id: `standing-${row.userId}`, tabIndex: -1, 'aria-current': showSelf && row.userId === userId ? 'true' : undefined })} />
            <div className="standings-pagination flex flex-wrap items-center justify-between gap-x-4 gap-y-2 px-5 py-3 max-md:px-4 [&>p]:m-0 [&>p]:text-fine"><p className="text-muted" role="status">{(page - 1) * pageSize + 1}–{Math.min(page * pageSize, visible.length)} of {visible.length} {query.trim() ? 'matches' : 'participants'}</p><Pagination totalPages={totalPages} currentPage={page} onPageChange={next => setPageState({ key: filterKey, page: next })} label="Standings pages" /></div>
          </>}
        </section>
        <p className="mb-0 mt-4 max-w-measure-footnote text-fine leading-165 text-muted">Leaderboard updates may take a few seconds. <abbr title="Tied rank">T</abbr> means a tie.</p>
        </div>
        <aside className="min-w-0 pt-1 text-sm leading-170 [&_h2]:mb-3 [&_h2]:mt-0 [&_p]:mb-5 [&_p]:mt-0 [&>section+section]:mt-7 [&>section+section]:border-t [&>section+section]:border-rule [&>section+section]:pt-6 max-tablet:grid max-tablet:grid-cols-pair max-tablet:gap-7 max-tablet:pt-2 max-tablet:[&>section+section]:m-0 max-tablet:[&>section+section]:border-0 max-tablet:[&>section+section]:p-0 max-md:grid-cols-single max-md:gap-6">
          <section aria-labelledby="standings-about-heading"><h2 id="standings-about-heading" className={typography({ role: 'component' })}>{scope === 'contest' ? 'Contest conditions' : 'About these standings'}</h2>
            {scope === 'contest' && contest ? <><DescriptionList className="contest-facts mb-5" items={[
              { label: 'Languages', value: contest.languages.includes('All languages') ? 'All languages allowed' : `${contest.languages.join(', ')} only` },
              { label: 'Activities', value: contest.activities.includes('Reading') && contest.activities.includes('Listening') ? 'All activities allowed' : `${contest.activities.join(', ')} only` },
              { label: 'Moderators', value: moderators.length ? moderators.map((person, index) => <span key={person.id}>{index ? ', ' : ''}<Link className="text-link" to={`/users/${person.id}`}>{person.name}</Link></span>) : 'Not listed' },
              { label: 'Registration', value: registrationOpen ? `Until ${formatDateTime(contestEnd(contest.registrationDeadline))}` : 'Closed to new participants' },
            ]} /><Link className="text-link" to={contestHref()}>Contest details &amp; schedule</Link></> : <p>Includes eligible official activity for this period. Community contest scores are not added together to make this board.</p>}
          </section>
          {scope === 'contest' && status === 'live' && !isError && !isEmpty && recent.length > 0 ? <section><h2 className={typography({ role: 'component' })}>Recent activity</h2><ul className="m-0 list-none p-0 [&_li+li]:mt-4 [&_a]:flex [&_a]:justify-between [&_a]:gap-4 [&_a]:no-underline [&_a:hover]:underline [&_small]:mt-0.5 [&_small]:block">{recent.map(log => <li key={log.id}><Link className="text-link" to={`/logs/${log.id}?contest=${contest?.id}`}><span>{users.find(person => person.id === log.userId)?.name ?? log.userId}</span><span>+{formatNumber(log.submissions.find(submission => submission.contestId === contest?.id)?.score ?? 0)}</span></Link><small className="text-muted">{log.activity} · {formatDate(log.date)}</small></li>)}</ul></section> : null}
        </aside>
      </div>
    </>}
  </>
}
