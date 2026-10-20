import { z } from 'zod'

export const featureFlagRegistry = {
  'release-log-entry-v2': { defaultValue: false },
  'identity-v2': { defaultValue: false },
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

/** The registered decisions in an API answer, with the safe default for any it lacks. */
export const registeredDecisions = (
  decisions: Partial<Record<FeatureFlagKey, boolean>> = {},
) =>
  featureFlagKeys.reduce<FeatureFlagDecisions>((result, flagKey) => {
    const decision = decisions[flagKey]
    result[flagKey] =
      typeof decision === 'boolean'
        ? decision
        : featureFlagRegistry[flagKey].defaultValue
    return result
  }, {} as FeatureFlagDecisions)

export const featureFlagSubjectSchema = z.string().uuid()
