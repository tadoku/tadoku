import { readFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { defaultFeatureFlagDecisions, featureFlagKeys } from './registry'

type ContractFixture = {
  booleanFlags: Record<string, { safeDefault: boolean }>
}

describe('feature flag cross-stack contract', () => {
  it('matches the shared pilot key and safe default', async () => {
    const path = fileURLToPath(
      new URL('../../../../../feature-flags.contract.json', import.meta.url),
    )
    const fixture = JSON.parse(await readFile(path, 'utf8')) as ContractFixture

    expect(Object.keys(fixture.booleanFlags)).toEqual(featureFlagKeys)
    expect(
      Object.fromEntries(
        Object.entries(fixture.booleanFlags).map(([key, definition]) => [
          key,
          definition.safeDefault,
        ]),
      ),
    ).toEqual(defaultFeatureFlagDecisions)
  })
})
