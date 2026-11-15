const dayMs = 86_400_000

const daysBefore = (date: string, count: number) => new Date(Date.parse(`${date}T00:00:00Z`) - count * dayMs).toISOString().slice(0, 10)
const daysBetween = (from: string, to: string) => Math.round((Date.parse(to) - Date.parse(from)) / dayMs)

/** How many weeks the signed-in home's activity map shows, which is also how far back the page reads. */
export const activityWeeks = 13

/** Whether the activity map reaches back into last year from `today`, so the page reads last year's activity as well. */
export const readsLastYear = (today: string) => daysBetween(`${today.slice(0, 4)}-01-01`, today) < activityWeeks * 7

export type GettingGoingSteps = {
  logged: boolean
  loggedToday: boolean
  joinedContest: boolean
  twoDaysInARow: boolean
}

/** What the signed-in home shows beside the contest board. */
export type HomeActivity =
  | { kind: 'getting-going'; activeDays: number; steps: GettingGoingSteps }
  | { kind: 'active'; total: number; today: number; streak: number; activeDays: number }

export const activeDaysGoal = 5
const windowDays = 30

const activeDates = (scores: ReadonlyMap<string, number>) =>
  [...scores].filter(([, score]) => score > 0).map(([date]) => date).sort()

/** A contest the learner registered for, by its UTC dates. */
export type ContestPeriod = { contest_start: string; contest_end: string }

/**
 * Whether `scores` show the learner has cleared the getting-going checklist: 5 active days within 30, or two active days
 * in a row inside one of `contests`, which is every step.
 */
export function clearedGettingGoing(scores: ReadonlyMap<string, number>, contests: readonly ContestPeriod[]) {
  const active = activeDates(scores)
  const inContest = (date: string) => contests.some(contest => contest.contest_start <= date && date <= contest.contest_end)
  return active.some((date, index) => index + activeDaysGoal - 1 < active.length
    && daysBetween(date, active[index + activeDaysGoal - 1]) < windowDays)
    || active.some((date, index) => index > 0 && daysBetween(active[index - 1], date) === 1
      && inContest(active[index - 1]) && inContest(date))
}

/**
 * The learner's past 30 UTC days, or the getting-going checklist until they've cleared it in `scores` and `contests` or,
 * as `clearedBefore` says, anywhere in their history. They keep the map from then on even after a quiet month.
 */
export function homeActivity(
  scores: ReadonlyMap<string, number>,
  today: string,
  contests: readonly ContestPeriod[],
  clearedBefore: boolean,
): HomeActivity {
  const active = activeDates(scores)
  const todayScore = scores.get(today) ?? 0
  const recent = active.filter(date => date > daysBefore(today, windowDays) && date <= today)

  let streak = 0
  for (let back = todayScore > 0 ? 0 : 1; (scores.get(daysBefore(today, back)) ?? 0) > 0; back++) streak++

  if (!clearedBefore && !clearedGettingGoing(scores, contests)) {
    return {
      kind: 'getting-going',
      activeDays: recent.length,
      steps: { logged: active.length > 0, loggedToday: todayScore > 0, joinedContest: contests.length > 0, twoDaysInARow: streak >= 2 },
    }
  }

  return {
    kind: 'active',
    total: recent.reduce((sum, date) => sum + scores.get(date)!, 0),
    today: todayScore,
    streak,
    activeDays: recent.length,
  }
}
