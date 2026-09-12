import { useEffect, useState } from 'react'
import { FormProvider, useForm, useWatch } from 'react-hook-form'
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom'
import { Button, Input, Pagination, Select, Table, Tabbar, buttonClassName, type TableColumn } from 'paper-ui'
import { contestStatus, formatDate, formatNumber, initialLogs, leaderboardPeople, supportedLanguages, users, type SampleLog } from '../data'
import { usePlayground, useScenario } from '../state'
import './contests.css'

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
  const methods = useForm<BoardFilters>({ defaultValues: { language: 'all', activity: 'all', query: '', year } })
  const { language = 'all', activity = 'all', query = '' } = useWatch({ control: methods.control })
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
    methods.setValue('query', '')
    setPageState({ key: `${scope}:${contest?.id}:${year}:${scenario}:${language}:${activity}:`, page: Math.floor(index / pageSize) + 1 })
    setFocusRequest(value => value + 1)
  }
  const columns: TableColumn<Standing>[] = [
    { id: 'rank', header: 'Rank', width: '4.5rem', cell: row => <span className="standings-rank">{row.tied ? <abbr title="Tied rank">T</abbr> : null}{row.rank}</span> },
    { id: 'participant', header: 'Participant', rowHeader: true, cell: row => <div className="flex flex-wrap items-center gap-2"><Link className="text-link" to={`/users/${row.userId}${scope === 'contest' ? `?contest=${contest?.id}` : scope === 'year' ? `?year=${year}` : ''}`}>{row.name}</Link>{row.userId === userId && viewer !== 'guest' && (scope !== 'contest' || joined) ? <span className="standings-you">You</span> : null}</div> },
    { id: 'score', header: 'Score', align: 'end', width: '7rem', cell: row => formatNumber(row.score) },
  ]
  const recent = logs.filter(log => log.date <= asOf && log.submissions.some(submission => submission.contestId === contest?.id)).sort((left, right) => right.date.localeCompare(left.date)).slice(0, 3)
  const title = scope === 'contest' ? contest?.title ?? 'Contest not found' : scope === 'year' ? `${year} official standings` : 'All-time official standings'

  const roundAction = scope === 'contest' && contest ? status === 'live' && joined ? <Link className={buttonClassName({ variant: 'outline' })} to={`/logs/new?contest=${contest.id}&asOf=${asOf}`}>Log activity</Link> : registrationOpen ? <Link className={buttonClassName()} to={viewer === 'guest' ? signInHref : registrationHref}>{viewer === 'guest' ? 'Sign in to join' : joined ? 'Update registration' : 'Join this round'}</Link> : <Link className={buttonClassName({ variant: 'outline' })} to="/contests/official">Browse contests</Link> : null

  return <>
    <header className="page-header competition-header"><h1 className="paper-type-page">Leaderboard</h1>{roundAction}</header>
    <Tabbar label="Leaderboard period" links={[
      { id: 'contest', label: 'Latest official', href: boardHref('/leaderboard/latest'), current: scope === 'contest' && !contestId },
      ...(contestId ? [{ id: 'selected', label: contest?.title ?? 'Selected contest', href: boardHref(`/contests/${contestId}/leaderboard`), current: true }] : []),
      { id: 'year', label: 'Yearly', href: boardHref('/leaderboard/yearly/2026'), current: scope === 'year' },
      { id: 'all', label: 'All time', href: boardHref('/leaderboard/all-time'), current: scope === 'all' },
    ]} renderLink={({ href, ...props }) => <Link to={href} {...props} />} />
    {!contest && scope === 'contest' ? <section className="empty-state"><h2 className="paper-type-section">This contest is not available.</h2><p>Check the link or find another round.</p><Link className={buttonClassName()} to="/contests/official">Browse contests</Link></section> : <>
      <div className="standings-workspace">
        <section className="standings-sheet" aria-label="Standings">
          <div className="standings-context">
            <div><div className="standings-context__title"><h2 className="paper-type-section">{title}</h2>{scope === 'contest' && contest ? <span className="status-tag" data-status={status}>{status === 'ended' ? 'Complete' : status === 'upcoming' ? 'Upcoming' : 'Live now'}</span> : null}</div>{scope === 'contest' && contest ? <p className="muted">{formatDate(contest.start)} – {formatDate(contest.end)} (UTC)</p> : null}</div>
            {scope === 'year' ? <FormProvider {...methods}><div className="standings-year"><Select name="year" label="Year" value={year} options={['2026', '2025', '2024'].map(value => ({ value, label: value }))} onInput={event => navigate(`/leaderboard/yearly/${event.currentTarget.value}${location.search}`)} /></div></FormProvider> : null}
          </div>
          {!isError && !isEmpty ? <FormProvider {...methods}><form className="standings-filters" onSubmit={event => event.preventDefault()}>
            <Select name="language" label="Language" options={[{ value: 'all', label: 'All languages' }, ...supportedLanguages.map(value => ({ value, label: value }))]} />
            <Select name="activity" label="Activity" options={[{ value: 'all', label: 'All activities' }, { value: 'Reading', label: 'Reading' }, { value: 'Listening', label: 'Listening' }]} />
            <Input name="query" label="Find a participant" type="search" placeholder="Search by name" />
          </form></FormProvider> : null}
          {showSelf ? <div className="standings-personal"><p><strong>Your rank: {self.tied ? 'T' : '#'}{self.rank}</strong><span>{formatNumber(self.score)} points</span></p><Button variant="ghost" onClick={findMyRow}>Find my row</Button></div> : null}
          {isError || isEmpty || visible.length === 0 ? <section className="empty-state" aria-live="polite">
            <h2 className="paper-type-section">{isError ? 'Standings could not be loaded' : isEmpty || ranked.length === 0 ? scope === 'contest' ? 'The first page is yours to write' : 'No official scores for this period' : 'No participants match that name'}</h2>
            <p>{isError ? 'Try loading the standings again. The available contest details are still shown alongside.' : isEmpty || ranked.length === 0 ? scope === 'contest' ? status === 'upcoming' ? 'Scores begin when the round starts. Join now, then return after your first eligible activity.' : 'No participants are listed yet. Join the round and log your first activity.' : 'Scores will appear here when eligible activity is logged in this period.' : 'Try another display name. Searching does not change anyone’s rank.'}</p>
            {isError ? <Button onClick={() => setScenario('participant')}>Try again</Button> : isEmpty || ranked.length === 0 ? <Link className={buttonClassName({ variant: 'outline' })} to="/contests/official">Browse contests</Link> : <Button variant="outline" onClick={() => { methods.setValue('query', ''); methods.setFocus('query') }}>Clear search</Button>}
          </section> : <>
            <Table captionVisibility="screen-reader" caption={`${language === 'all' ? 'All languages' : language} · ${activity === 'all' ? 'All activities' : activity}${query.trim() ? ' · Name search applied' : ''}`} columns={columns} rows={visible.slice((page - 1) * pageSize, page * pageSize)} getRowKey={row => row.userId} minWidth="17rem" tableClassName="standings-table" getRowProps={row => ({ id: `standing-${row.userId}`, tabIndex: -1, 'aria-current': showSelf && row.userId === userId ? 'true' : undefined })} />
            <div className="standings-pagination"><p className="muted" role="status">{(page - 1) * pageSize + 1}–{Math.min(page * pageSize, visible.length)} of {visible.length} {query.trim() ? 'matches' : 'participants'}</p><Pagination totalPages={totalPages} currentPage={page} onPageChange={next => setPageState({ key: filterKey, page: next })} label="Standings pages" /></div>
          </>}
        </section>
        <aside className="standings-aside">
          <section><h2 className="paper-type-component">{scope === 'contest' ? 'Taking part' : 'About these standings'}</h2>
            {scope === 'contest' && contest ? <><p>{joined && status === 'live' ? 'You’re registered for this round. Submit your reading and listening until the round ends.' : status === 'ended' ? 'This round has finished. Your results are here whenever you want to look back.' : status === 'upcoming' ? 'Join now and start logging when the round begins.' : !registrationOpen ? 'This round is still running for registered participants.' : 'Join the round, then submit eligible activity as you go.'}</p><dl className="contest-facts"><div><dt>Registration</dt><dd>{registrationOpen ? `Until ${formatDate(contest.registrationDeadline)}, 23:59 UTC` : 'Closed to new participants'}</dd></div><div><dt>Participants</dt><dd>{isError ? 'Unavailable' : isEmpty ? 0 : ranked.length}</dd></div></dl><Link className="text-link" to={contestHref()}>Contest details &amp; schedule</Link></> : <><p>Includes eligible official activity for this period. Community contest scores are not added together to make this board.</p><p>Language and activity filters recalculate the scores and ranks. Search only narrows the names.</p><Link className="text-link" to="/guide/scoring">Read the scoring rules</Link></>}
          </section>
          {scope === 'contest' && status === 'live' && !isError && !isEmpty && recent.length > 0 ? <section><h2 className="paper-type-component">Recent activity</h2><ul className="standings-activity">{recent.map(log => <li key={log.id}><Link className="text-link" to={`/logs/${log.id}?contest=${contest?.id}`}><span>{users.find(person => person.id === log.userId)?.name ?? log.userId}</span><span>+{formatNumber(log.submissions.find(submission => submission.contestId === contest?.id)?.score ?? 0)}</span></Link><small className="muted">{log.activity} · {formatDate(log.date)}</small></li>)}</ul></section> : null}
        </aside>
      </div>
      <p className="standings-footnote muted">Language and activity filters recalculate ranks; name search keeps them. <abbr title="Tied rank">T</abbr> means a tie. <Link className="text-link" to="/guide/scoring">How scoring works</Link></p>
    </>}
  </>
}
