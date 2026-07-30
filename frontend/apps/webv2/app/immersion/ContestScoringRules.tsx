import { zodResolver } from '@hookform/resolvers/zod'
import {
  LogConfigurationOptions,
  ScoringRuleSet,
  ScoringRuleSetDraftPayload,
  useActivateScoringRuleSet,
  useCreateContestScoringRuleSet,
  usePublishScoringRuleSet,
} from '@app/immersion/api'
import { FormProvider, useFieldArray, useForm } from 'react-hook-form'
import { z } from 'zod'
import { Checkbox, Flash, Input, Select } from 'ui'
import { Option } from 'ui/components/Form'
import { toast } from 'react-toastify'

const RuleFormSchema = z.object({
  stackable: z.boolean(),
  activity_id: z.number(),
  unit_key: z.string(),
  language_code: z.string(),
  tag: z.string().max(50),
  score_source: z.enum(['amount', 'duration_minutes']),
  rate: z.number().min(0),
})

const ContestScoringFormSchema = z
  .object({
    mode: z.enum(['override', 'replace']),
    fallback_rule_set_id: z.string(),
    rules: z.array(RuleFormSchema),
  })
  .superRefine((value, ctx) => {
    if (value.mode === 'override' && !value.fallback_rule_set_id) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['fallback_rule_set_id'],
        message: 'Choose the pinned platform version',
      })
    }
  })

type ContestScoringForm = z.infer<typeof ContestScoringFormSchema>

interface Props {
  contestId: string
  options: LogConfigurationOptions
  platformRuleSets: ScoringRuleSet[]
  contestRuleSets: ScoringRuleSet[]
  locked: boolean
}

const emptyRule = (): ContestScoringForm['rules'][number] => ({
  stackable: false,
  activity_id: 1,
  unit_key: '',
  language_code: '',
  tag: '',
  score_source: 'amount',
  rate: 1,
})

export const ContestScoringRules = ({
  contestId,
  options,
  platformRuleSets,
  contestRuleSets,
  locked,
}: Props) => {
  const publishedPlatformRuleSets = platformRuleSets.filter(
    ruleSet => ruleSet.status === 'published',
  )
  const activePlatformRuleSet =
    publishedPlatformRuleSets.find(ruleSet => ruleSet.active) ??
    publishedPlatformRuleSets[0]
  const methods = useForm<ContestScoringForm>({
    resolver: zodResolver(ContestScoringFormSchema),
    defaultValues: {
      mode: 'override',
      fallback_rule_set_id: activePlatformRuleSet?.id ?? '',
      rules: [emptyRule()],
    },
  })
  const rules = useFieldArray({ control: methods.control, name: 'rules' })
  const mode = methods.watch('mode')
  const ruleValues = methods.watch('rules')
  const createRuleSet = useCreateContestScoringRuleSet(contestId)
  const publishRuleSet = usePublishScoringRuleSet(contestId)
  const activateRuleSet = useActivateScoringRuleSet(contestId)

  const activityOptions: Option[] = options.activities.map(activity => ({
    value: activity.id.toString(),
    label: activity.name,
  }))
  const unitOptionsForActivity = (activityId: number): Option[] => [
    { value: '', label: 'Any unit' },
    ...options.units
      .filter(unit => unit.log_activity_id === activityId)
      .filter(
        (unit, index, units) =>
          units.findIndex(candidate => candidate.unit_key === unit.unit_key) ===
          index,
      )
      .map(unit => ({
        value: unit.unit_key,
        label: `${unit.name} (${unit.unit_key})`,
      })),
  ]
  const languageOptions: Option[] = [
    { value: '', label: 'Any language' },
    ...options.languages.map(language => ({
      value: language.code,
      label: language.name,
    })),
  ]
  const fallbackOptions: Option[] = publishedPlatformRuleSets.map(ruleSet => ({
    value: ruleSet.id,
    label: `Platform v${ruleSet.version}${ruleSet.active ? ' (active)' : ''}`,
  }))

  const onSubmit = async (form: ContestScoringForm) => {
    const payload: ScoringRuleSetDraftPayload = {
      mode: form.mode,
      ...(form.mode === 'override'
        ? { fallback_rule_set_id: form.fallback_rule_set_id }
        : {}),
      rules: form.rules.map((rule, index) => ({
        priority: index * 10,
        stackable: rule.stackable,
        activity_id: rule.activity_id,
        ...(rule.unit_key ? { unit_key: rule.unit_key } : {}),
        ...(rule.language_code
          ? { language_code: rule.language_code }
          : {}),
        ...(rule.tag.trim() ? { tag: rule.tag.trim() } : {}),
        score_source: rule.score_source,
        rate: rule.rate,
      })),
    }
    try {
      await createRuleSet.mutateAsync(payload)
      toast.success('Scoring rule-set draft created')
    } catch {
      toast.error('Could not create scoring rule-set draft')
    }
  }

  return (
    <div className="v-stack spaced">
      <Flash visible={locked} style="warning">
        Scoring is locked because this contest has started. Published snapshots
        and active rules remain visible below.
      </Flash>

      {!locked ? (
        <FormProvider {...methods}>
          <form
            className="card v-stack spaced"
            onSubmit={methods.handleSubmit(onSubmit)}
          >
            <div>
              <h2 className="subtitle">New rule-set version</h2>
              <p className="text-slate-600">
                Uncovered inputs score zero. Override mode falls back to the
                pinned platform version; replace mode does not.
              </p>
            </div>
            <Select
              name="mode"
              label="Contest behavior"
              values={[
                { value: 'override', label: 'Override platform rules' },
                { value: 'replace', label: 'Replace platform rules' },
              ]}
            />
            {mode === 'override' ? (
              <Select
                name="fallback_rule_set_id"
                label="Pinned platform version"
                values={fallbackOptions}
              />
            ) : null}

            <div className="v-stack spaced">
              <div className="h-stack justify-between items-center">
                <h3 className="font-semibold">Ordered rules</h3>
                <button
                  type="button"
                  className="btn secondary"
                  onClick={() => rules.append(emptyRule())}
                >
                  Add rule
                </button>
              </div>
              {rules.fields.length === 0 ? (
                <p className="text-slate-600">
                  This replacing rule set will score every input as zero.
                </p>
              ) : null}
              {rules.fields.map((field, index) => (
                <div key={field.id} className="card bg-slate-500/5">
                  <div className="h-stack justify-between items-center mb-3">
                    <strong>Rule {index + 1}</strong>
                    <div className="h-stack spaced">
                      <button
                        type="button"
                        className="btn ghost"
                        disabled={index === 0}
                        onClick={() => rules.move(index, index - 1)}
                      >
                        Up
                      </button>
                      <button
                        type="button"
                        className="btn ghost"
                        disabled={index === rules.fields.length - 1}
                        onClick={() => rules.move(index, index + 1)}
                      >
                        Down
                      </button>
                      <button
                        type="button"
                        className="btn danger"
                        onClick={() => rules.remove(index)}
                      >
                        Remove
                      </button>
                    </div>
                  </div>
                  <div className="grid md:grid-cols-2 gap-3">
                    <Select
                      name={`rules.${index}.activity_id`}
                      label="Activity"
                      values={activityOptions}
                      options={{ valueAsNumber: true }}
                    />
                    <Select
                      name={`rules.${index}.score_source`}
                      label="Score source"
                      values={[
                        { value: 'amount', label: 'Amount' },
                        {
                          value: 'duration_minutes',
                          label: 'Duration (minutes)',
                        },
                      ]}
                    />
                    <Select
                      name={`rules.${index}.unit_key`}
                      label="Unit matcher"
                      values={unitOptionsForActivity(
                        ruleValues[index]?.activity_id ?? 1,
                      )}
                    />
                    <Select
                      name={`rules.${index}.language_code`}
                      label="Language matcher"
                      values={languageOptions}
                    />
                    <Input
                      name={`rules.${index}.tag`}
                      label="Tag matcher"
                      placeholder="Any tag"
                    />
                    <Input
                      name={`rules.${index}.rate`}
                      label="Rate"
                      type="number"
                      min={0}
                      step="any"
                      options={{ valueAsNumber: true }}
                    />
                  </div>
                  <Checkbox
                    name={`rules.${index}.stackable`}
                    label="Stackable modifier"
                    hint="Modifiers multiply the selected base rule and never score alone."
                  />
                </div>
              ))}
            </div>
            <div className="h-stack justify-end">
              <button
                type="submit"
                className="btn primary"
                disabled={createRuleSet.isLoading}
              >
                Create draft
              </button>
            </div>
          </form>
        </FormProvider>
      ) : null}

      <section className="v-stack spaced">
        <h2 className="subtitle">Rule-set versions</h2>
        {contestRuleSets.length === 0 ? (
          <p>No contest-specific rule sets yet.</p>
        ) : null}
        {contestRuleSets.map(ruleSet => (
          <div key={ruleSet.id} className="card">
            <div className="h-stack justify-between items-start">
              <div>
                <h3 className="font-semibold">
                  Version {ruleSet.version}
                  {ruleSet.active ? ' · Active' : ''}
                </h3>
                <p>
                  {ruleSet.mode === 'override'
                    ? 'Overrides platform rules with a pinned fallback'
                    : 'Replaces platform rules; uncovered inputs score zero'}
                  {' · '}
                  {ruleSet.status}
                </p>
              </div>
              {!locked ? (
                <div className="h-stack spaced">
                  {ruleSet.status === 'draft' ? (
                    <button
                      type="button"
                      className="btn secondary"
                      disabled={publishRuleSet.isLoading}
                      onClick={async () => {
                        if (
                          !window.confirm(
                            'Publish this version? Published rules are immutable.',
                          )
                        ) {
                          return
                        }
                        try {
                          await publishRuleSet.mutateAsync(ruleSet.id)
                          toast.success('Rule-set version published')
                        } catch {
                          toast.error('Could not publish rule-set version')
                        }
                      }}
                    >
                      Publish
                    </button>
                  ) : null}
                  {ruleSet.status === 'published' && !ruleSet.active ? (
                    <button
                      type="button"
                      className="btn primary"
                      disabled={activateRuleSet.isLoading}
                      onClick={async () => {
                        if (
                          !window.confirm(
                            'Activate this version for new contest submissions?',
                          )
                        ) {
                          return
                        }
                        try {
                          await activateRuleSet.mutateAsync(ruleSet.id)
                          toast.success('Rule-set version activated')
                        } catch {
                          toast.error('Could not activate rule-set version')
                        }
                      }}
                    >
                      Activate
                    </button>
                  ) : null}
                </div>
              ) : null}
            </div>
            <ol className="mt-3 list-decimal pl-5">
              {ruleSet.rules.map(rule => (
                <li key={rule.id ?? rule.priority}>
                  {rule.stackable ? 'Modifier' : 'Base'} · activity{' '}
                  {rule.activity_id} · {rule.score_source} × {rule.rate}
                  {rule.unit_key ? ` · unit ${rule.unit_key}` : ''}
                  {rule.language_code
                    ? ` · language ${rule.language_code}`
                    : ''}
                  {rule.tag ? ` · tag ${rule.tag}` : ''}
                </li>
              ))}
            </ol>
          </div>
        ))}
      </section>
    </div>
  )
}
