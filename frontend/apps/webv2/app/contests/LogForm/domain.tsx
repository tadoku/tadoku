import { z } from 'zod'
import {
  Activity,
  ContestRegistrationsView,
  ContestRegistrationView,
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
