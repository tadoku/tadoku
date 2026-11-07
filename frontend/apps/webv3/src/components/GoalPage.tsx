import * as stylex from '@stylexjs/stylex'
import { Button, buttonStyles, FieldSet, PageHeader, Surface, usePaperForm } from 'paper-ui'
import { pageStyles } from './page-styles'
import { border, breakpoints, color, font, fontSize, fontWeight, lineHeight, measure, size, space, type } from 'paper-ui/tokens.stylex'
import { useState, useSyncExternalStore } from 'react'
import { activities, describeGoal, goalSchema, storeGoal, useStoredGoal, type Goal } from '../lib/goal'
import { useLinks } from '../lib/links'

const styles = stylex.create({
  layout: {
    gap: { default: space[64], [breakpoints.tablet]: space[40] },
    alignItems: 'start',
    display: 'grid',
    gridTemplateColumns: {
      default: `minmax(0, ${size[560]}) minmax(0, ${size[300]})`,
      [breakpoints.tablet]: `minmax(0, ${size[560]})`,
    },
    maxInlineSize: `calc(${size[560]} + ${size[300]} + ${space[64]})`,
  },
  note: { marginInline: 0, color: color.textMuted, fontSize: type.metadataSize, marginBlockEnd: 0, marginBlockStart: `calc(-1 * ${space[12]})`, },
  status: {
    margin: 0,
    color: color.statusSuccess,
    display: { default: null, ':empty': 'none' },
    fontSize: fontSize[14],
    fontWeight: fontWeight.semibold,
  },
  failed: { color: color.statusDanger },
  aside: {
    borderBlockStartColor: color.ruleDefault,
    borderBlockStartStyle: 'solid',
    borderBlockStartWidth: { default: 0, [breakpoints.tablet]: border.static },
    fontSize: fontSize[14],
    lineHeight: lineHeight.airy,
    maxInlineSize: { default: null, [breakpoints.tablet]: measure.text },
    minInlineSize: 0,
    paddingBlockStart: { default: null, [breakpoints.tablet]: space[24] },
  },
  summary: { paddingInline: space[24], marginBlockEnd: space[32], paddingBlockEnd: space[24], paddingBlockStart: space[20], },
  summaryTitle: {
    marginInline: 0,
    color: color.textMuted,
    fontFamily: font.body,
    fontSize: fontSize[14],
    lineHeight: lineHeight.normal,
    marginBlockEnd: space[12],
    marginBlockStart: 0,
  },
  statement: {
    margin: 0,
    fontFamily: font.display,
    fontSize: fontSize[20],
    fontWeight: fontWeight.bold,
    lineHeight: lineHeight.statement,
    textWrap: 'balance',
  },
  total: { marginInline: 0, fontSize: fontSize[16], fontVariantNumeric: 'tabular-nums', fontWeight: fontWeight.semibold, marginBlockEnd: 0, marginBlockStart: space[8], },
})

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
      <div {...stylex.props(styles.layout)}>
        <form.AppForm><form.Form xstyle={pageStyles.form}>
          <FieldSet disabled={!hydrated}>
            <form.AppField name="activity">{field => <field.RadioSelect label="Activity" variant="segmented" options={activities} />}</form.AppField>
            <div {...stylex.props(pageStyles.formRow)}>
              <form.AppField name="minutes">{field => <field.Input id="goal-minutes" label="Minutes each day (5–600)" type="number" inputMode="numeric" min={5} max={600} step={1} />}</form.AppField>
              <form.AppField name="days">{field => <field.Input id="goal-days" label="Days each week (1–7)" type="number" inputMode="numeric" min={1} max={7} step={1} />}</form.AppField>
            </div>

            <p {...stylex.props(styles.note)}>Saved in this browser only; it doesn’t count towards contests.</p>
            <div {...stylex.props(pageStyles.actions)}>
              <Button type="submit">Save goal</Button>
              <p {...stylex.props(styles.status, result === 'failed' && styles.failed)} role="status" data-result={result ?? undefined}>
                {result === 'saved' ? 'Goal saved in this browser.' : result === 'failed' ? 'This browser didn’t allow saving. Allow site storage, then try again.' : ''}
              </p>
            </div>
          </FieldSet>
        </form.Form></form.AppForm>

        {saved ? (
          <aside {...stylex.props(styles.aside)} aria-label="Your goal">
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
    <Surface as="section" elevation="floating" accent inset="none" xstyle={styles.summary} aria-labelledby="goal-summary-title">
      <h2 id="goal-summary-title" {...stylex.props(styles.summaryTitle)}>Your daily goal</h2>
      <p {...stylex.props(styles.statement)}>{describeGoal(goal)}</p>
      <p {...stylex.props(styles.total)}>{formatMinutes(goal.minutes * goal.days)} each week</p>
      <a href={links.logs} {...stylex.props(...buttonStyles({ variant: 'outline' }))}>Log activity</a>
    </Surface>
  )
}
