/**
 * Each immersion activity's colour and lower-case name, keyed by the API's activity ID. Activities the map doesn't know
 * read as muted and can't be filtered by name. Class names stay literal so Tailwind generates them.
 */
const tones = {
  1: { slug: 'reading', text: 'text-activity-reading', fill: 'bg-activity-reading' },
  2: { slug: 'listening', text: 'text-activity-listening', fill: 'bg-activity-listening' },
  3: { slug: 'writing', text: 'text-activity-writing', fill: 'bg-activity-writing' },
  4: { slug: 'speaking', text: 'text-activity-speaking', fill: 'bg-activity-speaking' },
  5: { slug: 'study', text: 'text-activity-study', fill: 'bg-activity-study' },
} as const
const other = { slug: undefined, text: 'text-activity-other', fill: 'bg-activity-other' }

/** An activity's name in addresses, such as `listening` in `?activity=listening`. */
export type ActivitySlug = (typeof tones)[keyof typeof tones]['slug']

const known = (id: number): id is keyof typeof tones => id in tones

export const activityTone = (id: number) => (known(id) ? tones[id] : other)

/** The activity's colour as a CSS value, for charts. */
export const activityColor = (id: number) => `var(--color-activity-${activityTone(id).slug ?? 'other'})`

export const activitySlug = (id: number): ActivitySlug | undefined => activityTone(id).slug

export const activityId = (slug: string) => {
  const entry = Object.entries(tones).find(([, tone]) => tone.slug === slug)
  return entry ? Number(entry[0]) : undefined
}
