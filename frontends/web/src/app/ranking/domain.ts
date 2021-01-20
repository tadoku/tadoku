import { languageByCode } from './database'
import { Contest } from '@app/contest/interfaces'
import { RankingRegistration } from './interfaces'
import { User } from '@app/session/interfaces'

export const validateAmount = (amount: string): boolean =>
  Number(amount) !== NaN && Number(amount) > 0

export const validateLanguageCode = (code: string): boolean =>
  code != '' && languageByCode[code] !== undefined

export const isContestRunning = (contest: Contest): boolean =>
  contest.end > new Date()

export const isContestActive = (contest: Contest): boolean =>
  contest.open && isContestRunning(contest)

export const isContestEditable = (contest: Contest): boolean =>
  contest.open || isContestRunning(contest)

export const isRegisteredForContest = (
  registration: RankingRegistration | undefined,
  contest: Contest,
): boolean => registration?.contestId === contest.id

export const canJoinContest = (
  user: User | undefined,
  registration: RankingRegistration | undefined,
  contest: Contest,
) =>
  user &&
  contest &&
  isContestActive(contest) &&
  !isRegisteredForContest(registration, contest)

export const isRegistrationClosedFor = (
  user: User | undefined,
  registration: RankingRegistration | undefined,
  contest: Contest,
) =>
  user &&
  contest &&
  isContestRunning(contest) &&
  !isRegisteredForContest(registration, contest)
