import { validateDisplayName } from './domain'

describe('validateDisplayName', () => {
  it('should allow a correct value', () => {
    const valid = validateDisplayName('foo-bar_123')
    expect(valid).toBeTruthy()
  })

  it('should allow unicode', () => {
    const valid = validateDisplayName('神様')
    expect(valid).toBeTruthy()
  })

  it('should allow spaces', () => {
    const valid = validateDisplayName('foo bar')
    expect(valid).toBeTruthy()
  })

  it('should disallow 1 character names', () => {
    const valid = validateDisplayName('a')
    expect(valid).toBeFalsy()
  })

  it('should disallow names that are too long', () => {
    const valid = validateDisplayName('abcdefghijklmnopqwrstuvw')
    expect(valid).toBeFalsy()
  })

  it('should disallow names with strange characters', () => {
    const valid = validateDisplayName("Robert'); DROP TABLE students;--")
    expect(valid).toBeFalsy()
  })
})
