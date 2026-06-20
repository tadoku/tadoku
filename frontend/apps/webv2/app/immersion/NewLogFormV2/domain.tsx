import { z } from 'zod'
import { Activity, Unit } from '@app/immersion/api'
import { filterUnits, estimateScore } from '@app/immersion/NewLogForm/domain'

export { filterUnits, estimateScore }

export const NewLogFormV2Schema = z
  .object({
    languageCode: z.string().length(3, 'invalid language'),
    activityId: z.number(),
    inputType: z.enum(['time', 'amount']).default('amount'),
    amountValue: z
      .number({ invalid_type_error: 'Please enter a number' })
      .optional(),
    amountUnit: z.string().optional(),
    durationMinutes: z
      .number({ invalid_type_error: 'Please enter a number' })
      .optional(),
    allUnits: z.array(Unit),
    tags: z.array(z.string().max(50)).max(10, 'Maximum 10 tags allowed'),
    description: z.string().optional(),
  })
  .superRefine((log, ctx) => {
    const hasTime = log.durationMinutes != null && log.durationMinutes > 0
    const hasAmount =
      log.amountValue != null && log.amountValue > 0 && log.amountUnit != null

    if (log.inputType === 'time') {
      if (!hasTime) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ['durationMinutes'],
          message: 'Please enter time',
        })
      }
      return
    }

    if (!hasAmount && !hasTime) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['amountValue'],
        message: 'Please enter an amount or time',
      })
    }

    if (log.amountValue != null && log.amountValue <= 0 && !hasTime) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['amountValue'],
        message: 'Number must be greater than 0',
      })
    }

    if (log.durationMinutes != null && log.durationMinutes <= 0 && !hasAmount) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['durationMinutes'],
        message: 'Number must be greater than 0',
      })
    }

    if (log.amountUnit) {
      const unit = log.allUnits.find(it => it.id === log.amountUnit)
      if (unit?.log_activity_id !== log.activityId) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ['amountUnit'],
          message: 'This unit cannot be used for this activity',
        })
      }
    }
  })
  .transform(log => {
    const includeAmount =
      log.inputType === 'amount' &&
      log.amountValue != null &&
      log.amountValue > 0 &&
      log.amountUnit != null

    return {
      ...log,
      amountValue: includeAmount ? log.amountValue : undefined,
      amountUnit: includeAmount ? log.amountUnit : undefined,
    }
  })

export type NewLogFormV2Schema = z.infer<typeof NewLogFormV2Schema>

export const NewLogV2APISchema = NewLogFormV2Schema.transform(log => ({
  language_code: log.languageCode,
  activity_id: log.activityId,
  ...(log.amountValue != null && log.amountUnit != null
    ? { amount: log.amountValue, unit_id: log.amountUnit }
    : {}),
  ...(log.durationMinutes != null
    ? { duration_seconds: Math.round(log.durationMinutes * 60) }
    : {}),
  tags: log.tags,
  description: log.description,
}))

export type NewLogV2APISchema = z.infer<typeof NewLogV2APISchema>

export const getInputType = (
  activities: Activity[],
  activityId: number | undefined,
): 'time' | 'amount' => {
  const activity = activities.find(it => it.id === activityId)
  return activity?.input_type ?? 'amount'
}
