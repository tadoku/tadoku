const LAST_LOGGED_LANGUAGE_STORAGE_KEY = 'log-form-v2:last-logged-language'
const LAST_LOGGED_ACTIVITY_STORAGE_KEY = 'log-form-v2:last-logged-activity'

export type StorageLike = Pick<Storage, 'getItem' | 'setItem'>

const getLocalStorage = (): StorageLike | undefined => {
  if (typeof window === 'undefined') {
    return undefined
  }

  try {
    return window.localStorage
  } catch {
    return undefined
  }
}

export const getLastLoggedLanguage = (
  storage: StorageLike | undefined = getLocalStorage(),
) => {
  try {
    return storage?.getItem(LAST_LOGGED_LANGUAGE_STORAGE_KEY) ?? null
  } catch {
    return null
  }
}

export const getLastLoggedActivityId = (
  activities: ReadonlyArray<{ id: number }>,
  storage: StorageLike | undefined = getLocalStorage(),
) => {
  let storedActivityId: string | null
  try {
    storedActivityId =
      storage?.getItem(LAST_LOGGED_ACTIVITY_STORAGE_KEY) ?? null
  } catch {
    return null
  }

  return (
    activities.find(activity => activity.id.toString() === storedActivityId)
      ?.id ?? null
  )
}

export const storeLastLoggedPreferences = (
  languageCode: string,
  activityId: number,
  storage: StorageLike | undefined = getLocalStorage(),
) => {
  if (storage === undefined) {
    return
  }

  try {
    storage.setItem(LAST_LOGGED_LANGUAGE_STORAGE_KEY, languageCode)
  } catch {}

  try {
    storage.setItem(LAST_LOGGED_ACTIVITY_STORAGE_KEY, activityId.toString())
  } catch {}
}
