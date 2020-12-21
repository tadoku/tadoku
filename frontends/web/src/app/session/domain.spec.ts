import { validateDisplayName, isAdmin } from './domain'
import { RoleBasedEntity, User } from './interfaces'

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

describe('isAdmin', () => {
  it('should disallow undefined', () => {
    const result = isAdmin(undefined)
    expect(result).toBeFalsy()
  })

  it('should disallow banned users', () => {
    const user: RoleBasedEntity = { role: 1 }
    const result = isAdmin(user)
    expect(result).toBeFalsy()
  })

  it('should disallow normal users', () => {
    const user: RoleBasedEntity = { role: 2 }
    const result = isAdmin(user)
    expect(result).toBeFalsy()
  })

  it('should allow admins', () => {
    const user: RoleBasedEntity = { role: 3 }
    const result = isAdmin(user)
    expect(result).toBeTruthy()
  })

  it('should accept a user as parameter', () => {
    const user: User = {
      id: 1,
      displayName: 'John Doe',
      email: 'foo@bar.com',
      role: 3,
    }
    const result = isAdmin(user)
    expect(result).toBeTruthy()
  })
})
