import { Button, PageHeader, Surface, buttonClassName, usePaperForm } from 'paper-ui'
import { useState, useSyncExternalStore } from 'react'
import { activities, describeGoal, goalSchema, storeGoal, useStoredGoal, type Goal } from '../lib/goal'
import { useLinks } from '../lib/links'

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
    <div className="goal-page">
      <PageHeader title="Set a daily goal" />
      <div className="goal-layout">
        <form.AppForm><form.Form className="goal-form">
          <fieldset className="goal-form__controls" disabled={!hydrated}>
            <form.AppField name="activity">{field => <field.RadioSelect label="Activity" variant="segmented" options={activities} />}</form.AppField>
            <div className="goal-form__row">
              <form.AppField name="minutes">{field => <field.Input id="goal-minutes" label="Minutes each day (5–600)" type="number" inputMode="numeric" min={5} max={600} step={1} />}</form.AppField>
              <form.AppField name="days">{field => <field.Input id="goal-days" label="Days each week (1–7)" type="number" inputMode="numeric" min={1} max={7} step={1} />}</form.AppField>
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
          <p className="goal-aside__note">Your goal stays in this browser; it doesn’t change your account or count towards contests.</p>
        </aside>
      </div>
    </div>
  )
}

function GoalSummary({ goal }: { goal: Goal }) {
  const links = useLinks()
  return (
    <Surface as="section" elevation="floating" accent className="goal-summary" aria-labelledby="goal-summary-title">
      <h2 id="goal-summary-title">Your daily goal</h2>
      <p className="goal-summary__statement">{describeGoal(goal)}</p>
      <p className="goal-summary__total">{formatMinutes(goal.minutes * goal.days)} each week</p>
      <a href={links.logs} className={buttonClassName({ variant: 'outline' })}>Log activity</a>
    </Surface>
  )
}
