import { Button, PageHeader, Surface, buttonClassName, usePaperForm } from 'paper-ui'
import { useMemo, useState, useSyncExternalStore } from 'react'
import { z } from 'zod'
import { useLinks } from '../lib/links'

const storageKey = 'tadoku-daily-goal'

const activities = [
  { value: 'reading', label: 'Reading', phrase: 'reading' },
  { value: 'listening', label: 'Listening', phrase: 'listening' },
  { value: 'both', label: 'Both', phrase: 'reading or listening' },
] as const

const goalSchema = z.object({
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

type Goal = z.infer<typeof goalSchema>

const firstGoal: Goal = { activity: 'reading', minutes: 30, days: 5 }

// The goal lives in localStorage, which the server cannot read: the server snapshot is always "no goal".
const subscribeToStorage = (onChange: () => void) => {
  addEventListener('storage', onChange)
  return () => removeEventListener('storage', onChange)
}

const subscribeToNothing = () => () => {}

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

function storeGoal(goal: Goal) {
  try {
    localStorage.setItem(storageKey, JSON.stringify(goal))
  } catch {
    return false
  }
  dispatchEvent(new StorageEvent('storage', { key: storageKey }))
  return true
}

function formatMinutes(total: number) {
  const hours = Math.floor(total / 60)
  const minutes = total % 60
  return [hours ? `${hours} h` : '', minutes ? `${minutes} min` : ''].filter(Boolean).join(' ')
}

export function GoalPage() {
  const links = useLinks()
  // Controls stay disabled until hydration, so nothing typed into the server-rendered form is lost.
  const hydrated = useSyncExternalStore(subscribeToNothing, () => true, () => false)
  const storedGoal = useSyncExternalStore(subscribeToStorage, readStoredGoal, () => null)
  const saved = useMemo(() => parseGoal(storedGoal), [storedGoal])
  const [result, setResult] = useState<'saved' | 'failed' | null>(null)
  const form = usePaperForm({
    // Until hydration the server snapshot is empty; then the stored goal becomes the untouched form's values.
    defaultValues: saved ?? firstGoal,
    validators: { onChange: goalSchema, onSubmit: goalSchema },
    onSubmit: ({ value }) => setResult(storeGoal(value) ? 'saved' : 'failed'),
    onSubmitInvalid: ({ formApi }) => {
      const invalid = (['minutes', 'days'] as const).find(name => formApi.getFieldMeta(name)?.errors.length)
      if (invalid) document.getElementById(`goal-${invalid}`)?.focus()
    },
  })

  return (
    <div className="goal-page">
      <PageHeader
        title="Set a daily goal"
        lead="Decide how long you want to read or listen each day. Small, steady sessions add up over a round."
      />
      <div className="goal-layout">
        <form.AppForm><form.Form className="goal-form">
          <fieldset className="goal-form__controls" disabled={!hydrated}>
            <form.AppField name="activity">{field => <field.RadioSelect label="Activity" variant="segmented" options={activities} />}</form.AppField>
            <div className="goal-form__row">
              <form.AppField name="minutes">{field => <field.Input id="goal-minutes" label="Minutes each day" hint="Between 5 and 600." type="number" inputMode="numeric" min={5} max={600} step={1} />}</form.AppField>
              <form.AppField name="days">{field => <field.Input id="goal-days" label="Days each week" hint="Between 1 and 7." type="number" inputMode="numeric" min={1} max={7} step={1} />}</form.AppField>
            </div>

            <div className="goal-actions">
              <Button type="submit">Save goal</Button>
              <p className="goal-status" role="status" data-result={result ?? undefined}>
                {result === 'saved' ? 'Goal saved in this browser.' : result === 'failed' ? 'This browser didn’t allow saving. Allow site storage, then try again.' : ''}
              </p>
            </div>
          </fieldset>
        </form.Form></form.AppForm>

        <aside className="goal-aside" aria-label="Your goal">
          {saved ? <GoalSummary goal={saved} /> : null}
          <h2>How goals work</h2>
          <p>Your goal is kept in this browser. It doesn’t log activity or change your Tadoku account.</p>
          <p>After each session, log what you actually read or listened to. That is what counts towards contests.</p>
          <a className="text-link" href={links.guide}>How Tadoku works</a>
        </aside>
      </div>
    </div>
  )
}

function GoalSummary({ goal }: { goal: Goal }) {
  const links = useLinks()
  const phrase = activities.find(activity => activity.value === goal.activity)?.phrase
  return (
    <Surface as="section" elevation="floating" accent className="goal-summary" aria-labelledby="goal-summary-title">
      <h2 id="goal-summary-title">Your daily goal</h2>
      <p className="goal-summary__statement">
        {goal.minutes} minutes of {phrase}, {goal.days === 7 ? 'every day' : `${goal.days} ${goal.days === 1 ? 'day' : 'days'} a week`}
      </p>
      <p className="goal-summary__total">{formatMinutes(goal.minutes * goal.days)} each week</p>
      <p className="goal-summary__note">Saved in this browser.</p>
      <a href={links.logs} className={buttonClassName({ variant: 'outline' })}>Log activity</a>
    </Surface>
  )
}
