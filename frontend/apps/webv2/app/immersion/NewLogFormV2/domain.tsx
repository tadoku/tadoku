import { z } from 'zod'
import { Activity, Unit } from '@app/immersion/api'
import { filterUnits as filterLegacyUnits } from '@app/immersion/NewLogForm/domain'

const legacyReadingUnitKeys = new Set([
  'reading_two_column_page',
  'reading_comic_page',
])

export const filterUnits = (
  units: Unit[],
  activityId: number | undefined,
  languageCode: string | undefined,
) =>
  filterLegacyUnits(units, activityId, languageCode).filter(
    unit => !legacyReadingUnitKeys.has(unit.unit_key),
  )

const optionalNumber = z.preprocess(
  value =>
    typeof value === 'number' && Number.isNaN(value) ? undefined : value,
  z.number({ invalid_type_error: 'Please enter a number' }).optional(),
)

const optionalPositiveNumber = z.preprocess(
  value =>
    typeof value === 'number' && Number.isNaN(value) ? undefined : value,
  z.number({ invalid_type_error: 'Please enter a number' }).positive().optional(),
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
    allUnits: z.array(Unit),
    allActivities: z.array(Activity),
    tags: z.array(z.string().max(50)).max(10, 'Maximum 10 tags allowed'),
    description: z.string().optional(),
  })
  .superRefine((log, ctx) => {
    const activity = log.allActivities.find(it => it.id === log.activityId)
    const inputType = activity?.input_type ?? 'amount_primary'
    const unit = log.allUnits.find(it => it.id === log.amountUnit)
    const hasAmount = log.amountValue !== undefined
    const hasPositiveAmount =
      log.amountValue !== undefined && log.amountValue > 0
    const hasUnit = log.amountUnit !== undefined && log.amountUnit !== ''
    const hasDuration = log.durationMinutes !== undefined
    const hasValidAmountUnit =
      hasPositiveAmount && hasUnit && unit?.log_activity_id === log.activityId
    const durationSeconds = durationSecondsFromMinutes(log.durationMinutes)

    if (durationSeconds !== undefined && durationSeconds <= 0) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['durationMinutes'],
        message: 'Time spent must be at least 1 second',
      })
    }

    if (inputType === 'time_primary') {
      if (hasDuration) {
        return
      }

      if (hasAmount && !hasPositiveAmount) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ['amountValue'],
          message: 'Amount must be greater than 0',
        })
      }

      if (!hasValidAmountUnit) {
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

    if (unit?.log_activity_id !== log.activityId) {
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
  const unit = log.allUnits.find(it => it.id === log.amountUnit)
  const hasValidAmountUnit =
    log.amountValue !== undefined &&
    log.amountValue > 0 &&
    log.amountUnit !== undefined &&
    log.amountUnit !== '' &&
    unit?.log_activity_id === log.activityId
  const durationSeconds = durationSecondsFromMinutes(log.durationMinutes)
  const base = {
    language_code: log.languageCode,
    activity_id: log.activityId,
    tags: log.tags,
    description: log.description,
  }

  if (inputType === 'time_primary' && durationSeconds !== undefined) {
    return {
      ...base,
      duration_seconds: durationSeconds,
    }
  }

  if (hasValidAmountUnit) {
    return {
      ...base,
      amount: log.amountValue,
      unit_id: log.amountUnit,
      ...(durationSeconds !== undefined
        ? { duration_seconds: durationSeconds }
        : {}),
    }
  }

  return base
})

export type NewLogV2APISchema = z.infer<typeof NewLogV2APISchema>
