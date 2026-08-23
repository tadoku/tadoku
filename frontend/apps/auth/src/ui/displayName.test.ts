import { readFileSync } from 'node:fs'
import Ajv from 'ajv'
import { parse } from 'yaml'
import { describe, expect, it } from 'vitest'
import {
  DISPLAY_NAME_MAX_LENGTH,
  inputMaxLength,
} from './displayName'

const kratosValuesPath = new URL(
  '../../../../../infra/dev/ory/kratos_values.yaml',
  import.meta.url,
)

const renderedValues = readFileSync(kratosValuesPath, 'utf8').replace(
  /\{\{[^}]+\}\}/g,
  'template-value',
)
const values = parse(renderedValues)
const identitySchema = JSON.parse(
  values.kratos.identitySchemas['identity.default.schema.json'],
)
const validateIdentity = new Ajv().compile(identitySchema)

const identityWithDisplayName = (displayName: string) => ({
  traits: {
    email: 'reader@example.com',
    display_name: displayName,
  },
})

describe('display-name length limit', () => {
  it.each([
    ['ASCII', 'a'.repeat(32)],
    ['Japanese', '読'.repeat(32)],
    ['astral Unicode', '📚'.repeat(32)],
  ])('accepts 32 %s characters in the Kratos schema', (_, displayName) => {
    expect(validateIdentity(identityWithDisplayName(displayName))).toBe(true)
  })

  it.each([
    ['ASCII', 'a'.repeat(33)],
    ['Japanese', '読'.repeat(33)],
    ['astral Unicode', '📚'.repeat(33)],
  ])('rejects 33 %s characters in the Kratos schema', (_, displayName) => {
    expect(validateIdentity(identityWithDisplayName(displayName))).toBe(false)
  })

  it('exposes the limit on the shared registration and settings input', () => {
    expect(inputMaxLength('traits.display_name')).toBe(DISPLAY_NAME_MAX_LENGTH)
    expect(inputMaxLength('traits.email')).toBeUndefined()
  })
})
