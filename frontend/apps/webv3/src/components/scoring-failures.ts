import type { LogFailure } from '../lib/logs'

const messages: Partial<Record<LogFailure, string>> = {
  'account-unavailable': 'Account service unavailable. Try again when the service recovers.',
  switched: 'You are now signed in as a different account. Reload the page to manage scoring as that account.',
  invalid: 'The API refused the rules. Check each rule’s unit, language and tag, then try again.',
  'signed-out': 'Your session ended. Log in again, then retry.',
  forbidden: 'Your account cannot change this contest’s scoring.',
  // The API answers 409 once the contest has started.
  deleting: 'Scoring is locked because this contest has started.',
  'not-found': 'This contest or rule-set version no longer exists.',
  unavailable: 'We could not confirm the change. Reload the page to see the current versions before retrying.',
}

/** Why a scoring change failed, after what did not happen unless the change may have been made. */
export const scoringFailures = (reason: LogFailure, outcome: string) =>
  reason === 'unavailable' || !messages[reason] ? messages.unavailable! : `${outcome} ${messages[reason]}`
