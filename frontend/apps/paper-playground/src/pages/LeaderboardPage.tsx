import * as stylex from '@stylexjs/stylex'
import { useStore } from '@tanstack/react-form'
import { useEffect, useId, useRef, useState } from 'react'
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom'
import { AppliedFilters, Button, buttonStyles, DescriptionList, EmptyState, iconSizes, linkStyles, PageHeader, Pagination, SectionHeading, StatusTag, Tabbar, Table, typography, usePaperForm, type TableColumn } from 'paper-ui'
import { size } from 'paper-ui/tokens.stylex'
import { ChevronDownIcon } from 'paper-ui/icons'
import { contestStatus, formatDate, formatNumber, initialLogs, leaderboardPeople, supportedLanguages, users, type SampleLog } from '../data'
import { contestEnd, contestStart, formatDateRange, formatDateTime } from '../dates'
import { usePlayground, useScenario } from '../state'
import { contestsStyles } from './contests.styles'
import { utilities } from '../utilities.styles'
import { appStyles } from '../app.styles'
import { descendantStyles } from '../descendants.styles'

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
    { id: 'rank', header: 'Rank', width: size[72], cell: row => <span {...stylex.props(contestsStyles.standingsRank)}>{row.tied ? <abbr {...stylex.props(descendantStyles.standingsRankU_uAbbr)} title="Tied rank">T</abbr> : null}{row.rank}</span> },
    { id: 'participant', header: 'Participant', rowHeader: true, cell: row => <div {...stylex.props(utilities.tw_flex, utilities.tw_flex_wrap, utilities.tw_items_center, utilities.tw_gap_2)}><Link {...stylex.props(linkStyles.base)} to={`/users/${row.userId}${scope === 'contest' ? `?contest=${contest?.id}` : scope === 'year' ? `?year=${year}` : ''}`}>{row.name}</Link>{row.userId === userId && viewer !== 'guest' && (scope !== 'contest' || joined) ? <span {...stylex.props(contestsStyles.standingsYou)}>You</span> : null}</div> },
    { id: 'score', header: 'Score', align: 'end', width: size[112], cell: row => formatNumber(row.score) },
  ]
  const recent = logs.filter(log => log.date <= asOf && log.submissions.some(submission => submission.contestId === contest?.id)).sort((left, right) => right.date.localeCompare(left.date)).slice(0, 3)
  const title = scope === 'contest' ? contest?.title ?? 'Contest not found' : 'Official standings'
  const moderators = (contest?.moderatorIds ?? []).map(id => users.find(person => person.id === id)).filter(person => person !== undefined)

  const roundAction = scope === 'contest' && contest ? status === 'live' && joined ? <Link {...stylex.props(...buttonStyles({ variant: 'outline' }))} to={`/logs/new?contest=${contest.id}&asOf=${asOf}`}>Log activity</Link> : registrationOpen ? <Link {...stylex.props(...buttonStyles())} to={viewer === 'guest' ? signInHref : registrationHref}>{viewer === 'guest' ? 'Sign in to join' : joined ? 'Update registration' : 'Join this round'}</Link> : <Link {...stylex.props(...buttonStyles({ variant: 'outline' }))} to="/contests/official">Browse contests</Link> : null

  return <>
    <PageHeader xstyle={[contestsStyles.standingsHeader]} title="Leaderboard" actions={roundAction} />
    <Tabbar label="Leaderboard period" links={[
      { id: 'contest', label: 'Latest official', href: boardHref('/leaderboard/latest'), current: scope === 'contest' && !contestId },
      ...(contestId ? [{ id: 'selected', label: contest?.title ?? 'Selected contest', href: boardHref(`/contests/${contestId}/leaderboard`), current: true }] : []),
      { id: 'year', label: 'Yearly', href: boardHref('/leaderboard/yearly/2026'), current: scope === 'year' },
      { id: 'all', label: 'All time', href: boardHref('/leaderboard/all-time'), current: scope === 'all' },
    ]} renderLink={({ href, ...props }) => <Link to={href} {...props} />} />
    {!contest && scope === 'contest' ? <EmptyState headingLevel={2} title="This contest is not available." description="Check the link or find another round." actions={<Link {...stylex.props(...buttonStyles())} to="/contests/official">Browse contests</Link>} /> : <>
      <div {...stylex.props(contestsStyles.standingsWorkspace)}>
        <div {...stylex.props(contestsStyles.standingsMain)}>
        <section {...stylex.props(contestsStyles.standingsSheet)} aria-label="Standings">
          <SectionHeading
            level={2}
            xstyle={[contestsStyles.standingsContext]}
            title={title}
            status={scope === 'contest' && contest ? <StatusTag tone={status === 'live' ? 'success' : status === 'upcoming' ? 'information' : 'neutral'}>{status === 'ended' ? 'Complete' : status === 'upcoming' ? 'Upcoming' : 'Live now'}</StatusTag> : null}
            description={scope === 'contest' && contest ? formatDateRange(contestStart(contest.start), contestEnd(contest.end)) : null}
            actions={showFilters ? <Button ref={filtersToggle} variant="ghost" aria-expanded={filtersOpen} aria-controls={filtersId} onClick={() => setFilterExpansion({ scope, open: !filtersOpen })} trailingIcon={<ChevronDownIcon {...stylex.props(iconSizes.compact, filtersOpen && utilities.tw_rotate_180)} />}>Filters{appliedFilters.length ? ` (${appliedFilters.length})` : ''}</Button> : null}
          />
          {showFilters ? <><div id={filtersId} hidden={!filtersOpen}><methods.AppForm><form {...stylex.props(contestsStyles.standingsFilters, scope === 'year' && contestsStyles.standingsFilters_yearly)} onSubmit={event => event.preventDefault()}>
            {scope === 'year' ? <methods.AppField name="year">{field => <field.Select label="Year" value={year} options={['2026', '2025', '2024'].map(value => ({ value, label: value }))} onInput={event => navigate(`/leaderboard/yearly/${event.currentTarget.value}${location.search}`)} />}</methods.AppField> : null}
            <methods.AppField name="language">{field => <field.Select label="Language" options={[{ value: 'all', label: 'All languages' }, ...supportedLanguages.map(value => ({ value, label: value }))]} />}</methods.AppField>
            <methods.AppField name="activity">{field => <field.Select label="Activity" options={[{ value: 'all', label: 'All activities' }, { value: 'Reading', label: 'Reading' }, { value: 'Listening', label: 'Listening' }]} />}</methods.AppField>
            <methods.AppField name="query">{field => <field.Input label="Find a participant" type="search" placeholder="Search by name" />}</methods.AppField>
          </form></methods.AppForm></div><AppliedFilters xstyle={[contestsStyles.standingsAppliedFilters]} filters={appliedFilters} onClear={clearFilters} /></> : null}
          {showSelf ? <div {...stylex.props(contestsStyles.standingsPersonal)}><p {...stylex.props(descendantStyles.standingsPersonalU_uP)}><strong {...stylex.props(descendantStyles.standingsPersonalU_uStrong)}>Your rank: {self.tied ? 'T' : '#'}{self.rank}</strong><span {...stylex.props(descendantStyles.standingsPersonalU_uSpan)}>{formatNumber(self.score)} points</span></p><Button variant="ghost" onClick={findMyRow}>Jump to me</Button></div> : null}
          {isError || isEmpty || visible.length === 0 ? <EmptyState
            variant="plain"
            xstyle={[contestsStyles.standingsEmpty]}
            aria-live="polite"
            headingLevel={3}
            title={isError ? 'Standings could not be loaded' : isEmpty || ranked.length === 0 ? scope === 'contest' ? 'The first page is yours to write' : 'No official scores for this period' : 'No participants match that name'}
            description={isError ? 'Try loading the standings again. The available contest details are still shown alongside.' : isEmpty || ranked.length === 0 ? scope === 'contest' ? status === 'upcoming' ? 'Scores begin when the round starts. Join now, then return after your first eligible activity.' : 'No participants are listed yet. Join the round and log your first activity.' : 'Scores will appear here when eligible activity is logged in this period.' : 'Try another display name. Searching does not change anyone’s rank.'}
            actions={isError ? <Button onClick={() => setScenario('participant')}>Try again</Button> : isEmpty || ranked.length === 0 ? <Link {...stylex.props(...buttonStyles({ variant: 'outline' }))} to="/contests/official">Browse contests</Link> : <Button variant="outline" onClick={() => { methods.setFieldValue('query', ''); if (filtersOpen) document.getElementsByName('query')[0]?.focus(); else filtersToggle.current?.focus() }}>Clear search</Button>}
          /> : <>
            <Table captionVisibility="screen-reader" caption={`${language === 'all' ? 'All languages' : language} · ${activity === 'all' ? 'All activities' : activity}${query.trim() ? ' · Name search applied' : ''}`} columns={columns} rows={visible.slice((page - 1) * pageSize, page * pageSize)} getRowKey={row => row.userId} minWidth={size[272]} getRowProps={row => ({ id: `standing-${row.userId}`, tabIndex: -1, 'aria-current': showSelf && row.userId === userId ? 'true' : undefined })} />
            <div {...stylex.props(contestsStyles.standingsPagination)}><p {...stylex.props(appStyles.muted, descendantStyles.standingsPaginationU_uP)} role="status">{(page - 1) * pageSize + 1}–{Math.min(page * pageSize, visible.length)} of {visible.length} {query.trim() ? 'matches' : 'participants'}</p><Pagination totalPages={totalPages} currentPage={page} onPageChange={next => setPageState({ key: filterKey, page: next })} label="Standings pages" /></div>
          </>}
        </section>
        <p {...stylex.props(contestsStyles.standingsFootnote, appStyles.muted)}>Leaderboard updates may take a few seconds. <abbr title="Tied rank">T</abbr> means a tie.</p>
        </div>
        <aside {...stylex.props(contestsStyles.standingsAside)}>
          <section {...stylex.props(descendantStyles.standingsAsideU_uSection)} aria-labelledby="standings-about-heading"><h2 id="standings-about-heading" {...stylex.props(typography.component, descendantStyles.standingsAsideU_uH2)}>{scope === 'contest' ? 'Contest conditions' : 'About these standings'}</h2>
            {scope === 'contest' && contest ? <><DescriptionList xstyle={[contestsStyles.contestFacts]} items={[
              { label: 'Languages', value: contest.languages.includes('All languages') ? 'All languages allowed' : `${contest.languages.join(', ')} only` },
              { label: 'Activities', value: contest.activities.includes('Reading') && contest.activities.includes('Listening') ? 'All activities allowed' : `${contest.activities.join(', ')} only` },
              { label: 'Moderators', value: moderators.length ? moderators.map((person, index) => <span key={person.id}>{index ? ', ' : ''}<Link {...stylex.props(linkStyles.base)} to={`/users/${person.id}`}>{person.name}</Link></span>) : 'Not listed' },
              { label: 'Registration', value: registrationOpen ? `Until ${formatDateTime(contestEnd(contest.registrationDeadline))}` : 'Closed to new participants' },
            ]} /><Link {...stylex.props(linkStyles.base)} to={contestHref()}>Contest details &amp; schedule</Link></> : <p {...stylex.props(descendantStyles.standingsAsideU_uP)}>Includes eligible official activity for this period. Community contest scores are not added together to make this board.</p>}
          </section>
          {scope === 'contest' && status === 'live' && !isError && !isEmpty && recent.length > 0 ? <section {...stylex.props(descendantStyles.standingsAsideU_uSection)}><h2 {...stylex.props(typography.component, descendantStyles.standingsAsideU_uH2)}>Recent activity</h2><ul {...stylex.props(contestsStyles.standingsActivity)}>{recent.map(log => <li {...stylex.props(descendantStyles.standingsActivityU_uLi)} key={log.id}><Link {...stylex.props(linkStyles.base)} to={`/logs/${log.id}?contest=${contest?.id}`}><span>{users.find(person => person.id === log.userId)?.name ?? log.userId}</span><span>+{formatNumber(log.submissions.find(submission => submission.contestId === contest?.id)?.score ?? 0)}</span></Link><small {...stylex.props(appStyles.muted, descendantStyles.standingsActivityU_uSmall)}>{log.activity} · {formatDate(log.date)}</small></li>)}</ul></section> : null}
        </aside>
      </div>
    </>}
  </>
}
