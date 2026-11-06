import type { LastLogged } from './log-draft'

// Webv2's keys, so remembered choices survive the host switch.
const languageKey = 'log-form-v2:last-logged-language'
const activityKey = 'log-form-v2:last-logged-activity'

/** The browser's remembered choices; nothing when storage is unavailable. Read it only after hydration. */
export function readLastLogged(): LastLogged {
  try {
    return { languageCode: localStorage.getItem(languageKey), activityId: localStorage.getItem(activityKey) }
  } catch {
    return { languageCode: null, activityId: null }
  }
}

/** Remembers a saved log's language and activity. Unavailable storage is ignored. */
export function rememberLastLogged(languageCode: string, activityId: number) {
  try {
    localStorage.setItem(languageKey, languageCode)
    localStorage.setItem(activityKey, String(activityId))
  } catch {
    // The form works without remembered choices.
  }
}
