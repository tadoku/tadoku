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
  tracking_mode: z.enum(['automatic', 'manual', 'personal']),
  registrations: z.array(ContestRegistrationView),
  selected_registrations: z.array(ContestRegistrationView),
  language: Language,
  activity: Activity,
  amount: z.number().positive(),
  unit: Unit,
  tags: z.array(Tag).max(3, 'Must select three or fewer'),
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
    return it.log_activity_id == activity.id
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
      it => it.language_code === language?.code,
    )
    const fallback = units.find(it => it.language_code === undefined)

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

  return tags.filter(it => it.log_activity_id === activity.id)
}

export const filterActivities = (
  activities: Activity[],
  registrations: ContestRegistrationsView['registrations'],
  trackingMode: LogFormSchema['tracking_mode'],
) => {
  if (trackingMode === 'personal') {
    return activities
  }

  const acts = []

  const ids = new Set(
    registrations.flatMap(it =>
      it.contest?.allowed_activities.map(it => it.id),
    ),
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
  trackingMode: LogFormSchema['tracking_mode']
  language: Language
  activity: Activity
}) => {
  if (trackingMode === 'personal') {
    return []
  }

  const eligibleContests = registrations
    .filter(it => it.contest)
    .filter(it => it.languages.includes(language))
    .filter(it => it.contest!.allowed_activities.includes(activity))
    .filter(it =>
      Interval.fromDateTimes(
        it.contest!.contest_start,
        it.contest!.contest_end,
      ).contains(DateTime.now()),
    )

  const eligibleContestIds = new Set(eligibleContests.map(it => it.contest_id))

  if (trackingMode === 'manual') {
    for (const registration of manualContests) {
      if (!eligibleContestIds.has(registration.contest_id)) {
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
  } (${contest.contest_start.toLocaleString(
    DateTime.DATE_MED,
  )} ~ ${contest.contest_end.toLocaleString(DateTime.DATE_MED)})`
