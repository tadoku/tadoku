import { z } from 'zod'

export const featureFlagRegistry = {
  'release-log-entry-v2': { defaultValue: false },
} as const

export type FeatureFlagKey = keyof typeof featureFlagRegistry
export type FeatureFlagDecisions = Record<FeatureFlagKey, boolean>

export const featureFlagKeys = Object.keys(
  featureFlagRegistry,
) as FeatureFlagKey[]

export const defaultFeatureFlagDecisions =
  featureFlagKeys.reduce<FeatureFlagDecisions>((decisions, flagKey) => {
    decisions[flagKey] = featureFlagRegistry[flagKey].defaultValue
    return decisions
  }, {} as FeatureFlagDecisions)

export const featureFlagDecisionsSchema = z.object({
  'release-log-entry-v2': z.boolean(),
})

export const featureFlagResponseSchema = z.object({
  decisions: featureFlagDecisionsSchema,
})

export const featureFlagSubjectSchema = z.string().uuid()

export type FeatureFlagResponse = z.infer<typeof featureFlagResponseSchema>
