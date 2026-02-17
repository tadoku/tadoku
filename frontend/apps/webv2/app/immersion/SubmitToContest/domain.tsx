import { ContestRegistrationView, Log } from '@app/immersion/api'
import { DateTime, Interval } from 'luxon'

export interface ContestOption {
  registration: ContestRegistrationView
  eligible: boolean
  reason?: string
}

export function classifyRegistrations(
  log: Log,
  registrations: ContestRegistrationView[],
): ContestOption[] {
  return registrations
    .filter(r => r.contest)
    .map(registration => {
      const contest = registration.contest!

      const activityAllowed = contest.allowed_activities
        .map(a => a.id)
        .includes(log.activity.id)
      if (!activityAllowed) {
        return {
          registration,
          eligible: false,
          reason: 'Activity not allowed in this contest',
        }
      }

      const languageRegistered = registration.languages
        .map(l => l.code)
        .includes(log.language.code)
      if (!languageRegistered) {
        return {
          registration,
          eligible: false,
          reason: 'Language not registered for this contest',
        }
      }

      const contestOngoing = Interval.fromDateTimes(
        DateTime.fromISO(contest.contest_start),
        DateTime.fromISO(contest.contest_end).endOf('day'),
      ).contains(DateTime.now())
      if (!contestOngoing) {
        return {
          registration,
          eligible: false,
          reason: 'Contest has ended',
        }
      }

      return { registration, eligible: true }
    })
}
