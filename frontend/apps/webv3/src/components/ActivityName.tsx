import { activityTone } from '../lib/activity'

/** An activity's name behind a dot in its colour, kept on one line. */
export function ActivityName({ activity }: { activity: { id: number; name: string } }) {
  return (
    <span className="inline-flex items-baseline gap-1.5 whitespace-nowrap">
      <span className={`inline-block size-2 self-center rounded-full bg-current ${activityTone(activity.id).text}`} aria-hidden="true" />
      {activity.name}
    </span>
  )
}
