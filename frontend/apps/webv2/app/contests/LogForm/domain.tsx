import { z } from 'zod'
import {
  Activity,
  ContestRegistrationsView,
  ContestRegistrationView,
  ContestView,
  Language,
  Tag,
  Unit,
} from '@app/contests/api'
import {
  AdjustmentsHorizontalIcon,
  LinkIcon,
  UserIcon,
} from '@heroicons/react/20/solid'
import { RadioProps } from 'ui/components/Form'
import { DateTime, Interval } from 'luxon'

export const LogFormSchema = z.object({
  trackingMode: z.enum(['automatic', 'manual', 'personal']),
  registrations: z.array(ContestRegistrationView),
  selectedRegistrations: z.array(ContestRegistrationView),
  language: Language,
  activity: Activity,
  amount: z.number().positive(),
  unit: z.object({
    id: z.string(),
    logActivityId: z.number(),
    name: z.string(),
    modifier: z.number(),
    languageCode: z.string().nullable().optional(),
  }),
  tags: z
    .array(
      z.object({
        id: z.string(),
        logActivityId: z.number(),
        name: z.string(),
      }),
    )
    .max(3, 'Must select three or fewer'),
  description: z.string().optional(),
})

export type LogFormSchema = z.infer<typeof LogFormSchema>

export const filterUnits = (
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

export const filterTags = (tags: Tag[], activity: Activity | undefined) => {
  if (!activity) {
    return []
  }

  return tags.filter(it => it.logActivityId === activity.id)
}

export const filterActivities = (
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

export const trackingModesForRegistrations = (registrationCount: number) => {
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

export const estimateScore = (
  amount: number | undefined,
  unit: Unit | undefined,
) => {
  if (!amount || !unit) {
    return undefined
  }

  return amount * unit.modifier
}

export const contestsForLog = ({
  registrations,
  manualContests,
  trackingMode,
  language,
  activity,
}: {
  registrations: ContestRegistrationsView['registrations']
  manualContests: ContestRegistrationsView['registrations']
  trackingMode: LogFormSchema['trackingMode']
  language: Language
  activity: Activity
}) => {
  if (trackingMode === 'personal') {
    return []
  }

  const eligibleContests = registrations
    .filter(it => it.contest)
    .filter(it => it.languages.includes(language))
    .filter(it => it.contest!.allowedActivities.includes(activity))
    .filter(it =>
      Interval.fromDateTimes(
        it.contest!.contestStart,
        it.contest!.contestEnd,
      ).contains(DateTime.now()),
    )

  const eligibleContestIds = new Set(eligibleContests.map(it => it.contestId))

  if (trackingMode === 'manual') {
    for (const registration of manualContests) {
      if (!eligibleContestIds.has(registration.contestId)) {
        throw Error(
          `Contest "${formatContestLabel(
            registration.contest!,
          )}" is does not allow this log to be submitted`,
        )
      }
    }

    return manualContests
  }

  return eligibleContests
}

export const formatContestLabel = (contest: ContestView) =>
  `${contest.private ? '' : 'Official: '}${
    contest.description
  } (${contest.contestStart.toLocaleString(
    DateTime.DATE_MED,
  )} ~ ${contest.contestEnd.toLocaleString(DateTime.DATE_MED)})`
