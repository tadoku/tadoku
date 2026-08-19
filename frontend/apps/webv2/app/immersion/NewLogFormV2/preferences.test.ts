import { describe, expect, it } from 'vitest'
import {
  getLastLoggedActivityId,
  storeLastLoggedPreferences,
  StorageLike,
} from './preferences'

const createStorage = (initialValues: Record<string, string> = {}) => {
  const values = new Map(Object.entries(initialValues))
  const storage: StorageLike = {
    getItem: key => values.get(key) ?? null,
    setItem: (key, value) => values.set(key, value),
  }

  return { storage, values }
}

describe('last logged activity preference', () => {
  it('restores a stored activity that is still available', () => {
    const { storage } = createStorage({
      'log-form-v2:last-logged-activity': '2',
    })

    expect(getLastLoggedActivityId([{ id: 1 }, { id: 2 }], storage)).toBe(2)
  })

  it('ignores a stored activity that is no longer available', () => {
    const { storage } = createStorage({
      'log-form-v2:last-logged-activity': '3',
    })

    expect(getLastLoggedActivityId([{ id: 1 }, { id: 2 }], storage)).toBeNull()
  })

  it('stores the language and activity from the successful log', () => {
    const { storage, values } = createStorage()

    storeLastLoggedPreferences('jpn', 2, storage)

    expect(values.get('log-form-v2:last-logged-language')).toBe('jpn')
    expect(values.get('log-form-v2:last-logged-activity')).toBe('2')
  })
})
