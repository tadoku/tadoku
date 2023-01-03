import { ContestView } from '@app/contests/api'
import { CheckBadgeIcon } from '@heroicons/react/20/solid'
import { ArrowRightIcon } from '@heroicons/react/24/outline'
import Link from 'next/link'
import { DateTime, Duration } from 'luxon'

interface Props {
  contest: ContestView
  hasStarted: boolean
  hasEnded: boolean
  now: DateTime
}

export const ContestOverview = ({
  contest,
  hasStarted,
  hasEnded,
  now,
}: Props) => (
  <div className="card text-sm">
    <div className="-mx-7 -mt-7 py-4 px-4 h-stack justify-between items-center">
      <div>
        <div className="subtitle text-sm">
          {hasStarted ? 'Started' : 'Starting'}
        </div>
        <div className="font-bold">
          {contest.contestStart.toLocaleString(DateTime.DATE_MED)}
        </div>
      </div>
      <ArrowRightIcon className="w-7 h-7 text-slate-500/30" />
      <div>
        <div className="subtitle text-sm">{hasEnded ? 'Ended' : 'Ending'}</div>
        <div className="font-bold">
          {contest.contestEnd.toLocaleString(DateTime.DATE_MED)}
        </div>
      </div>
    </div>
    <div className="-mx-7 px-4 py-2 border-t-2 border-slate-500/5">
      {!hasStarted ? (
        <>
          Starting in{' '}
          <strong>
            {contest.contestStart.diffNow(['days', 'hours', 'minute']).toHuman({
              maximumFractionDigits: 0,
              unitDisplay: 'short',
            })}
          </strong>
        </>
      ) : hasEnded ? (
        <>
          Contest has <strong>ended</strong>
        </>
      ) : (
        <>
          Ending in{' '}
          <strong>
            {contest.contestEnd
              .plus(Duration.fromObject({ days: 1 }))
              .diffNow(['days', 'hours', 'minute'])
              .toHuman({
                maximumFractionDigits: 0,
                unitDisplay: 'short',
              })}
          </strong>
        </>
      )}
    </div>
    <div className="-mx-7 px-4 py-2 border-t-2 border-slate-500/5">
      Tadoku time:{' '}
      <strong>{now.toLocaleString(DateTime.DATETIME_MED_WITH_SECONDS)}</strong>
    </div>
    <div className="-mx-7 -mb-7 px-4 py-2 bg-slate-500/5 flex items-center space-x-1">
      <span>Administered by</span>
      {contest.official ? (
        <div className="flex items-center">
          <strong>Tadoku</strong>
          <CheckBadgeIcon className="ml-1 w-4 h-4 text-lime-700" />
        </div>
      ) : (
        <Link href={`/profile/${contest.ownerUserId}`} className="font-bold">
          {contest.ownerUserDisplayName}
        </Link>
      )}
    </div>
  </div>
)
