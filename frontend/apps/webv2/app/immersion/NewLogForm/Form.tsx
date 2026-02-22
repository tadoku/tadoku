import {
  AutocompleteMultiInput,
  Input,
  RadioGroup,
  TagsInput,
} from 'ui'
import { TagsSidebar } from '@app/immersion/components/TagsSidebar'
import { FormProvider, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  ContestRegistrationsView,
  fetchTagSuggestions,
  LogConfigurationOptions,
  useCreateLog,
} from '@app/immersion/api'
import { useRouter } from 'next/router'
import { routes } from '@app/common/routes'
import {
  estimateScore,
  filterActivities,
  filterUnits,
  formatContestLabel,
  NewLogAPISchema,
  NewLogFormSchema,
  trackingModesForRegistrations,
} from '@app/immersion/NewLogForm/domain'
import { formatScore } from '@app/common/format'
import { useDebouncedCallback } from 'use-debounce'
import { useSessionOrRedirect } from '@app/common/session'
import { useEffect } from 'react'
import { AmountWithUnit, Option, Select } from 'ui/components/Form'

interface Props {
  registrations: ContestRegistrationsView
  options: LogConfigurationOptions
  defaultValues?: Partial<NewLogFormSchema>
}

export const LogForm = ({
  registrations: { registrations },
  options,
  defaultValues: originalDefaultValues,
}: Props) => {
  const defaultValues: Partial<NewLogFormSchema> = {
    ...originalDefaultValues,
    activityId: options.activities[0].id,
    tracking_mode: registrations.length > 0 ? 'automatic' : 'personal',
    languageCode:
      registrations.length > 0 ? registrations[0].languages[0].code : undefined,
    amountUnit: options.units.filter(
      it => it.log_activity_id === options.activities[0].id,
    )[0]?.id,
    registrations,
    selected_registrations: registrations,
    allUnits: options.units,
  }

  const methods = useForm({
    resolver: zodResolver(NewLogFormSchema),
    defaultValues,
  })
  methods.trigger

  useSessionOrRedirect()

  const trackingMode = methods.watch('tracking_mode') ?? 'personal'
  const activityId = methods.watch('activityId')
  const languageCode = methods.watch('languageCode')
  const unitId = methods.watch('amountUnit')
  const amount = methods.watch('amountValue')

  const languages =
    trackingMode === 'personal'
      ? options.languages
      : registrations
          .flatMap(it => it.languages)
          .filter(
            ({ code }, index, self) =>
              index === self.findIndex(it => it.code === code),
          )
          .sort((a, b) => {
            if (a.name < b.name) {
              return -1
            }
            if (a.name > b.name) {
              return 1
            }
            return 0
          })
  const languagesAsOptions: Option[] = languages.map(it => ({
    value: it.code,
    label: it.name,
  }))

  const activity = options.activities.find(it => it.id === activityId)
  const units = filterUnits(options.units, activity?.id, languageCode)
  const unitsAsOptions: Option[] = units.map(it => ({
    value: it.id,
    label: it.name,
  }))
  const currentSelectedUnit = units.find(it => it.id === unitId)
  const activities = filterActivities(
    options.activities,
    registrations,
    trackingMode,
  )
  const activitiesAsOptions: Option[] = activities.map(it => ({
    value: it.id.toString(),
    label: it.name,
  }))
  const estimatedScore = estimateScore(amount, currentSelectedUnit)

  const router = useRouter()
  const createLogMutation = useCreateLog(id => {
    router.replace(routes.log(id))
  })

  const createLog = useDebouncedCallback(createLogMutation.mutate, 2500, {
    leading: true,
    trailing: false,
  })

  const onSubmit = (data: any) => {
    createLog(NewLogAPISchema.parse(data))
  }

  useEffect(() => {
    const subscription = methods.watch((value, { name, type }) => {
      // reset unit if activity or language was changed
      if (
        (name === 'languageCode' || name === 'activityId') &&
        type === 'change'
      ) {
        // sus
        const id = filterUnits(
          options.units,
          value.activityId,
          languageCode,
        )?.[0].id
        if (id !== methods.getValues('amountUnit')) {
          methods.setValue('amountUnit', id)
        }
      }
    })
    return () => subscription.unsubscribe()
  }, [methods, languageCode, options.units])

  return (
    <FormProvider {...methods}>
      <div className="flex flex-col lg:flex-row lg:gap-6">
        <form
          onSubmit={methods.handleSubmit(onSubmit, errors => console.log(errors))}
          className="v-stack spaced max-w-4xl flex-1"
        >
          <div className="card">
            <div className="v-stack spaced lg:h-stack lg:!space-x-8 w-full">
              <div className="flex-grow v-stack spaced lg:w-2/5">
                <RadioGroup
                  options={trackingModesForRegistrations(registrations.length)}
                  label="Contests"
                  name="tracking_mode"
                />
                {trackingMode === 'manual' ? (
                  <AutocompleteMultiInput
                    name="selected_registrations"
                    label="Contest selection"
                    options={registrations}
                    match={(option, query) =>
                      option.contest?.title
                        .toLowerCase()
                        .replace(/[^a-zA-Z0-9]/g, '')
                        .includes(query.toLowerCase()) ?? false
                    }
                    getIdForOption={option => option.id}
                    format={option => formatContestLabel(option.contest!)}
                  />
                ) : null}
              </div>
              <div className="v-stack spaced lg:w-3/5">
                <Select
                  name="languageCode"
                  label="Language"
                  values={languagesAsOptions}
                />
                <Select
                  name="activityId"
                  label="Activity"
                  values={activitiesAsOptions}
                  options={{ valueAsNumber: true }}
                />
                <div className="h-stack spaced">
                  <AmountWithUnit
                    label="Amount"
                    name="amount"
                    defaultValue={0}
                    min={0}
                    step="any"
                    units={unitsAsOptions}
                    unitsLabel="Unit"
                  />
                </div>
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
                  renderSuggestion={s => s.count > 0 ? `${s.tag} (${s.count}×)` : s.tag}
                  getValue={s => s.tag}
                />
                <div className="lg:hidden">
                  <TagsSidebar activityId={activityId} />
                </div>
              </div>
            </div>
            <div className="-mx-4 -mb-4 mt-4 px-4 py-2 md:-mx-7 md:-mb-7 md:px-7 md:py-2 bg-slate-500/5 text-center lg:text-right font-mono">
              Estimated score: <strong>{formatScore(estimatedScore)}</strong>
            </div>
          </div>
          <div className="h-stack spaced justify-end">
            <a href={routes.contestListOfficial()} className="btn ghost">
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
          <div className="sticky top-4">
            <TagsSidebar activityId={activityId} />
          </div>
        </aside>
      </div>
    </FormProvider>
  )
}
