import { Link, useParams, useSearchParams } from 'react-router-dom'
import { DescriptionList, Flash, PageHeader, SectionHeading, StatusTag, buttonClassName } from 'paper-ui'
import { contestStatus } from '../data'
import { renderLink } from '../routerLink'
import { usePlayground } from '../state'
import { ContestFacts, ContestNotFound, ContestPrimaryAction, ContestStatus } from './contestParts'
import { contestPath, registrationIsOpen, useContestDate } from './contestRound'
import './contests.css'

export function ContestDetailPage() {
  const { contestId } = useParams()
  const { contests, viewer, userId, joinedContests, registrations } = usePlayground()
  const asOf = useContestDate()
  const [params] = useSearchParams()
  const contest = contests.find(item => item.id === contestId)
  if (!contest) return <ContestNotFound />
  const joined = viewer !== 'guest' && joinedContests.includes(contest.id)
  const canManage = ['organizer', 'admin'].includes(viewer) && contest.ownerId === userId
  const managedBack = contest.unlisted && contest.ownerId === userId && viewer !== 'guest'
  return <>
    <PageHeader
      backLink={managedBack ? { href: '/contests/managed', label: 'Back to managed contests' } : { href: `/contests/${contest.scope}`, label: `Back to ${contest.scope} contests` }}
      renderLink={renderLink}
      title={contest.title}
      status={<><ContestStatus contest={contest} asOf={asOf} />{contest.unlisted ? <StatusTag>Unlisted</StatusTag> : null}</>}
      lead={contest.description}
    />
    {params.has('registered') && joined ? <Flash variant="success" title="Registration saved">You’re registered for {contest.title}. Choose a registered language when you log activity.</Flash> : null}
    {params.has('saved') ? <Flash variant="success" title="Contest saved">Your contest details have been updated.</Flash> : null}
    <div className="page-columns contest-detail-layout"><section className="contest-detail-body">
      <SectionHeading level={2} title="Taking part" /><p>Join this contest, then submit your reading or listening as you go. Choose a registered language and one of the activities below when you log.</p>
      <DescriptionList className="contest-eligibility" items={[{ label: 'Languages', value: contest.languages.join(', ') }, { label: 'Activities', value: contest.activities.join(' and ') }]} />
      {contest.unlisted ? <Flash variant="information" title="Accessible with the link">Anyone with this link can view and join while registration is open. This contest is excluded from public discovery; it is not a private space.</Flash> : null}
      {joined ? <p><strong>You are registered.</strong>{registrations[contest.id]?.length ? ` Your languages: ${registrations[contest.id].join(', ')}.` : ''}</p> : null}
      <div className="flex flex-wrap gap-3 my-6"><ContestPrimaryAction contest={contest} asOf={asOf} prominent /><Link className={buttonClassName({ variant: 'outline' })} to={contestPath(contest, asOf, '/leaderboard')}>View standings</Link>{canManage ? <Link className={buttonClassName({ variant: 'ghost' })} to={contestPath(contest, asOf, '/edit')}>Manage contest</Link> : null}</div>
      {joined && registrationIsOpen(contest, asOf) && contestStatus(contest, asOf) === 'live' ? <p><Link className="text-link" to={contestPath(contest, asOf, '/registration')}>Update your registered languages</Link></p> : null}
      <p className="muted">Personal tracking stays available outside contest rounds. A contest’s eligibility and saved score can differ from your personal total.</p><Link className="text-link" to="/guide/scoring">Understand scores and eligibility</Link>
    </section><aside className="contest-detail-timeline"><h2 className="paper-type-component">Dates &amp; registration</h2><ContestFacts contest={contest} asOf={asOf} /><p className="muted">Times are shown in your local time zone. Registration can close before a live round ends; registered participants can continue logging until the end.</p></aside></div>
  </>
}

