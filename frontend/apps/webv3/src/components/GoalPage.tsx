import { useForm, useStore } from '@tanstack/react-form'
import { Button, PageHeader, Surface, buttonClassName } from 'paper-ui'
import { CheckIcon, iconClassName } from 'paper-ui/icons'
import { useMemo, useState, useSyncExternalStore, type ChangeEvent } from 'react'
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
  const form = useForm({
    // Until hydration the server snapshot is empty; then the stored goal becomes the untouched form's values.
    defaultValues: saved ?? firstGoal,
    validators: { onChange: goalSchema, onSubmit: goalSchema },
    onSubmit: ({ value }) => setResult(storeGoal(value) ? 'saved' : 'failed'),
    onSubmitInvalid: ({ formApi }) => {
      const invalid = (['minutes', 'days'] as const).find(name => formApi.getFieldMeta(name)?.errors.length)
      if (invalid) document.getElementById(`goal-${invalid}`)?.focus()
    },
  })

  // Errors wait for the first save attempt, then follow each change, so they never shift the layout mid-click.
  const submitted = useStore(form.store, state => state.submissionAttempts > 0)

  return (
    <div className="goal-page">
      <PageHeader
        title="Set a daily goal"
        lead="Decide how long you want to read or listen each day. Small, steady sessions add up over a round."
      />
      <div className="goal-layout">
        <form
          className="goal-form"
          noValidate
          onSubmit={event => {
            event.preventDefault()
            void form.handleSubmit()
          }}
        >
          <fieldset className="goal-form__controls" disabled={!hydrated}>
            <form.Field name="activity">
              {field => (
                <fieldset className="paper-choice-field paper-radio-select paper-radio-select--segmented">
                  <legend className="paper-field__label">Activity</legend>
                  <div className="paper-choice-list">
                    {activities.map(activity => (
                      <label className="paper-radio-select__segment" key={activity.value}>
                        <input
                          type="radio"
                          name={field.name}
                          value={activity.value}
                          checked={field.state.value === activity.value}
                          onChange={() => field.handleChange(activity.value)}
                          onBlur={field.handleBlur}
                        />
                        <span className="paper-radio-select__label">{activity.label}</span>
                        <CheckIcon className={iconClassName('compact', 'paper-radio-select__check')} aria-hidden="true" />
                      </label>
                    ))}
                  </div>
                </fieldset>
              )}
            </form.Field>

            <div className="goal-form__row">
              <form.Field name="minutes">
                {field => (
                  <NumberField
                    id="goal-minutes"
                    label="Minutes each day"
                    hint="Between 5 and 600."
                    min={5}
                    max={600}
                    value={field.state.value}
                    error={submitted ? errorText(field.state.meta.errors) : undefined}
                    onChange={event => field.handleChange(event.target.valueAsNumber)}
                    onBlur={field.handleBlur}
                  />
                )}
              </form.Field>
              <form.Field name="days">
                {field => (
                  <NumberField
                    id="goal-days"
                    label="Days each week"
                    hint="Between 1 and 7."
                    min={1}
                    max={7}
                    value={field.state.value}
                    error={submitted ? errorText(field.state.meta.errors) : undefined}
                    onChange={event => field.handleChange(event.target.valueAsNumber)}
                    onBlur={field.handleBlur}
                  />
                )}
              </form.Field>
            </div>

            <div className="goal-actions">
              <Button type="submit">Save goal</Button>
              <p className="goal-status" role="status" data-result={result ?? undefined}>
                {result === 'saved' ? 'Goal saved in this browser.' : result === 'failed' ? 'This browser didn’t allow saving. Allow site storage, then try again.' : ''}
              </p>
            </div>
          </fieldset>
        </form>

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

function errorText(errors: readonly unknown[]) {
  const error = errors[0]
  return typeof error === 'string' ? error : (error as { message?: string } | undefined)?.message
}

/** Paper's field recipe on a native number input, so any form library can own the value. */
function NumberField({ id, label, hint, min, max, value, error, onChange, onBlur }: {
  id: string
  label: string
  hint: string
  min: number
  max: number
  value: number
  error?: string
  onChange: (event: ChangeEvent<HTMLInputElement>) => void
  onBlur: () => void
}) {
  return (
    <div className={`paper-field${error ? ' paper-field--invalid' : ''}`}>
      <label className="paper-field__label" htmlFor={id}>{label}</label>
      <input
        id={id}
        className="paper-input"
        type="number"
        inputMode="numeric"
        min={min}
        max={max}
        step={1}
        value={Number.isNaN(value) ? '' : value}
        aria-invalid={error ? true : undefined}
        aria-describedby={`${id}-hint${error ? ` ${id}-error` : ''}`}
        onChange={onChange}
        onBlur={onBlur}
      />
      <p className="paper-field__hint" id={`${id}-hint`}>{hint}</p>
      {error ? <p className="paper-field__error" id={`${id}-error`} role="alert">{error}</p> : null}
    </div>
  )
}
