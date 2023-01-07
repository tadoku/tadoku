import {
  AutocompleteInput,
  AutocompleteMultiInput,
  Input,
  RadioGroup,
} from 'ui'
import { FormProvider, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import {
  Activity,
  ContestRegistrationsView,
  ContestRegistrationView,
  Language,
  LogConfigurationOptions,
  Tag,
  Unit,
} from '@app/contests/api'
import { useRouter } from 'next/router'
import { routes } from '@app/common/routes'
import {
  AdjustmentsHorizontalIcon,
  LinkIcon,
  UserIcon,
} from '@heroicons/react/20/solid'
import { RadioProps } from 'ui/components/Form'
import { DateTime, Duration, Interval } from 'luxon'

export const LogFormSchema = z.object({
  trackingMode: z.enum(['automatic', 'manual', 'personal']),
  registrations: z.array(ContestRegistrationView),
  selectedRegistrations: z.array(ContestRegistrationView),
  language: Language,
  activity: Activity,
  amount: z.number().positive(),
  unit: Unit,
  tags: z
    .array(z.string())
    .min(1, 'Must select at least one tag')
    .max(3, 'Must select three or fewer'),
  description: z.string().optional(),
})

export type LogFormSchema = z.infer<typeof LogFormSchema>

interface Props {
  registrations: ContestRegistrationsView
  options: LogConfigurationOptions
}

const filterUnits = (
  units: Unit[],
  activity: Activity | undefined,
  language: Language | undefined,
) => {
  if (!activity) {
    return []
  }

  const base = units.filter(it => {
    return it.logActivityId == activity.id
  })

  const grouped = base.reduce((acc, unit) => {
    if (!acc.has(unit.name)) {
      acc.set(unit.name, [])
    }

    acc.get(unit.name)?.push(unit)

    return acc
  }, new Map<string, Unit[]>())

  const filteredUnits = []
  for (const units of grouped.values()) {
    const unitForCurrentLanguage = units.find(
      it => it.languageCode === language?.code,
    )
    const fallback = units.find(it => it.languageCode === undefined)

    if (units.length > 1 && unitForCurrentLanguage) {
      filteredUnits.push(unitForCurrentLanguage)
    } else if (fallback) {
      filteredUnits.push(fallback)
    }
  }

  return filteredUnits
}

const filterTags = (tags: Tag[], activity: Activity | undefined) => {
  if (!activity) {
    return []
  }

  return tags.filter(it => it.logActivityId === activity.id)
}

const filterActivities = (
  activities: Activity[],
  registrations: ContestRegistrationsView['registrations'],
  trackingMode: LogFormSchema['trackingMode'],
) => {
  if (trackingMode === 'personal') {
    return activities
  }

  const acts = []

  const ids = new Set(
    registrations.flatMap(it => it.contest?.allowedActivities.map(it => it.id)),
  )

  return activities.filter(it => ids.has(it.id))
}

const trackingModesForRegistrations = (registrationCount: number) => {
  const personalOnly = registrationCount === 0

  return [
    {
      value: 'automatic',
      label: 'Automatic',
      description: 'Submit log to all eligible contests',
      IconComponent: LinkIcon,
      disabled: personalOnly,
      title: personalOnly ? 'No eligible contests found' : undefined,
    },
    {
      value: 'manual',
      label: 'Manual',
      description: 'Choose which contests to submit to',
      IconComponent: AdjustmentsHorizontalIcon,
      disabled: personalOnly,
      title: personalOnly ? 'No eligible contests found' : undefined,
    },
    {
      value: 'personal',
      label: 'Personal',
      description: 'Do not submit to any contests',
      IconComponent: UserIcon,
    },
  ] satisfies RadioProps['options']
}

export const LogForm = ({
  registrations: { registrations },
  options,
}: Props) => {
  const defaultValues: Partial<LogFormSchema> = {
    activity: options.activities[0],
    trackingMode: registrations.length > 0 ? 'automatic' : 'personal',
    language:
      registrations.length > 0 ? registrations[0].languages[0] : undefined,
    unit: options.units.filter(
      it => it.logActivityId === options.activities[0].id,
    )[0],
  }

  const methods = useForm({
    resolver: zodResolver(LogFormSchema),
    defaultValues,
  })

  const trackingMode = methods.watch('trackingMode') ?? 'personal'
  const activity = methods.watch('activity')
  const language = methods.watch('language')

  const languages =
    trackingMode === 'personal'
      ? options.languages
      : registrations.flatMap(it => it.languages)
  const tags = filterTags(options.tags, activity)
  const units = filterUnits(options.units, activity, language)
  const activities = filterActivities(
    options.activities,
    registrations,
    trackingMode,
  )

  const router = useRouter()
  // const createContestMutation = useCreateContest(id =>
  //   router.replace(routes.contestLeaderboard(id)),
  // )

  // const [createContest] = useDebounce(createContestMutation.mutate, 2500)

  const onSubmit = (data: any) => {
    console.log(data)
    // createContest(data)
  }

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
                name="trackingMode"
              />
              {trackingMode === 'manual' ? (
                <AutocompleteMultiInput
                  name="selectedRegistrations"
                  label="Contest selection"
                  options={registrations}
                  match={(option, query) =>
                    option.contest?.description
                      .toLowerCase()
                      .replace(/[^a-zA-Z0-9]/g, '')
                      .includes(query.toLowerCase()) ?? false
                  }
                  getIdForOption={option => option.id}
                  format={option =>
                    `${option.contest!.private ? '' : 'Official: '}${
                      option.contest?.description ?? option.contestId
                    } (${option.contest!.contestStart.toLocaleString(
                      DateTime.DATE_MED,
                    )} ~ ${option.contest!.contestEnd.toLocaleString(
                      DateTime.DATE_MED,
                    )})`
                  }
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
            Estimated score: <strong>0</strong>
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
