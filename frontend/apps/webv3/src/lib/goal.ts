import { useMemo, useSyncExternalStore } from 'react'
import { z } from 'zod'

/** The daily goal lives in this browser's localStorage only; it never reaches the account. */
const storageKey = 'tadoku-daily-goal'

export const activities = [
  { value: 'reading', label: 'Reading', phrase: 'reading' },
  { value: 'listening', label: 'Listening', phrase: 'listening' },
  { value: 'both', label: 'Both', phrase: 'reading or listening' },
] as const

export const goalSchema = z.object({
  activity: z.enum(['reading', 'listening', 'both'], { error: 'Choose reading, listening or both.' }),
  minutes: z.number({ error: 'Enter how many minutes, for example 30.' })
    .int('Use whole minutes.')
    .min(5, 'Set at least 5 minutes.')
    .max(600, 'Set 600 minutes or fewer.'),
  days: z.number({ error: 'Enter how many days, from 1 to 7.' })
    .int('Use whole days.')
    .min(1, 'Choose at least 1 day.')
    .max(7, 'A week has 7 days.'),
})

export type Goal = z.infer<typeof goalSchema>


// The goal lives in localStorage, which the server cannot read: the server snapshot is always "no goal".
const subscribeToStorage = (onChange: () => void) => {
  addEventListener('storage', onChange)
  return () => removeEventListener('storage', onChange)
}

function readStoredGoal() {
  try { return localStorage.getItem(storageKey) } catch { return null }
}

function parseGoal(json: string | null): Goal | null {
  try {
    const result = goalSchema.safeParse(JSON.parse(json ?? 'null'))
    return result.success ? result.data : null
  } catch {
    return null
  }
}

export function storeGoal(goal: Goal) {
  try {
    localStorage.setItem(storageKey, JSON.stringify(goal))
  } catch {
    return false
  }
  dispatchEvent(new StorageEvent('storage', { key: storageKey }))
  return true
}

/** The stored goal after hydration; the server and the first client render see none. */
export function useStoredGoal(): Goal | null {
  const stored = useSyncExternalStore(subscribeToStorage, readStoredGoal, () => null)
  return useMemo(() => parseGoal(stored), [stored])
}

/** For example "30 minutes of reading, 5 days a week". */
export function describeGoal(goal: Goal) {
  const phrase = activities.find(activity => activity.value === goal.activity)?.phrase
  return `${goal.minutes} minutes of ${phrase}, ${goal.days === 7 ? 'every day' : `${goal.days} ${goal.days === 1 ? 'day' : 'days'} a week`}`
}
