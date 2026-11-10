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

/**
 * The learner's past 30 UTC days, or the getting-going checklist until they've cleared it. A learner has cleared it once
 * `scores` shows 5 active days within 30, or two active days in a row while they're in a contest, which is every step,
 * and keeps the map from then on even after a quiet month.
 */
export function homeActivity(scores: ReadonlyMap<string, number>, today: string, joinedContest: boolean): HomeActivity {
  const active = [...scores].filter(([, score]) => score > 0).map(([date]) => date).sort()
  const todayScore = scores.get(today) ?? 0
  const recent = active.filter(date => date > daysBefore(today, windowDays) && date <= today)

  let streak = 0
  for (let back = todayScore > 0 ? 0 : 1; (scores.get(daysBefore(today, back)) ?? 0) > 0; back++) streak++

  const cleared = active.some((date, index) => index + activeDaysGoal - 1 < active.length
    && daysBetween(date, active[index + activeDaysGoal - 1]) < windowDays)
    || (joinedContest && active.some((date, index) => index > 0 && daysBetween(active[index - 1], date) === 1))

  if (!cleared) {
    return {
      kind: 'getting-going',
      activeDays: recent.length,
      steps: { logged: active.length > 0, loggedToday: todayScore > 0, joinedContest, twoDaysInARow: streak >= 2 },
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
