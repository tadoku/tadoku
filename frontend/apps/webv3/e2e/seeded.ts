import { createHash } from 'node:crypto'
import { branch } from './logging-fixtures'

/** The ID `.dev/seed.sql` derives for a branch tenant's fixture: a version 5 UUID of the tenant key in `namespace`. */
function seededId(namespace: string) {
  const bytes = createHash('sha1')
    .update(Buffer.from(namespace.replaceAll('-', ''), 'hex'))
    .update(`tadoku/${branch}`)
    .digest()
    .subarray(0, 16)
  bytes[6] = (bytes[6] & 0x0f) | 0x50
  bytes[8] = (bytes[8] & 0x3f) | 0x80
  const hex = bytes.toString('hex')
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`
}

/** The branch tenant's Dev Finished Round, which ended a week before seeding and holds one Dev Reader log. */
export const finishedRound = () => ({
  contestId: seededId('00000000-0000-4000-8000-000000000104'),
  logId: seededId('00000000-0000-4000-8000-000000000304'),
  score: 10,
})

/**
 * The branch tenant's Dev Scoring Round, which started on seeding day and scores with its own override rule set:
 * reading earns three points a page, doubled by the `webv3-bonus` tag, and every other input falls back to the pinned
 * platform version.
 */
export const scoringRound = () => ({
  contestId: seededId('00000000-0000-4000-8000-000000000105'),
  ruleSetId: seededId('00000000-0000-4000-8000-000000000501'),
  readingRate: 3,
  bonusTag: 'webv3-bonus',
  bonusRate: 2,
})
