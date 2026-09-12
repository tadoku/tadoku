import { useEffect, useRef } from 'react'
import { FormProvider, useForm, useWatch } from 'react-hook-form'
import { Link, useLocation, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { AutocompleteMultiInput, Button, Checkbox, Flash, Input, Select, Surface, Tabbar, TextArea, buttonClassName } from 'paper-ui'
import { contestStatus, supportedLanguages, type Activity, type SampleContest } from '../data'
import { contestEnd, contestStart, formatDateRange, formatDateTime } from '../dates'
import { usePlayground, useScenario } from '../state'
import './contests.css'

const statusLabels = { live: 'Live now', upcoming: 'Upcoming', ended: 'Ended' }
const scopeLabels = { official: 'Official', community: 'Community', mine: 'My contests' }

function useContestDate() {
  const [scenario] = useScenario('contests')
  const [params] = useSearchParams()
  const requested = params.get('asOf') ?? ''
  return /^\d{4}-\d{2}-\d{2}$/.test(requested) ? requested : scenario === 'between' ? '2026-10-01' : '2026-09-05'
}
function registrationIsOpen(contest: SampleContest, asOf: string) {
  return contestStatus(contest, asOf) !== 'ended' && asOf <= contest.registrationDeadline
}
function contestPath(contest: SampleContest, asOf: string, suffix = '') {
  return `/contests/${contest.id}${suffix}?asOf=${asOf}`
}
function ContestStatus({ contest, asOf }: { contest: SampleContest; asOf: string }) {
  const status = contestStatus(contest, asOf)
  return <span className="status-tag" data-status={status}>{statusLabels[status]}</span>
}
function ContestPrimaryAction({ contest, asOf, prominent = false }: { contest: SampleContest; asOf: string; prominent?: boolean }) {
  const { viewer, joinedContests } = usePlayground()
  const joined = viewer !== 'guest' && joinedContests.includes(contest.id)
  const live = contestStatus(contest, asOf) === 'live'
  const registration = contestPath(contest, asOf, '/registration')
  const target = live && joined ? `/logs/new?contest=${contest.id}&asOf=${asOf}` : registrationIsOpen(contest, asOf) ? viewer === 'guest' ? `/sign-in?next=${encodeURIComponent(registration)}` : registration : contestPath(contest, asOf, '/leaderboard')
  const label = live && joined ? 'Log activity' : registrationIsOpen(contest, asOf) ? viewer === 'guest' ? 'Sign in to join' : joined ? 'Update registration' : 'Join contest' : contestStatus(contest, asOf) === 'ended' ? 'View results' : 'View standings'
  return <Link className={buttonClassName({ variant: prominent ? 'default' : 'outline' })} to={target}>{label}</Link>
}
function ContestFacts({ contest, asOf }: { contest: SampleContest; asOf: string }) {
  return <dl className="contest-facts">
    <div><dt>Round begins</dt><dd>{formatDateTime(contestStart(contest.start))}</dd></div>
    <div><dt>Round ends</dt><dd>{formatDateTime(contestEnd(contest.end))}</dd></div>
    <div><dt>Joining this round</dt><dd>{registrationIsOpen(contest, asOf) ? `Open until ${formatDateTime(contestEnd(contest.registrationDeadline))}` : 'Registration closed'}</dd></div>
  </dl>
}
function ContestFeature({ contest, asOf }: { contest: SampleContest; asOf: string }) {
  const { viewer, joinedContests } = usePlayground()
  return <Surface as="article" className="contest-feature">
    <div className="contest-feature__body"><div className="contest-feature__heading"><h2 className="paper-type-section"><Link to={contestPath(contest, asOf)}>{contest.title}</Link></h2><ContestStatus contest={contest} asOf={asOf} /></div><p>{contest.description}</p><p className="muted">{contest.languages.join(', ')} · {contest.activities.join(' and ')}</p>
      {viewer !== 'guest' && joinedContests.includes(contest.id) ? <p><strong>You’re registered.</strong></p> : null}
      <div className="flex flex-wrap gap-3"><ContestPrimaryAction contest={contest} asOf={asOf} prominent /><Link className={buttonClassName({ variant: 'outline' })} to={contestPath(contest, asOf, '/leaderboard')}>View standings</Link></div>
    </div>
    <div className="contest-feature__timeline"><ContestFacts contest={contest} asOf={asOf} /><Link className="text-link" to="/guide">Read the contest guide</Link></div>
  </Surface>
}
function ContestRow({ contest, asOf }: { contest: SampleContest; asOf: string }) {
  const { viewer, userId, joinedContests } = usePlayground()
  const date = new Date(contestStart(contest.start))
  const canManage = ['organizer', 'admin'].includes(viewer) && contest.ownerId === userId
  return <article className="contest-row">
    <div className="contest-date" aria-hidden="true"><span>{new Intl.DateTimeFormat(undefined, { month: 'short', year: 'numeric' }).format(date)}</span><strong>{new Intl.DateTimeFormat(undefined, { day: '2-digit' }).format(date)}</strong></div>
    <div className="contest-row__body"><div className="flex flex-wrap items-center gap-2"><h3 className="paper-type-component"><Link className="text-link" to={contestPath(contest, asOf)}>{contest.title}</Link></h3>{contest.unlisted ? <span className="status-tag">Unlisted</span> : null}</div><p>{formatDateRange(contestStart(contest.start), contestEnd(contest.end))}</p><p className="muted">{contest.languages.join(', ')} · {contest.activities.join(' and ')}</p></div>
    <div className="contest-row__actions"><ContestStatus contest={contest} asOf={asOf} /><ContestPrimaryAction contest={contest} asOf={asOf} /></div>
    <details className="contest-row__details"><summary>Contest details</summary><div><p>{contest.description}</p><p><strong>{registrationIsOpen(contest, asOf) ? `Registration until ${formatDateTime(contestEnd(contest.registrationDeadline))}.` : 'Registration closed.'}</strong>{viewer !== 'guest' && joinedContests.includes(contest.id) ? ' You are registered.' : ''}</p>{contest.unlisted ? <p>Anyone with the link can view this contest. It does not appear in public discovery.</p> : null}<div className="flex flex-wrap gap-x-6 gap-y-2"><Link className="text-link" to={contestPath(contest, asOf, '/leaderboard')}>{contestStatus(contest, asOf) === 'ended' ? 'View final standings' : 'View standings'}</Link><Link className="text-link" to="/guide/scoring">How contest scoring works</Link>{canManage ? <Link className="text-link" to={contestPath(contest, asOf, '/edit')}>Manage contest</Link> : null}</div></div></details>
  </article>
}

export function ContestsPage() {
  const location = useLocation()
  const navigate = useNavigate()
  const { contests, viewer, userId } = usePlayground()
  const [scenario, setScenario] = useScenario('contests')
  const asOf = useContestDate()
  const collection = location.pathname.split('/')[2]
  const scope = collection === 'community' || collection === 'mine' ? collection : 'official'
  const methods = useForm({ defaultValues: { query: '', period: 'all' } })
  const { query = '', period = 'all' } = useWatch({ control: methods.control })
  const requestedScenario = new URLSearchParams(location.search).get('scenario')
  const initializedScenario = useRef<string | null>(null)
  // A named scenario starts in its authored collection; subsequent tab navigation stays user-controlled.
  useEffect(() => {
    if (initializedScenario.current === requestedScenario) return
    initializedScenario.current = requestedScenario
    if (requestedScenario === 'empty' || requestedScenario === 'creator') navigate(`/contests/mine?scenario=${requestedScenario}`, { replace: true })
  }, [requestedScenario, navigate])
  const blocked = scope === 'mine' && viewer === 'guest'
  const isError = scenario === 'unavailable'
  const canCreate = ['organizer', 'admin'].includes(viewer)
  const collectionRows = contests.filter(contest => scope === 'mine' ? contest.ownerId === userId && scenario !== 'empty' : !contest.unlisted && contest.scope === scope)
  const rows = collectionRows.filter(contest => (period === 'all' || contestStatus(contest, asOf) === period) && `${contest.title} ${contest.languages.join(' ')}`.toLocaleLowerCase().includes(query.trim().toLocaleLowerCase()))
    .sort((left, right) => left.start.localeCompare(right.start) || left.title.localeCompare(right.title))
  const feature = scope === 'official' && !query.trim() && period === 'all' ? rows.find(contest => contestStatus(contest, asOf) === 'live') ?? rows.find(contest => contestStatus(contest, asOf) === 'upcoming') : undefined
  const descriptions = { official: 'Official rounds bring the Tadoku community together throughout the year.', community: 'Public challenges organized by Tadoku members. Check the languages and activities before joining.', mine: 'Contests you organize, including unlisted contests. Your participation is shown on each contest’s page.' }
  const clearFilters = () => { methods.reset(); methods.setFocus('query') }

  return <>
    <header className="page-header contest-page-header"><div><h1 className="paper-type-page">Contests</h1><p className="page-lead">{descriptions[scope]}</p></div>{canCreate ? <Link className={buttonClassName()} to={`/contests/new?asOf=${asOf}`}>Create contest</Link> : null}</header>
    <Tabbar label="Contest collections" links={(['official', 'community', 'mine'] as const).map(id => ({ id, label: scopeLabels[id], href: `/contests/${id}${location.search}`, current: scope === id, onSelect: () => methods.reset() }))} renderLink={({ href, ...props }) => <Link to={href} {...props} />} />
    {blocked ? <section className="empty-state"><h2 className="paper-type-section">Your contests, in one place.</h2><p>Sign in to see the contests you organize. You can still browse official rounds and public community challenges.</p><Link className={buttonClassName()} to={`/sign-in?next=${encodeURIComponent('/contests/mine')}`}>Sign in</Link></section> : isError ? <section className="empty-state" role="alert"><h2 className="paper-type-section">Contests could not be loaded.</h2><p>Try loading this collection again.</p><Button onClick={() => setScenario('member')}>Try again</Button></section> : <>
      <div className="contest-toolbar"><FormProvider {...methods}><form className="contest-filters" onSubmit={event => event.preventDefault()}><Input name="query" label="Find a contest" type="search" placeholder="Search title or language" /><Select name="period" label="Dates" options={[{ value: 'all', label: 'All dates' }, { value: 'live', label: 'Live now' }, { value: 'upcoming', label: 'Upcoming' }, { value: 'ended', label: 'Past contests' }]} />{query || period !== 'all' ? <Button variant="ghost" onClick={clearFilters}>Clear filters</Button> : null}</form></FormProvider>
      <p className="contest-count muted" role="status">{rows.length} {rows.length === 1 ? 'contest' : 'contests'}{query.trim() || period !== 'all' ? ' match your filters' : ''}</p></div>
      {rows.length === 0 ? <section className="empty-state"><h2 className="paper-type-section">{scope === 'mine' && collectionRows.length === 0 && !query && period === 'all' ? 'You haven’t organized a contest yet.' : 'No contests match these filters.'}</h2><p>{scope === 'mine' && collectionRows.length === 0 && !query && period === 'all' ? 'Find a round to join in Official or Community. Accounts with organizer permission can also create a contest.' : 'Try another title or language, or include all dates.'}</p>{scope === 'mine' && collectionRows.length === 0 && !query && period === 'all' ? <Link className={buttonClassName({ variant: 'outline' })} to="/contests/official">Explore official contests</Link> : <Button variant="outline" onClick={clearFilters}>Clear filters</Button>}</section> : <>
        {feature ? <ContestFeature contest={feature} asOf={asOf} /> : null}
        {(['live', 'upcoming', 'ended'] as const).map(status => {
          const group = rows.filter(contest => contest.id !== feature?.id && contestStatus(contest, asOf) === status)
          if (status === 'ended') group.reverse()
          return group.length > 0 ? <section className="contest-group" key={status} aria-label={{ live: 'Happening now', upcoming: 'Coming up', ended: 'Past contests' }[status]}><h2 className="paper-type-section">{{ live: 'Happening now', upcoming: 'Coming up', ended: 'Past contests' }[status]}</h2>{group.map(contest => <ContestRow key={contest.id} contest={contest} asOf={asOf} />)}</section> : null
        })}
      </>}
    </>}
    <p className="contest-time-note muted">Contest times are shown in your local time zone.</p>
  </>
}

function ContestNotFound() {
  return <><header className="page-header"><h1 className="paper-type-page">Contest not found</h1></header><section className="empty-state"><p>This link does not match a contest in the collection.</p><Link className={buttonClassName()} to="/contests/official">Browse contests</Link></section></>
}

export function ContestDetailPage() {
  const { contestId } = useParams()
  const { contests, viewer, userId, joinedContests, registrations } = usePlayground()
  const asOf = useContestDate()
  const [params] = useSearchParams()
  const contest = contests.find(item => item.id === contestId)
  if (!contest) return <ContestNotFound />
  const joined = viewer !== 'guest' && joinedContests.includes(contest.id)
  const canManage = ['organizer', 'admin'].includes(viewer) && contest.ownerId === userId
  return <>
    <Link className="text-link page-back-link" to={contest.unlisted && contest.ownerId === userId && viewer !== 'guest' ? '/contests/mine' : `/contests/${contest.scope}`}>Back to {contest.unlisted && contest.ownerId === userId && viewer !== 'guest' ? 'my contests' : `${contest.scope} contests`}</Link>
    <header className="page-header"><div className="contest-detail-heading"><h1 className="paper-type-page">{contest.title}</h1><ContestStatus contest={contest} asOf={asOf} />{contest.unlisted ? <span className="status-tag">Unlisted</span> : null}</div><p className="page-lead">{contest.description}</p></header>
    {params.has('registered') && joined ? <Flash variant="success" title="Registration saved">You’re registered for {contest.title}. Choose a registered language when you log activity.</Flash> : null}
    {params.has('saved') ? <Flash variant="success" title="Contest saved">Your contest details have been updated.</Flash> : null}
    <div className="page-columns contest-detail-layout"><section className="contest-detail-body">
      <h2 className="paper-type-section section-heading">Taking part</h2><p>Join this contest, then submit your reading or listening as you go. Choose a registered language and one of the activities below when you log.</p>
      <dl className="contest-eligibility"><div><dt>Languages</dt><dd>{contest.languages.join(', ')}</dd></div><div><dt>Activities</dt><dd>{contest.activities.join(' and ')}</dd></div></dl>
      {contest.unlisted ? <Flash variant="information" title="Accessible with the link">Anyone with this link can view and join while registration is open. This contest is excluded from public discovery; it is not a private space.</Flash> : null}
      {joined ? <p><strong>You are registered.</strong>{registrations[contest.id]?.length ? ` Your languages: ${registrations[contest.id].join(', ')}.` : ''}</p> : null}
      <div className="flex flex-wrap gap-3 my-6"><ContestPrimaryAction contest={contest} asOf={asOf} prominent /><Link className={buttonClassName({ variant: 'outline' })} to={contestPath(contest, asOf, '/leaderboard')}>View standings</Link>{canManage ? <Link className={buttonClassName({ variant: 'ghost' })} to={contestPath(contest, asOf, '/edit')}>Manage contest</Link> : null}</div>
      {joined && registrationIsOpen(contest, asOf) && contestStatus(contest, asOf) === 'live' ? <p><Link className="text-link" to={contestPath(contest, asOf, '/registration')}>Update your registered languages</Link></p> : null}
      <p className="muted">Personal tracking stays available outside contest rounds. A contest’s eligibility and saved score can differ from your personal total.</p><Link className="text-link" to="/guide/scoring">Understand scores and eligibility</Link>
    </section><aside className="contest-detail-timeline"><h2 className="paper-type-component">Dates &amp; registration</h2><ContestFacts contest={contest} asOf={asOf} /><p className="muted">Times are shown in your local time zone. Registration can close before a live round ends; registered participants can continue logging until the end.</p></aside></div>
  </>
}

export function ContestRegistrationPage() {
  const { contestId } = useParams()
  const { contests, viewer, registrations, joinContest } = usePlayground()
  const navigate = useNavigate()
  const asOf = useContestDate()
  const contest = contests.find(item => item.id === contestId)
  const methods = useForm<{ languages: string[] }>({ defaultValues: { languages: registrations[contestId ?? ''] ?? [] } })
  if (!contest) return <ContestNotFound />
  const allowed = contest.languages.includes('All languages') ? supportedLanguages : contest.languages
  const submit = methods.handleSubmit(values => { joinContest(contest.id, values.languages); navigate(`${contestPath(contest, asOf)}&registered=1`) })
  return <>
    <Link className="text-link page-back-link" to={contestPath(contest, asOf)}>Back to {contest.title}</Link>
    <header className="page-header"><h1 className="paper-type-page">Join {contest.title}</h1><p className="page-lead">Choose the languages you’ll read or listen in. You can register up to three.</p></header>
    <div className="page-columns contest-registration-layout"><section>
      {!registrationIsOpen(contest, asOf) ? <section className="empty-state"><h2 className="paper-type-section">Registration is closed.</h2><p>{contestStatus(contest, asOf) === 'live' ? 'This round is still running. Participants who already registered can continue logging.' : 'Browse another round to join, or explore the results from this one.'}</p><ContestPrimaryAction contest={contest} asOf={asOf} /></section> : viewer === 'guest' ? <section className="empty-state"><h2 className="paper-type-section">Sign in to choose your languages.</h2><p>You’ll return to this registration after signing in.</p><Link className={buttonClassName()} to={`/sign-in?next=${encodeURIComponent(contestPath(contest, asOf, '/registration'))}`}>Sign in to join</Link></section> : <FormProvider {...methods}><form onSubmit={submit} className="app-form" noValidate><div className="app-form__fields"><AutocompleteMultiInput name="languages" label="Your languages" hint="Choose one to three languages. Remove a language to make room for another." options={allowed} format={value => value} getId={value => value} required maxSelections={3} rules={{ validate: value => Array.isArray(value) && value.length >= 1 && value.length <= 3 && value.every(language => allowed.includes(language)) || 'Choose one to three of the available languages.' }} /><p className="muted">Allowed activities: {contest.activities.join(' and ')}. After joining, choose which eligible logs to submit; joining does not submit past activity automatically.</p></div><div className="app-form__actions"><Button type="submit">Save registration</Button><Link className={buttonClassName({ variant: 'outline' })} to={contestPath(contest, asOf)}>Cancel</Link></div></form></FormProvider>}
    </section><aside className="contest-detail-timeline"><h2 className="paper-type-component">Your round</h2><ContestFacts contest={contest} asOf={asOf} /></aside></div>
  </>
}

interface ContestFields { title: string; description: string; start: string; end: string; registrationDeadline: string; allLanguages: boolean; languages: string[]; reading: boolean; listening: boolean; unlisted: boolean }
export function ContestEditorPage() {
  const { contestId } = useParams()
  const { contests, viewer, userId, saveContest } = usePlayground()
  const navigate = useNavigate()
  const asOf = useContestDate()
  const contest = contests.find(item => item.id === contestId)
  const methods = useForm<ContestFields>({ defaultValues: { title: contest?.title ?? '', description: contest?.description ?? '', start: contest?.start ?? '2026-10-01', end: contest?.end ?? '2026-10-31', registrationDeadline: contest?.registrationDeadline ?? '2026-10-05', allLanguages: !contest || contest.languages.includes('All languages'), languages: contest?.languages.filter(language => language !== 'All languages') ?? [], reading: contest?.activities.includes('Reading') ?? true, listening: contest?.activities.includes('Listening') ?? true, unlisted: contest?.unlisted ?? false } })
  const allLanguages = useWatch({ control: methods.control, name: 'allLanguages' })
  const [start, end, registrationDeadline] = useWatch({ control: methods.control, name: ['start', 'end', 'registrationDeadline'] })
  const canManage = ['organizer', 'admin'].includes(viewer) && (!contestId || contest?.ownerId === userId)
  if (contestId && !contest) return <ContestNotFound />
  if (!canManage) return <><header className="page-header"><h1 className="paper-type-page">Organizer access required</h1></header><section className="empty-state"><p>Creating contests requires organizer permission. You can manage only the contests you organize.</p><Link className={buttonClassName({ variant: 'outline' })} to="/contests/official">Browse contests</Link></section></>
  const save = methods.handleSubmit(values => {
    const activities: Activity[] = []
    if (values.reading) activities.push('Reading')
    if (values.listening) activities.push('Listening')
    const next: SampleContest = { id: contest?.id ?? `contest-${crypto.randomUUID()}`, title: values.title.trim(), description: values.description.trim(), start: values.start, end: values.end, registrationDeadline: values.registrationDeadline, languages: values.allLanguages ? ['All languages'] : values.languages, activities, unlisted: values.unlisted, scope: contest?.scope ?? 'community', ownerId: userId, moderatorIds: contest ? contest.moderatorIds : [userId] }
    saveContest(next)
    navigate(`${contestPath(next, asOf)}&saved=1`)
  })
  return <>
    <Link className="text-link page-back-link" to={contest ? contestPath(contest, asOf) : '/contests/managed'}>{contest ? `Back to ${contest.title}` : 'Back to contests'}</Link>
    <header className="page-header"><h1 className="paper-type-page">{contest ? 'Manage contest' : 'Create a community contest'}</h1><p className="page-lead">Set a shared window for immersion, with clear dates and eligibility.</p></header>
    <FormProvider {...methods}><form onSubmit={save} noValidate className="app-form contest-editor-form"><div className="app-form__fields">
      <Input name="title" label="Contest title" required maxLength={100} rules={{ validate: value => Boolean(value?.trim()) || 'Enter a contest title.' }} />
      <TextArea name="description" label="What is this contest about?" rows={3} required rules={{ validate: value => Boolean(value?.trim()) || 'Describe what participants will do.' }} />
      <div className="app-form__row"><Input name="start" label="Begins (UTC)" type="date" required hint={start ? `Your time: ${formatDateTime(contestStart(start))}` : undefined} /><Input name="end" label="Ends (UTC)" type="date" required hint={end ? `Your time: ${formatDateTime(contestEnd(end))}` : undefined} rules={{ validate: value => value >= methods.getValues('start') || 'The end date must be on or after the start date.' }} /></div>
      <Input name="registrationDeadline" label="Registration closes (UTC)" type="date" required hint={registrationDeadline ? `Your time: ${formatDateTime(contestEnd(registrationDeadline))}. Must close on or before the round ends.` : 'Must close on or before the round ends.'} rules={{ validate: value => value <= methods.getValues('end') || 'Registration must close on or before the round ends.' }} />
      <fieldset className="contest-form-group"><legend className="paper-type-component">Languages</legend><Checkbox name="allLanguages" label="Allow all languages" />{!allLanguages ? <AutocompleteMultiInput name="languages" label="Allowed languages" options={supportedLanguages} format={value => value} getId={value => value} required rules={{ validate: value => methods.getValues('allLanguages') || Array.isArray(value) && value.length > 0 || 'Choose at least one allowed language.' }} /> : <p className="muted">Participants choose up to three languages when registering.</p>}</fieldset>
      <fieldset className="contest-form-group"><legend className="paper-type-component">Activities</legend><Checkbox name="reading" label="Reading" rules={{ validate: () => methods.getValues('reading') || methods.getValues('listening') || 'Choose at least one activity.' }} /><Checkbox name="listening" label="Listening" /></fieldset>
      <fieldset className="contest-form-group"><legend className="paper-type-component">Discovery</legend><Checkbox name="unlisted" label="Unlisted — accessible with the link" hint="Unlisted contests appear in My contests and can be opened by anyone with their link. They are excluded from public discovery." /></fieldset>
    </div><div className="app-form__actions"><Button type="submit">{contest ? 'Save changes' : 'Create contest'}</Button><Link className={buttonClassName({ variant: 'outline' })} to={contest ? contestPath(contest, asOf) : '/contests/mine'}>Cancel</Link></div></form></FormProvider>
  </>
}
