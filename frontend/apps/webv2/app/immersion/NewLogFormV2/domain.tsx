import { z } from 'zod'
import { Unit } from '@app/immersion/api'
import { filterUnits, estimateScore } from '@app/immersion/NewLogForm/domain'

export { filterUnits, estimateScore }

export const NewLogFormV2Schema = z
  .object({
    languageCode: z.string().length(3, 'invalid language'),
    activityId: z.number(),
    amountValue: z
      .number({ invalid_type_error: 'Please enter a number' })
      .positive(),
    amountUnit: z.string(),
    allUnits: z.array(Unit),
    tags: z.array(z.string().max(50)).max(10, 'Maximum 10 tags allowed'),
    description: z.string().optional(),
  })
  .refine(
    log => {
      const unit = log.allUnits.find(it => it.id === log.amountUnit)
      return unit?.log_activity_id === log.activityId
    },
    {
      path: ['amountUnit'],
      message: 'This unit is cannot be used for this activity',
    },
  )

export type NewLogFormV2Schema = z.infer<typeof NewLogFormV2Schema>

export const NewLogV2APISchema = NewLogFormV2Schema.transform(log => ({
  language_code: log.languageCode,
  activity_id: log.activityId,
  amount: log.amountValue,
  unit_id: log.amountUnit,
  tags: log.tags,
  description: log.description,
}))

export type NewLogV2APISchema = z.infer<typeof NewLogV2APISchema>
