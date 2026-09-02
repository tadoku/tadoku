import { z } from 'zod'

const nilUuid = '00000000-0000-0000-0000-000000000000'

export const targetUserIdSchema = z
  .string()
  .uuid()
  .refine(value => value !== nilUuid, 'target user ID must not be nil')

export const featureFlagKeySchema = z.enum(['release-log-entry-v2'])
export type FeatureFlagKey = z.infer<typeof featureFlagKeySchema>

export const featureAccessResultSchema = z
  .object({
    enabled: z.boolean(),
    changed: z.boolean(),
    environment: z.enum(['local', 'production']),
    revision: z.string().regex(/^[0-9a-f]{40}$/u),
  })
  .strict()

export type FeatureAccessResult = z.infer<typeof featureAccessResultSchema>
