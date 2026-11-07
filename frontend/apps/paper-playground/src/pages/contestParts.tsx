import * as stylex from '@stylexjs/stylex'
import { Link } from 'react-router-dom'
import { buttonStyles, DescriptionList, EmptyState, StatusTag } from 'paper-ui'
import { contestStatus, type SampleContest } from '../data'
import { contestEnd, contestStart, formatDateTime } from '../dates'
import { usePlayground } from '../state'
import { contestPath, registrationIsOpen } from './contestRound'
import { contestsStyles } from './contests.styles'

const statusLabels = { live: 'Live now', upcoming: 'Upcoming', ended: 'Ended' }
const statusTones = { live: 'success', upcoming: 'information', ended: 'neutral' } as const
export function ContestStatus({ contest, asOf }: { contest: SampleContest; asOf: string }) {
  const status = contestStatus(contest, asOf)
  return <StatusTag tone={statusTones[status]}>{statusLabels[status]}</StatusTag>
}
export function ContestPrimaryAction({ contest, asOf, prominent = false }: { contest: SampleContest; asOf: string; prominent?: boolean }) {
  const { viewer, joinedContests } = usePlayground()
  const joined = viewer !== 'guest' && joinedContests.includes(contest.id)
  const live = contestStatus(contest, asOf) === 'live'
  const registration = contestPath(contest, asOf, '/registration')
  const target = live && joined ? `/logs/new?contest=${contest.id}&asOf=${asOf}` : registrationIsOpen(contest, asOf) ? viewer === 'guest' ? `/sign-in?next=${encodeURIComponent(registration)}` : registration : contestPath(contest, asOf, '/leaderboard')
  const label = live && joined ? 'Log activity' : registrationIsOpen(contest, asOf) ? viewer === 'guest' ? 'Sign in to join' : joined ? 'Update registration' : 'Join contest' : contestStatus(contest, asOf) === 'ended' ? 'View results' : 'View standings'
  return <Link {...stylex.props(...buttonStyles({ variant: prominent ? 'default' : 'outline' }))} to={target}>{label}</Link>
}
export function ContestFacts({ contest, asOf }: { contest: SampleContest; asOf: string }) {
  return <DescriptionList xstyle={[contestsStyles.contestFacts]} items={[
    { label: 'Round begins', value: formatDateTime(contestStart(contest.start)) },
    { label: 'Round ends', value: formatDateTime(contestEnd(contest.end)) },
    { label: 'Joining this round', value: registrationIsOpen(contest, asOf) ? `Open until ${formatDateTime(contestEnd(contest.registrationDeadline))}` : 'Registration closed' },
  ]} />
}
export function ContestNotFound() {
  return <EmptyState headingLevel={1} title="Contest not found" description="This link does not match a contest in the collection." actions={<Link {...stylex.props(...buttonStyles())} to="/contests/official">Browse contests</Link>} />
}

