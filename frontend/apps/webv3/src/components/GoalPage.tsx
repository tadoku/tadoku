import { Button, PageHeader, Surface, buttonClassName, usePaperForm } from 'paper-ui'
import { useState, useSyncExternalStore } from 'react'
import { activities, describeGoal, goalSchema, storeGoal, useStoredGoal, type Goal } from '../lib/goal'
import { useLinks } from '../lib/links'
import { formActions, formControls, formNote, formRow } from './form-styles'

const firstGoal: Goal = { activity: 'reading', minutes: 30, days: 5 }

const subscribeToNothing = () => () => {}

function formatMinutes(total: number) {
  const hours = Math.floor(total / 60)
  const minutes = total % 60
  return [hours ? `${hours} h` : '', minutes ? `${minutes} min` : ''].filter(Boolean).join(' ')
}

export function GoalPage() {
  // Controls stay disabled until hydration, so nothing typed into the server-rendered form is lost.
  const hydrated = useSyncExternalStore(subscribeToNothing, () => true, () => false)
  const saved = useStoredGoal()
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
    <div>
      <PageHeader title="Set a daily goal" />
      <div className="grid max-w-goal grid-cols-goal items-start gap-16 max-tablet:grid-cols-goal-narrow max-tablet:gap-10">
        <form.AppForm><form.Form className="min-w-0">
          <fieldset className={formControls} disabled={!hydrated}>
            <form.AppField name="activity">{field => <field.RadioSelect label="Activity" variant="segmented" options={activities} />}</form.AppField>
            <div className={formRow}>
              <form.AppField name="minutes">{field => <field.Input id="goal-minutes" label="Minutes each day (5–600)" type="number" inputMode="numeric" min={5} max={600} step={1} />}</form.AppField>
              <form.AppField name="days">{field => <field.Input id="goal-days" label="Days each week (1–7)" type="number" inputMode="numeric" min={1} max={7} step={1} />}</form.AppField>
            </div>

            <p className={formNote}>Saved in this browser only; it doesn’t count towards contests.</p>
            <div className={formActions}>
              <Button type="submit">Save goal</Button>
              <p
                className={`m-0 text-sm font-semibold empty:hidden ${result === 'failed' ? 'text-danger' : 'text-success'}`}
                role="status"
                data-result={result ?? undefined}
              >
                {result === 'saved' ? 'Goal saved in this browser.' : result === 'failed' ? 'This browser didn’t allow saving. Allow site storage, then try again.' : ''}
              </p>
            </div>
          </fieldset>
        </form.Form></form.AppForm>

        {saved ? (
          <aside className="min-w-0 text-sm leading-170 max-tablet:max-w-measure-note max-tablet:border-t max-tablet:border-rule max-tablet:pt-6" aria-label="Your goal">
            <GoalSummary goal={saved} />
          </aside>
        ) : null}
      </div>
    </div>
  )
}

function GoalSummary({ goal }: { goal: Goal }) {
  const links = useLinks()
  return (
    <Surface as="section" accent className="mb-8 px-6 pb-6 pt-5" aria-labelledby="goal-summary-title">
      <h2 id="goal-summary-title" className="mb-3 mt-0 font-body text-sm leading-150 text-muted">Your daily goal</h2>
      <p className="m-0 text-balance font-display text-xl leading-145 font-bold">{describeGoal(goal)}</p>
      <p className="mb-0 mt-2 text-base font-semibold tabular-nums">{formatMinutes(goal.minutes * goal.days)} each week</p>
      <a href={links.logs} className={buttonClassName({ variant: 'outline' })}>Log activity</a>
    </Surface>
  )
}
