import { useStore } from '@tanstack/react-form'
import { useHydrated } from '@tanstack/react-router'
import type { ImmersionLogConfigurationOptions, ImmersionScoringRuleSet } from 'api-client'
import { Button, SectionHeading, Surface, usePaperForm, withPaperForm } from 'paper-ui'
import { useState } from 'react'
import type { LogFailure } from '../lib/logs'
import { useChangeScoring } from '../lib/scoring'
import {
  emptyRule,
  firstScoringDraft,
  publishedPlatformRuleSets,
  ruleUnits,
  scoreSources,
  scoringDraftSchema,
  toNewRuleSet,
  type ScoringDraft,
} from '../lib/scoring-draft'
import { scoringFailures } from './scoring-failures'

/** Creates a contest rule-set draft for the administrator who opened the page. A created draft starts a new form. */
export function RuleSetForm({ contestId, viewerId, options, platformRuleSets }: {
  contestId: string
  viewerId: string
  options: ImmersionLogConfigurationOptions
  platformRuleSets: ImmersionScoringRuleSet[]
}) {
  const hydrated = useHydrated()
  const changeScoring = useChangeScoring()
  const [outcome, setOutcome] = useState<{ ok: true } | { ok: false; reason: LogFailure } | null>(null)
  const [defaultValues] = useState(() => firstScoringDraft(options, platformRuleSets))
  const form = usePaperForm({
    defaultValues,
    validators: { onChange: scoringDraftSchema, onSubmit: scoringDraftSchema },
    onSubmit: async ({ value }) => {
      setOutcome(null)
      const result = await changeScoring({ create: toNewRuleSet(value, viewerId, contestId) })
      if (result.ok) form.reset(defaultValues, { keepDefaultValues: true })
      setOutcome(result.ok ? { ok: true } : result)
    },
  })
  const pending = useStore(form.store, state => state.isSubmitting)
  const mode = useStore(form.store, state => state.values.mode)
  const disabled = !hydrated || pending

  return (
    <Surface className="log-surface scoring-form">
      <SectionHeading
        level={2}
        title="New rule-set version"
        description="Uncovered inputs score zero. Override mode falls back to the pinned platform version; replace mode does not."
      />
      <form.AppForm>
        <form.Form className="goal-form" aria-busy={pending}>
          <fieldset className="goal-form__controls" disabled={disabled}>
            <form.AppField name="mode">
              {field => (
                <field.Select
                  label="Contest behavior"
                  options={[
                    { value: 'override', label: 'Override platform rules' },
                    { value: 'replace', label: 'Replace platform rules' },
                  ]}
                />
              )}
            </form.AppField>
            {mode === 'override' ? (
              <form.AppField name="fallbackRuleSetId">
                {field => (
                  <field.Select
                    label="Pinned platform version"
                    placeholder="Choose a version"
                    options={publishedPlatformRuleSets(platformRuleSets).map(ruleSet => ({
                      value: ruleSet.id,
                      label: `Platform v${ruleSet.version}${ruleSet.active ? ' (active)' : ''}`,
                    }))}
                  />
                )}
              </form.AppField>
            ) : null}
            <form.AppField name="rules" mode="array">
              {rules => (
                <fieldset className="contest-form-group">
                  <legend className="paper-field__label">Ordered rules</legend>
                  {rules.state.value.length === 0 ? <p>This replacing rule set will score every input as zero.</p> : null}
                  {rules.state.value.map((_, index) => (
                    <RuleFields
                      key={index}
                      form={form}
                      index={index}
                      count={rules.state.value.length}
                      options={options}
                      move={to => rules.moveValue(index, to)}
                      remove={() => rules.removeValue(index)}
                    />
                  ))}
                  <div>
                    <Button type="button" variant="outline" onClick={() => rules.pushValue(emptyRule(options))}>
                      Add rule
                    </Button>
                  </div>
                </fieldset>
              )}
            </form.AppField>
          </fieldset>

          {outcome ? (
            <p role={outcome.ok ? 'status' : 'alert'}>
              {outcome.ok ? 'Scoring rule-set draft created.' : scoringFailures(outcome.reason, 'The draft was not created.')}
            </p>
          ) : null}
          <div className="goal-actions">
            <Button type="submit" disabled={disabled}>{pending ? 'Creating…' : 'Create draft'}</Button>
          </div>
        </form.Form>
      </form.AppForm>
    </Surface>
  )
}

/** One rule's matchers, score source and rate, with its place in the order. */
const RuleFields = withPaperForm({
  defaultValues: {} as ScoringDraft,
  props: {
    index: 0,
    count: 0,
    options: {} as ImmersionLogConfigurationOptions,
    move: (() => {}) as (to: number) => void,
    remove: () => {},
  },
  render: function RuleFields({ form, index, count, options, move, remove }) {
    const activityId = useStore(form.store, state => state.values.rules[index]?.activityId ?? '')
    const units = ruleUnits(options, activityId)

    return (
      <fieldset className="scoring-rule">
        <legend className="scoring-rule__title">Rule {index + 1}</legend>
        <div className="scoring-rule__actions">
          <Button type="button" variant="ghost" disabled={index === 0} onClick={() => move(index - 1)}>Up</Button>
          <Button type="button" variant="ghost" disabled={index === count - 1} onClick={() => move(index + 1)}>
            Down
          </Button>
          <Button type="button" variant="ghost" onClick={remove}>Remove</Button>
        </div>
        <div className="goal-form__row">
          <form.AppField
            name={`rules[${index}].activityId`}
            // A unit belongs to one activity, so another activity's unit can't stay selected.
            listeners={{ onChange: () => form.setFieldValue(`rules[${index}].unitKey`, '') }}
          >
            {field => (
              <field.Select
                label="Activity"
                options={options.activities.map(activity => ({ value: String(activity.id), label: activity.name }))}
              />
            )}
          </form.AppField>
          <form.AppField name={`rules[${index}].scoreSource`}>
            {field => <field.Select label="Score source" options={scoreSources} />}
          </form.AppField>
          <form.AppField name={`rules[${index}].unitKey`}>
            {field => (
              <field.Select
                label="Unit matcher"
                options={[
                  { value: '', label: 'Any unit' },
                  ...units.map(unit => ({ value: unit.unit_key, label: `${unit.name} (${unit.unit_key})` })),
                ]}
              />
            )}
          </form.AppField>
          <form.AppField name={`rules[${index}].languageCode`}>
            {field => (
              <field.Select
                label="Language matcher"
                options={[
                  { value: '', label: 'Any language' },
                  ...options.languages.map(language => ({ value: language.code, label: language.name })),
                ]}
              />
            )}
          </form.AppField>
          <form.AppField name={`rules[${index}].tag`}>
            {field => <field.Input label="Tag matcher" placeholder="Any tag" />}
          </form.AppField>
          <form.AppField name={`rules[${index}].rate`}>
            {field => <field.Input label="Rate" type="number" inputMode="decimal" min={0} step="any" />}
          </form.AppField>
        </div>
        <form.AppField name={`rules[${index}].stackable`}>
          {field => (
            <field.Checkbox
              label="Stackable modifier"
              hint="Modifiers multiply the selected base rule and never score alone."
            />
          )}
        </form.AppField>
      </fieldset>
    )
  },
})
