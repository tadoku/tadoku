import { describe, expect, it } from 'vitest'
import {
  featureAccessResultSchema,
  featureFlagKeySchema,
  targetUserIdSchema,
} from './contracts'

const userId = '0198f6c5-c4af-7b1d-9776-884c065d72db'

describe('feature access contracts', () => {
  it('accepts only the registered flag and a non-nil target UUID', () => {
    expect(targetUserIdSchema.parse(userId)).toEqual(userId)

    expect(() =>
      targetUserIdSchema.parse('00000000-0000-0000-0000-000000000000'),
    ).toThrow()
    expect(() => featureFlagKeySchema.parse('unregistered-flag')).toThrow()
  })

  it('rejects PII and arbitrary environment input', () => {
    expect(() =>
      featureAccessResultSchema.parse({
        enabled: true,
        changed: true,
        environment: 'production',
        revision: 'a'.repeat(40),
        email: 'private@example.test',
      }),
    ).toThrow()
    expect(() =>
      featureAccessResultSchema.parse({
        enabled: true,
        environment: 'production',
        revision: 'a'.repeat(40),
      }),
    ).toThrow()
    expect(() =>
      featureAccessResultSchema.parse({
        enabled: true,
        changed: true,
        environment: 'attacker-controlled',
        revision: 'a'.repeat(40),
      }),
    ).toThrow()
  })
})
