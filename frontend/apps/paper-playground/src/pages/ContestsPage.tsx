import * as stylex from '@stylexjs/stylex'
import { useStore } from '@tanstack/react-form'
import { Link, useLocation } from 'react-router-dom'
import { Button, buttonStyles, EmptyState, linkStyles, PageHeader, SectionHeading, StatusTag, Surface, Tabbar, typography, usePaperForm } from 'paper-ui'
import { contestStatus, type SampleContest } from '../data'
import { contestEnd, contestStart, formatDateRange, formatDateTime } from '../dates'
import { usePlayground, useScenario } from '../state'
import { ContestFacts, ContestPrimaryAction, ContestStatus } from './contestParts'
import { contestPath, registrationIsOpen, useContestDate } from './contestRound'
import { contestsStyles } from './contests.styles'
import { utilities } from '../utilities.styles'
import { appStyles } from '../app.styles'
import { descendantStyles } from '../descendants.styles'

const scopeLabels = { official: 'Official', community: 'Community', participating: 'Participating', managed: 'Managed by me' }
type ContestCollection = keyof typeof scopeLabels

function ContestFeature({ contest, asOf }: { contest: SampleContest; asOf: string }) {
  const { viewer, joinedContests } = usePlayground()
  return <Surface as="article" xstyle={[contestsStyles.contestFeature]}>
    <div {...stylex.props(contestsStyles.contestFeature__body)}><SectionHeading level={2} xstyle={[contestsStyles.contestFeature__heading]} title={<Link to={contestPath(contest, asOf)}>{contest.title}</Link>} status={<ContestStatus contest={contest} asOf={asOf} />} /><p {...stylex.props(descendantStyles.contestFeatureU_uBodyU_uP, descendantStyles.contestFeatureU_uBodyU_uAny)}>{contest.description}</p><p {...stylex.props(appStyles.muted, descendantStyles.contestFeatureU_uBodyU_uP, descendantStyles.contestFeatureU_uBodyU_uAny)}>{contest.languages.join(', ')} · {contest.activities.join(' and ')}</p>
      {viewer !== 'guest' && joinedContests.includes(contest.id) ? <p {...stylex.props(descendantStyles.contestFeatureU_uBodyU_uP, descendantStyles.contestFeatureU_uBodyU_uAny)}><strong>You’re registered.</strong></p> : null}
      <div {...stylex.props(utilities.tw_flex, utilities.tw_flex_wrap, utilities.tw_gap_3, descendantStyles.contestFeatureU_uBodyU_uAny)}><ContestPrimaryAction contest={contest} asOf={asOf} prominent /><Link {...stylex.props(...buttonStyles({ variant: 'outline' }))} to={contestPath(contest, asOf, '/leaderboard')}>View standings</Link></div>
    </div>
    <div {...stylex.props(contestsStyles.contestFeature__timeline)}><ContestFacts contest={contest} asOf={asOf} /><Link {...stylex.props(linkStyles.base)} to="/guide">Read the contest guide</Link></div>
  </Surface>
}
function ContestRow({ contest, asOf, managed = false }: { contest: SampleContest; asOf: string; managed?: boolean }) {
  const { viewer, userId, joinedContests } = usePlayground()
  const date = new Date(contestStart(contest.start))
  const canManage = ['organizer', 'admin'].includes(viewer) && contest.ownerId === userId
  return <article {...stylex.props(contestsStyles.contestRow)}>
    <div {...stylex.props(contestsStyles.contestDate)} aria-hidden="true"><span {...stylex.props(descendantStyles.contestDateU_uSpan)}>{new Intl.DateTimeFormat(undefined, { month: 'short', year: 'numeric' }).format(date)}</span><strong {...stylex.props(descendantStyles.contestDateU_uStrong)}>{new Intl.DateTimeFormat(undefined, { day: '2-digit' }).format(date)}</strong></div>
    <div {...stylex.props(contestsStyles.contestRow__body)}><div {...stylex.props(utilities.tw_flex, utilities.tw_flex_wrap, utilities.tw_items_center, utilities.tw_gap_2)}><h3 {...stylex.props(typography.component, descendantStyles.contestRowU_uBodyU_uH3)}><Link {...stylex.props(linkStyles.base)} to={contestPath(contest, asOf)}>{contest.title}</Link></h3>{contest.unlisted ? <StatusTag>Unlisted</StatusTag> : null}</div><p {...stylex.props(descendantStyles.contestRowU_uBodyU_uP)}>{formatDateRange(contestStart(contest.start), contestEnd(contest.end))}</p><p {...stylex.props(appStyles.muted, descendantStyles.contestRowU_uBodyU_uP)}>{contest.languages.join(', ')} · {contest.activities.join(' and ')}</p></div>
    <div {...stylex.props(contestsStyles.contestRow__actions)}><ContestStatus contest={contest} asOf={asOf} />{managed ? <Link {...stylex.props(...buttonStyles({ variant: 'outline' }))} to={contestPath(contest, asOf, canManage ? '/edit' : '')}>{canManage ? 'Manage contest' : 'View contest'}</Link> : <ContestPrimaryAction contest={contest} asOf={asOf} />}</div>
    <details {...stylex.props(contestsStyles.contestRow__details)}><summary {...stylex.props(descendantStyles.contestRowU_uDetailsU_uSummary)}>Contest details</summary><div {...stylex.props(descendantStyles.contestRowU_uDetailsU_uDiv)}><p {...stylex.props(descendantStyles.contestRowU_uDetailsU_uP)}>{contest.description}</p><p {...stylex.props(descendantStyles.contestRowU_uDetailsU_uP)}><strong>{registrationIsOpen(contest, asOf) ? `Registration until ${formatDateTime(contestEnd(contest.registrationDeadline))}.` : 'Registration closed.'}</strong>{viewer !== 'guest' && joinedContests.includes(contest.id) ? ' You are registered.' : ''}</p>{contest.unlisted ? <p {...stylex.props(descendantStyles.contestRowU_uDetailsU_uP)}>Anyone with the link can view this contest. It does not appear in public discovery.</p> : null}<div {...stylex.props(utilities.tw_flex, utilities.tw_flex_wrap, utilities.tw_gap_x_6, utilities.tw_gap_y_2)}><Link {...stylex.props(linkStyles.base)} to={contestPath(contest, asOf, '/leaderboard')}>{contestStatus(contest, asOf) === 'ended' ? 'View final standings' : 'View standings'}</Link><Link {...stylex.props(linkStyles.base)} to="/guide/scoring">How contest scoring works</Link>{canManage && !managed ? <Link {...stylex.props(linkStyles.base)} to={contestPath(contest, asOf, '/edit')}>Manage contest</Link> : null}</div></div></details>
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
    <PageHeader title="Contests" lead={descriptions[scope]} actions={canCreate ? <Link {...stylex.props(...buttonStyles())} to={`/contests/new?asOf=${asOf}`}>Create contest</Link> : null} />
    <div {...stylex.props(contestsStyles.contestCollections)}><Tabbar label="Contest collections" links={collections.map(id => ({ id, label: scopeLabels[id], href: `/contests/${id}${location.search}`, current: scope === id, onSelect: () => methods.reset() }))} renderLink={({ href, ...props }) => <Link to={href} {...props} />} /></div>
    {blocked ? <EmptyState headingLevel={2} title={scope === 'participating' ? 'Sign in to see the contests you’ve joined.' : 'Sign in to see the contests you manage.'} description="You can still browse official rounds and public community challenges." actions={<Link {...stylex.props(...buttonStyles())} to={`/sign-in?next=${encodeURIComponent(`/contests/${scope}`)}`}>Sign in</Link>} /> : isError ? <EmptyState role="alert" headingLevel={2} title="Contests could not be loaded." description="Try loading this collection again." actions={<Button onClick={() => setScenario('member')}>Try again</Button>} /> : <>
      <div {...stylex.props(contestsStyles.contestToolbar)}><methods.AppForm><form {...stylex.props(contestsStyles.contestFilters)} onSubmit={event => event.preventDefault()}><methods.AppField name="query">{field => <field.Input label="Find a contest" type="search" placeholder="Search title or language" />}</methods.AppField><methods.AppField name="period">{field => <field.Select label="Dates" options={[{ value: 'all', label: 'All dates' }, { value: 'live', label: 'Live now' }, { value: 'upcoming', label: 'Upcoming' }, { value: 'ended', label: 'Past contests' }]} />}</methods.AppField>{query || period !== 'all' ? <Button variant="ghost" onClick={clearFilters}>Clear filters</Button> : null}</form></methods.AppForm>
      <p {...stylex.props(contestsStyles.contestCount, appStyles.muted)} role="status">{rows.length} {rows.length === 1 ? 'contest' : 'contests'}{query.trim() || period !== 'all' ? ' match your filters' : ''}</p></div>
      {rows.length === 0 ? <EmptyState headingLevel={2} title={emptyCollection ? scope === 'participating' ? 'You haven’t joined a contest yet.' : 'You don’t manage any contests yet.' : 'No contests match these filters.'} description={emptyCollection ? scope === 'participating' ? 'Join an official round or community challenge to see it here.' : 'Contests appear here when you organize them or are assigned as a moderator.' : 'Try another title or language, or include all dates.'} actions={emptyCollection ? <Link {...stylex.props(...buttonStyles({ variant: 'outline' }))} to="/contests/official">Explore official contests</Link> : <Button variant="outline" onClick={clearFilters}>Clear filters</Button>} /> : <>
        {feature ? <ContestFeature contest={feature} asOf={asOf} /> : null}
        {(['live', 'upcoming', 'ended'] as const).map(status => {
          const group = rows.filter(contest => contest.id !== feature?.id && contestStatus(contest, asOf) === status)
          if (status === 'ended') group.reverse()
          return group.length > 0 ? <section {...stylex.props(contestsStyles.contestGroup)} key={status} aria-label={{ live: 'Happening now', upcoming: 'Coming up', ended: 'Past contests' }[status]}><SectionHeading level={2} title={{ live: 'Happening now', upcoming: 'Coming up', ended: 'Past contests' }[status]} />{group.map(contest => <ContestRow key={contest.id} contest={contest} asOf={asOf} managed={scope === 'managed'} />)}</section> : null
        })}
      </>}
    </>}
    <p {...stylex.props(contestsStyles.contestTimeNote, appStyles.muted)}>Contest times are shown in your local time zone.</p>
  </>
}
