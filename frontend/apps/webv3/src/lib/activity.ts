/**
 * Each immersion activity's colour, keyed by the API's activity ID. Activities the map doesn't know read as muted.
 * Class names stay literal so Tailwind generates them.
 */
const tones: Record<number, { text: string; fill: string }> = {
  1: { text: 'text-activity-reading', fill: 'bg-activity-reading' },
  2: { text: 'text-activity-listening', fill: 'bg-activity-listening' },
  3: { text: 'text-activity-writing', fill: 'bg-activity-writing' },
  4: { text: 'text-activity-speaking', fill: 'bg-activity-speaking' },
  5: { text: 'text-activity-study', fill: 'bg-activity-study' },
}
const other = { text: 'text-activity-other', fill: 'bg-activity-other' }

export const activityTone = (id: number) => tones[id] ?? other
