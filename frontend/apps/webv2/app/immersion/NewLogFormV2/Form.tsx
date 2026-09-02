import { Input, TagsInput } from 'ui'
import { TagsSidebar } from '@app/immersion/components/TagsSidebar'
import { FormProvider, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  fetchTagSuggestions,
  LogConfigurationOptions,
  useCreateLogV2,
  useOngoingContestRegistrations,
  useScorePreview,
} from '@app/immersion/api'
import { useRouter } from 'next/router'
import { routes } from '@app/common/routes'
import {
  filterUnits,
  NewLogFormV2Schema,
  NewLogV2APISchema,
} from '@app/immersion/NewLogFormV2/domain'
import { useDebounce, useDebouncedCallback } from 'use-debounce'
import { useSessionOrRedirect } from '@app/common/session'
import { useEffect } from 'react'
import { AmountWithUnit, Option, OptionGroup, Select } from 'ui/components/Form'
import { ScorePreviewEstimate } from '@app/immersion/components/ScorePreviewEstimate'
import {
  getLastLoggedActivityId,
  getLastLoggedLanguage,
  storeLastLoggedPreferences,
} from '@app/immersion/NewLogFormV2/preferences'

interface Props {
  options: LogConfigurationOptions
  defaultValues?: Partial<NewLogFormV2Schema>
}

export const LogFormV2 = ({
  options,
  defaultValues: originalDefaultValues,
}: Props) => {
  const defaultActivity = options.activities[0]
  const defaultActivityInputType =
    defaultActivity?.input_type ?? 'amount_primary'
  const defaultValues: Partial<NewLogFormV2Schema> = {
    ...originalDefaultValues,
    activityId: defaultActivity.id,
    amountUnit:
      defaultActivityInputType === 'amount_primary'
        ? filterUnits(
            options.units,
            defaultActivity.id,
            originalDefaultValues?.languageCode,
          )[0]?.id
        : undefined,
    allUnits: options.units,
    allActivities: options.activities,
  }

  const methods = useForm({
    resolver: zodResolver(NewLogFormV2Schema),
    defaultValues,
  })

  useSessionOrRedirect()

  const activityId = methods.watch('activityId')
  const languageCode = methods.watch('languageCode')
  const formValues = methods.watch()

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
  const activityInputType = activity?.input_type ?? 'amount_primary'
  const usesAmountUnit = activityInputType === 'amount_primary'
  const units = filterUnits(options.units, activity?.id, languageCode)
  const unitsAsOptions: Option[] = units.map(it => ({
    value: it.id,
    label: it.name,
  }))
  const activitiesAsOptions: Option[] = options.activities.map(it => ({
    value: it.id.toString(),
    label: it.name,
  }))
  const previewPayloadResult = NewLogV2APISchema.safeParse(formValues)
  const [previewPayloadJSON] = useDebounce(
    previewPayloadResult.success
      ? JSON.stringify(previewPayloadResult.data)
      : undefined,
    300,
  )
  const previewPayload = previewPayloadJSON
    ? JSON.parse(previewPayloadJSON)
    : undefined
  const scorePreview = useScorePreview(previewPayload)

  // Eagerly prefetch ongoing registrations (non-blocking)
  const registrations = useOngoingContestRegistrations()

  const router = useRouter()
  const createLogMutation = useCreateLogV2(log => {
    storeLastLoggedPreferences(log.language.code, log.activity.id)

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
    let restoredLanguageCode = methods.getValues('languageCode')
    let restoredActivityId = methods.getValues('activityId')
    let restoredPreference = false

    if (originalDefaultValues?.languageCode === undefined) {
      const lastLoggedLanguage = getLastLoggedLanguage()
      if (
        lastLoggedLanguage !== null &&
        options.languages.some(
          language => language.code === lastLoggedLanguage,
        )
      ) {
        restoredLanguageCode = lastLoggedLanguage
        methods.setValue('languageCode', lastLoggedLanguage)
        restoredPreference = true
      }
    }

    if (originalDefaultValues?.activityId === undefined) {
      const lastLoggedActivityId = getLastLoggedActivityId(
        options.activities,
      )
      if (lastLoggedActivityId !== null) {
        restoredActivityId = lastLoggedActivityId
        methods.setValue('activityId', lastLoggedActivityId)
        restoredPreference = true
      }
    }

    if (!restoredPreference) {
      return
    }

    const restoredActivity = options.activities.find(
      activity => activity.id === restoredActivityId,
    )
    if ((restoredActivity?.input_type ?? 'amount_primary') === 'time_primary') {
      methods.setValue('amountValue', undefined)
      methods.setValue('amountUnit', undefined)
      return
    }

    methods.setValue(
      'amountUnit',
      filterUnits(
        options.units,
        restoredActivityId,
        restoredLanguageCode,
      )[0]?.id,
    )
  }, [
    methods,
    options.activities,
    options.languages,
    options.units,
    originalDefaultValues?.activityId,
    originalDefaultValues?.languageCode,
  ])

  useEffect(() => {
    const subscription = methods.watch((value, { name, type }) => {
      if (
        (name === 'languageCode' || name === 'activityId') &&
        type === 'change'
      ) {
        const selectedActivity = options.activities.find(
          it => it.id === value.activityId,
        )
        const selectedInputType =
          selectedActivity?.input_type ?? 'amount_primary'

        if (selectedInputType === 'time_primary') {
          methods.setValue('amountValue', undefined)
          methods.setValue('amountUnit', undefined)
          return
        }

        const id = filterUnits(
          options.units,
          value.activityId,
          value.languageCode,
        )?.[0]?.id
        if (id !== methods.getValues('amountUnit')) {
          methods.setValue('amountUnit', id)
        }
      }
    })
    return () => subscription.unsubscribe()
  }, [methods, options.activities, options.units])

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
              {usesAmountUnit && (
                <AmountWithUnit
                  label="Amount"
                  name="amount"
                  defaultValue={0}
                  min={0}
                  step="any"
                  units={unitsAsOptions}
                  unitsLabel="Unit"
                />
              )}
              <Input
                name="durationMinutes"
                label="Time spent"
                type="number"
                min={0}
                step="any"
                hint="minutes"
                options={{ valueAsNumber: true }}
              />
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
              <ScorePreviewEstimate
                preview={scorePreview.data}
              />
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
