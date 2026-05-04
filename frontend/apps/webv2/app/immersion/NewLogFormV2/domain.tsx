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
      .positive()
      .optional(),
    amountUnit: z.string().optional(),
    durationMinutes: z
      .number({ invalid_type_error: 'Please enter a number' })
      .positive()
      .optional(),
    allUnits: z.array(Unit),
    tags: z.array(z.string().max(50)).max(10, 'Maximum 10 tags allowed'),
    description: z.string().optional(),
  })
  .refine(
    log => {
      if (log.inputType === 'time') {
        return log.durationMinutes != null && log.durationMinutes > 0
      }
      // amount type: need amount+unit or duration
      const hasAmount = log.amountValue != null && log.amountUnit != null
      const hasTime = log.durationMinutes != null && log.durationMinutes > 0
      return hasAmount || hasTime
    },
    {
      path: ['amountValue'],
      message: 'Please enter an amount or time',
    },
  )
  .refine(
    log => {
      if (log.inputType === 'time') return true
      if (!log.amountUnit) return true
      const unit = log.allUnits.find(it => it.id === log.amountUnit)
      return unit?.log_activity_id === log.activityId
    },
    {
      path: ['amountUnit'],
      message: 'This unit cannot be used for this activity',
    },
  )

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

