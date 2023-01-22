import {
  AutocompleteInput,
  AutocompleteMultiInput,
  Input,
  RadioGroup,
} from 'ui'
import { FormProvider, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  ContestRegistrationsView,
  LogConfigurationOptions,
  useCreateLog,
} from '@app/contests/api'
import { useRouter } from 'next/router'
import { routes } from '@app/common/routes'
import {
  estimateScore,
  filterActivities,
  filterTags,
  filterUnits,
  formatContestLabel,
  NewLogAPISchema,
  NewLogFormSchema,
  trackingModesForRegistrations,
} from '@app/contests/NewLogForm/domain'
import { formatScore } from '@app/common/format'
import { useDebounce } from 'use-debounce'
import { useSessionOrRedirect } from '@app/common/session'
import { useEffect } from 'react'

interface Props {
  registrations: ContestRegistrationsView
  options: LogConfigurationOptions
}

export const LogForm = ({
  registrations: { registrations },
  options,
}: Props) => {
  const defaultValues: Partial<NewLogFormSchema> = {
    activity: options.activities[0],
    tracking_mode: registrations.length > 0 ? 'automatic' : 'personal',
    language:
      registrations.length > 0 ? registrations[0].languages[0] : undefined,
    unit: options.units.filter(
      it => it.log_activity_id === options.activities[0].id,
    )[0],
    registrations,
    selected_registrations: registrations,
  }

  const methods = useForm({
    resolver: zodResolver(NewLogFormSchema),
    defaultValues,
  })
  methods.trigger
  const [session] = useSessionOrRedirect()

  const trackingMode = methods.watch('tracking_mode') ?? 'personal'
  const activity = methods.watch('activity')
  const language = methods.watch('language')
  const unit = methods.watch('unit')
  const amount = methods.watch('amount')

  const languages =
    trackingMode === 'personal'
      ? options.languages
      : registrations.flatMap(it => it.languages)
  const tags = filterTags(options.tags, activity)
  const units = filterUnits(options.units, activity?.id, language)
  const activities = filterActivities(
    options.activities,
    registrations,
    trackingMode,
  )
  const estimatedScore = estimateScore(amount, unit)

  const router = useRouter()
  const createLogMutation = useCreateLog(() => {
    const userId = session?.identity.id
    if (userId) {
      return router.replace(routes.userProfileStatistics(userId))
    }

    router.replace(routes.leaderboardLatestOfficial())
  })

  const [createLog] = useDebounce(createLogMutation.mutate, 2500)

  const onSubmit = (data: any) => {
    createLog(NewLogAPISchema.parse(data))
  }

  useEffect(() => {
    const subscription = methods.watch((value, { name, type }) => {
      // reset unit if activity or language was changed
      if ((name === 'language' || name === 'activity') && type === 'change') {
        methods.setValue(
          'unit',
          filterUnits(options.units, value.activity?.id, language)?.[0],
        )
      }
    })
    return () => subscription.unsubscribe()
  }, [methods.watch])

  return (
    <FormProvider {...methods}>
      <form
        onSubmit={methods.handleSubmit(onSubmit, errors => console.log(errors))}
        className="v-stack spaced max-w-4xl"
      >
        <div className="card">
          <div className="v-stack spaced lg:h-stack lg:!space-x-8 w-full">
            <div className="flex-grow v-stack spaced">
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
            <div className="v-stack spaced">
              <AutocompleteInput
                name="language"
                label="Language"
                options={languages}
                match={(option, query) =>
                  option.name
                    .toLowerCase()
                    .replace(/[^a-zA-Z0-9]/g, '')
                    .includes(query.toLowerCase())
                }
                format={option => option.name}
              />
              <AutocompleteInput
                name="activity"
                label="Activity"
                options={activities}
                match={(option, query) =>
                  option.name
                    .toLowerCase()
                    .replace(/[^a-zA-Z0-9]/g, '')
                    .includes(query.toLowerCase())
                }
                format={option => option.name}
              />
              <div className="h-stack spaced">
                <div className="flex-grow">
                  <Input
                    name="amount"
                    label="Amount"
                    type="number"
                    defaultValue={0}
                    options={{ valueAsNumber: true }}
                    min={0}
                  />
                </div>
                <div className="min-w-[150px]">
                  <AutocompleteInput
                    name="unit"
                    label="Unit"
                    options={units}
                    match={(option, query) =>
                      option.name
                        .toLowerCase()
                        .replace(/[^a-zA-Z0-9]/g, '')
                        .includes(query.toLowerCase())
                    }
                    format={option => option.name}
                  />
                </div>
              </div>
              <AutocompleteMultiInput
                name="tags"
                label="Tags"
                options={tags}
                match={(option, query) =>
                  option.name
                    .toLowerCase()
                    .replace(/[^a-zA-Z0-9]/g, '')
                    .includes(query.toLowerCase())
                }
                getIdForOption={option => option.id}
                format={option => option.name}
              />
              <Input
                name="description"
                label="Description"
                type="text"
                placeholder="e.g. One Piece volume 45"
              />
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
    </FormProvider>
  )
}
