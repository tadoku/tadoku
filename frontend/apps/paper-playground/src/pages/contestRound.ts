import { useSearchParams } from 'react-router-dom'
import { contestStatus, type SampleContest } from '../data'
import { useScenario } from '../state'

export function useContestDate() {
  const [scenario] = useScenario('contests')
  const [params] = useSearchParams()
  const requested = params.get('asOf') ?? ''
  return /^\d{4}-\d{2}-\d{2}$/.test(requested) ? requested : scenario === 'between' ? '2026-10-01' : '2026-09-05'
}
export function registrationIsOpen(contest: SampleContest, asOf: string) {
  return contestStatus(contest, asOf) !== 'ended' && asOf <= contest.registrationDeadline
}
export function contestPath(contest: SampleContest, asOf: string, suffix = '') {
  return `/contests/${contest.id}${suffix}?asOf=${asOf}`
}
