import * as stylex from '@stylexjs/stylex'
import { useStore } from '@tanstack/react-form'
import { useHydrated } from '@tanstack/react-router'
import type { ImmersionLogConfigurationOptions, ImmersionScoringRuleSet } from 'api-client'
import { Button, FieldSet, SectionHeading, Surface, usePaperForm, withPaperForm } from 'paper-ui'
import { pageStyles } from './page-styles'
import { border, color, fontWeight, space } from 'paper-ui/tokens.stylex'
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

const styles = stylex.create({
  rule: {
    margin: 0,
    padding: space[16],
    borderColor: color.ruleDefault,
    borderStyle: 'solid',
    borderWidth: border.static,
    gap: space[16],
    display: 'grid',
    minInlineSize: 0,
  },
  ruleTitle: { padding: 0, float: 'left', fontWeight: fontWeight.semibold, width: '100%', },
  ruleActions: { gap: space[8], display: 'flex', flexWrap: 'wrap', },
})

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
    <Surface inset="fluid">
      <SectionHeading
        level={2}
        title="New rule-set version"
        description="Uncovered inputs score zero. Override mode falls back to the pinned platform version; replace mode does not."
      />
      <form.AppForm>
        <form.Form xstyle={pageStyles.form} aria-busy={pending}>
          <FieldSet disabled={disabled}>
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
                <fieldset {...stylex.props(pageStyles.formGroup)}>
                  <legend {...stylex.props(pageStyles.formGroupLegend)}>Ordered rules</legend>
                  {rules.state.value.length === 0 ? <p {...stylex.props(pageStyles.flush)}>This replacing rule set will score every input as zero.</p> : null}
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
          </FieldSet>

          {outcome ? (
            <p role={outcome.ok ? 'status' : 'alert'}>
              {outcome.ok ? 'Scoring rule-set draft created.' : scoringFailures(outcome.reason, 'The draft was not created.')}
            </p>
          ) : null}
          <div {...stylex.props(pageStyles.actions)}>
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
      <fieldset {...stylex.props(styles.rule)} data-scoring-rule="">
        <legend {...stylex.props(styles.ruleTitle)}>Rule {index + 1}</legend>
        <div {...stylex.props(styles.ruleActions)}>
          <Button type="button" variant="ghost" disabled={index === 0} onClick={() => move(index - 1)}>Up</Button>
          <Button type="button" variant="ghost" disabled={index === count - 1} onClick={() => move(index + 1)}>
            Down
          </Button>
          <Button type="button" variant="ghost" onClick={remove}>Remove</Button>
        </div>
        <div {...stylex.props(pageStyles.formRow)}>
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
