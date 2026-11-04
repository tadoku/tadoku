import { zImmersionActivity, zImmersionUnit } from 'api-client'
import { z } from 'zod'
import { Unit } from '@app/immersion/api'

export const filterUnits = (
  units: Unit[],
  activityId: number | undefined,
  languageCode: string | undefined,
) => {
  if (!activityId) {
    return []
  }

  const base = units.filter(unit => unit.log_activity_id === activityId)

  const grouped = base.reduce((acc, unit) => {
    if (!acc.has(unit.unit_key)) {
      acc.set(unit.unit_key, [])
    }

    acc.get(unit.unit_key)?.push(unit)

    return acc
  }, new Map<string, Unit[]>())

  const filteredUnits = []
  for (const units of grouped.values()) {
    const unitForCurrentLanguage = units.find(
      it => it.language_code === languageCode,
    )
    const fallback = units.find(it => it.language_code === undefined)

    if (unitForCurrentLanguage) {
      filteredUnits.push(unitForCurrentLanguage)
    } else if (fallback) {
      filteredUnits.push(fallback)
    }
  }

  return filteredUnits
}

const activityHasUnit = (
  units: Unit[],
  activityId: number,
  unitKey: string | undefined,
) =>
  units.some(
    unit => unit.unit_key === unitKey && unit.log_activity_id === activityId,
  )

const optionalNumber = z.preprocess(
  (value: number | undefined) =>
    typeof value === 'number' && Number.isNaN(value) ? undefined : value,
  z.number({ error: 'Please enter a number' }).optional(),
)

const optionalPositiveNumber = z.preprocess(
  (value: number | undefined) =>
    typeof value === 'number' && Number.isNaN(value) ? undefined : value,
  z.number({ error: 'Please enter a number' }).positive().optional(),
)

const durationSecondsFromMinutes = (minutes?: number) =>
  minutes === undefined ? undefined : Math.round(minutes * 60)

export const NewLogFormV2Schema = z
  .object({
    languageCode: z.string().length(3, 'invalid language'),
    activityId: z.number(),
    amountValue: optionalNumber,
    amountUnit: z.string().optional(),
    durationMinutes: optionalPositiveNumber,
    allUnits: z.array(zImmersionUnit),
    allActivities: z.array(zImmersionActivity),
    tags: z.array(z.string().max(50)).max(10, 'Maximum 10 tags allowed'),
    description: z
      .string()
      .max(255, 'Description must be 255 characters or fewer')
      .optional(),
  })
  .superRefine((log, ctx) => {
    const activity = log.allActivities.find(it => it.id === log.activityId)
    const inputType = activity?.input_type ?? 'amount_primary'
    const hasAmount = log.amountValue !== undefined
    const hasPositiveAmount =
      log.amountValue !== undefined && log.amountValue > 0
    const hasUnit = log.amountUnit !== undefined && log.amountUnit !== ''
    const durationSeconds = durationSecondsFromMinutes(log.durationMinutes)

    if (durationSeconds !== undefined && durationSeconds <= 0) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['durationMinutes'],
        message: 'Time spent must be at least 1 second',
      })
    }

    // Timed activities are logged only by duration; any amount or unit left in the form is ignored.
    if (inputType === 'time_primary') {
      if (durationSeconds === undefined) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ['durationMinutes'],
          message: 'Time spent is required',
        })
      }
      return
    }

    if (!hasAmount) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['amountValue'],
        message: 'Amount is required',
      })
    } else if (!hasPositiveAmount) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['amountValue'],
        message: 'Amount must be greater than 0',
      })
    }

    if (!hasUnit) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['amountUnit'],
        message: 'Unit is required',
      })
      return
    }

    if (!activityHasUnit(log.allUnits, log.activityId, log.amountUnit)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['amountUnit'],
        message: 'This unit cannot be used for this activity',
      })
    }
  })

export type NewLogFormV2Schema = z.infer<typeof NewLogFormV2Schema>

export const NewLogV2APISchema = NewLogFormV2Schema.transform(log => {
  const activity = log.allActivities.find(it => it.id === log.activityId)
  const inputType = activity?.input_type ?? 'amount_primary'
  const durationSeconds = durationSecondsFromMinutes(log.durationMinutes)
  const base = {
    language_code: log.languageCode,
    activity_id: log.activityId,
    tags: log.tags,
    description: log.description,
  }

  if (inputType === 'time_primary') {
    return durationSeconds === undefined
      ? base
      : { ...base, duration_seconds: durationSeconds }
  }

  if (
    log.amountValue !== undefined &&
    log.amountValue > 0 &&
    log.amountUnit !== undefined &&
    activityHasUnit(log.allUnits, log.activityId, log.amountUnit)
  ) {
    return {
      ...base,
      amount: log.amountValue,
      unit_key: log.amountUnit,
      ...(durationSeconds !== undefined
        ? { duration_seconds: durationSeconds }
        : {}),
    }
  }

  return base
})

export type NewLogV2APISchema = z.infer<typeof NewLogV2APISchema>
