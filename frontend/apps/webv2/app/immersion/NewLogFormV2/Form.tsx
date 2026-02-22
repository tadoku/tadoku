import { Input, TagsInput } from 'ui'
import { TagsSidebar } from '@app/immersion/components/TagsSidebar'
import { FormProvider, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  fetchTagSuggestions,
  LogConfigurationOptions,
  useCreateLogV2,
  useOngoingContestRegistrations,
} from '@app/immersion/api'
import { useRouter } from 'next/router'
import { routes } from '@app/common/routes'
import {
  estimateScore,
  filterUnits,
  getInputType,
  NewLogFormV2Schema,
  NewLogV2APISchema,
} from '@app/immersion/NewLogFormV2/domain'
import { formatScore } from '@app/common/format'
import { useDebouncedCallback } from 'use-debounce'
import { useSessionOrRedirect } from '@app/common/session'
import { useEffect, useState } from 'react'
import { AmountWithUnit, Option, OptionGroup, Select } from 'ui/components/Form'

interface Props {
  options: LogConfigurationOptions
  defaultValues?: Partial<NewLogFormV2Schema>
}

export const LogFormV2 = ({ options, defaultValues: originalDefaultValues }: Props) => {
  const initialActivityId = options.activities[0].id
  const initialInputType = getInputType(options.activities, initialActivityId)
  const [showTimeInput, setShowTimeInput] = useState(false)

  const defaultValues: Partial<NewLogFormV2Schema> = {
    ...originalDefaultValues,
    activityId: initialActivityId,
    inputType: initialInputType,
    amountUnit: options.units.filter(
      it => it.log_activity_id === initialActivityId,
    )[0]?.id,
    allUnits: options.units,
  }

  const methods = useForm({
    resolver: zodResolver(NewLogFormV2Schema),
    defaultValues,
  })

  useSessionOrRedirect()

  const activityId = methods.watch('activityId')
  const languageCode = methods.watch('languageCode')
  const unitId = methods.watch('amountUnit')
  const amount = methods.watch('amountValue')
  const durationMinutes = methods.watch('durationMinutes')
  const inputType = getInputType(options.activities, activityId)

  const languagesAsOptions: Option[] = options.languages.map(it => ({
    value: it.code,
    label: it.name,
  }))

  const userLangSet = new Set(options.user_language_codes)
  const languageGroups: OptionGroup[] | undefined =
    userLangSet.size > 0
      ? [
          {
            label: 'Previously used',
            options: languagesAsOptions.filter(it => userLangSet.has(it.value)),
          },
          { label: 'All languages', options: languagesAsOptions },
        ]
      : undefined

  const activity = options.activities.find(it => it.id === activityId)
  const units = filterUnits(options.units, activity?.id, languageCode)
  const unitsAsOptions: Option[] = units.map(it => ({
    value: it.id,
    label: it.name,
  }))
  const currentSelectedUnit = units.find(it => it.id === unitId)
  const activitiesAsOptions: Option[] = options.activities.map(it => ({
    value: it.id.toString(),
    label: it.name,
  }))
  const estimatedScore = estimateScore(
    amount,
    currentSelectedUnit,
    durationMinutes,
    activity?.time_modifier ?? undefined,
  )

  // Eagerly prefetch ongoing registrations (non-blocking)
  const registrations = useOngoingContestRegistrations()

  const router = useRouter()
  const createLogMutation = useCreateLogV2(log => {
    const hasRegistrations =
      registrations.data &&
      registrations.data.registrations.length > 0
    if (hasRegistrations) {
      router.replace(routes.logContests(log.id) + '?preselect=1')
    } else {
      router.replace(routes.log(log.id))
    }
  })

  const createLog = useDebouncedCallback(createLogMutation.mutate, 2500, {
    leading: true,
    trailing: false,
  })

  const onSubmit = (data: any) => {
    createLog(NewLogV2APISchema.parse(data))
  }

  useEffect(() => {
    const subscription = methods.watch((value, { name, type }) => {
      if (name === 'activityId' && type === 'change') {
        const newInputType = getInputType(options.activities, value.activityId)
        methods.setValue('inputType', newInputType)
        setShowTimeInput(false)
        methods.setValue('durationMinutes', undefined)
      }
      if (
        (name === 'languageCode' || name === 'activityId') &&
        type === 'change'
      ) {
        const id = filterUnits(
          options.units,
          value.activityId,
          languageCode,
        )?.[0]?.id
        if (id !== methods.getValues('amountUnit')) {
          methods.setValue('amountUnit', id)
        }
      }
    })
    return () => subscription.unsubscribe()
  }, [methods, languageCode, options.units, options.activities])

  return (
    <FormProvider {...methods}>
      <div className="flex flex-col lg:flex-row lg:gap-6">
        <form
          onSubmit={methods.handleSubmit(onSubmit, errors => console.log(errors))}
          className="v-stack spaced max-w-lg flex-1"
        >
          <div className="card">
            <div className="v-stack spaced">
              <Select
                name="languageCode"
                label="Language"
                values={languagesAsOptions}
                groups={languageGroups}
              />
              <Select
                name="activityId"
                label="Activity"
                values={activitiesAsOptions}
                options={{ valueAsNumber: true }}
              />
              {inputType === 'time' ? (
                <Input
                  name="durationMinutes"
                  label="Time (minutes)"
                  type="number"
                  defaultValue={0}
                  min={0}
                  step="any"
                  options={{ valueAsNumber: true }}
                />
              ) : (
                <>
                  <AmountWithUnit
                    label="Amount"
                    name="amount"
                    defaultValue={0}
                    min={0}
                    step="any"
                    units={unitsAsOptions}
                    unitsLabel="Unit"
                  />
                  {showTimeInput ? (
                    <div className="h-stack items-end gap-2">
                      <div className="flex-1">
                        <Input
                          name="durationMinutes"
                          label="Time spent (minutes)"
                          type="number"
                          defaultValue={0}
                          min={0}
                          step="any"
                          options={{ valueAsNumber: true }}
                        />
                      </div>
                      <button
                        type="button"
                        className="btn ghost text-sm mb-0.5"
                        onClick={() => {
                          setShowTimeInput(false)
                          methods.setValue('durationMinutes', undefined)
                        }}
                      >
                        Remove
                      </button>
                    </div>
                  ) : (
                    <button
                      type="button"
                      className="text-sm text-primary hover:underline text-left"
                      onClick={() => setShowTimeInput(true)}
                    >
                      + Track time spent
                    </button>
                  )}
                </>
              )}
              <Input
                name="description"
                label="Description"
                type="text"
                placeholder="e.g. One Piece volume 45"
              />
              <TagsInput
                name="tags"
                label="Tags"
                placeholder="Add tags..."
                maxTags={10}
                getSuggestions={fetchTagSuggestions}
                renderSuggestion={s => (s.count > 0 ? `${s.tag} (${s.count}×)` : s.tag)}
                getValue={s => s.tag}
              />
              <div className="lg:hidden">
                <TagsSidebar activityId={activityId} />
              </div>
            </div>
            <div className="-mx-4 -mb-4 mt-4 px-4 py-2 md:-mx-7 md:-mb-7 md:px-7 md:py-2 bg-slate-500/5 text-center lg:text-right font-mono">
              Estimated score: <strong>{formatScore(estimatedScore)}</strong>
            </div>
          </div>
          <div className="h-stack spaced justify-end">
            <a href={routes.home()} className="btn ghost">
              Cancel
            </a>
            <button
              type="submit"
              className="btn primary"
              disabled={methods.formState.isSubmitting}
            >
              Create
            </button>
          </div>
        </form>
        <aside className="hidden lg:block lg:w-56 lg:pt-1">
          <div className="sticky top-14 sm:top-20">
            <TagsSidebar activityId={activityId} />
          </div>
        </aside>
      </div>
    </FormProvider>
  )
}
