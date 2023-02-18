import { ContestView } from '@app/immersion/api'
import { CheckBadgeIcon } from '@heroicons/react/20/solid'
import { ArrowRightIcon } from '@heroicons/react/24/outline'
import Link from 'next/link'
import { DateTime } from 'luxon'
import { routes } from '@app/common/routes'

interface Props {
  contest: ContestView
  hasStarted: boolean
  hasEnded: boolean
  registrationClosed: boolean
  now: DateTime
}

export const ContestOverview = ({
  contest,
  hasStarted,
  hasEnded,
  registrationClosed,
  now,
}: Props) => (
  <div className="card narrow text-sm">
    <div className="h-stack justify-between items-center max-w-[300px]">
      <div>
        <div className="subtitle text-sm">
          {hasStarted ? 'Started' : 'Starting'}
        </div>
        <div className="font-bold">
          {DateTime.fromISO(contest.contest_start).toLocaleString(
            DateTime.DATE_MED,
          )}
        </div>
      </div>
      <ArrowRightIcon className="w-7 h-7 text-slate-500/30" />
      <div>
        <div className="subtitle text-sm">{hasEnded ? 'Ended' : 'Ending'}</div>
        <div className="font-bold">
          {DateTime.fromISO(contest.contest_end).toLocaleString(
            DateTime.DATE_MED,
          )}
        </div>
      </div>
    </div>
    <div className="py-2 flex flex-col space-y-2">
      {!hasStarted ? (
        <div>
          Starting in{' '}
          <strong>
            {DateTime.fromISO(contest.contest_start)
              .diffNow(['days', 'hours', 'minute'])
              .toHuman({
                maximumFractionDigits: 0,
                unitDisplay: 'short',
              })}
          </strong>
        </div>
      ) : hasEnded ? (
        <div>
          Contest has <strong>ended</strong>
        </div>
      ) : (
        <>
          <div>
            Ending in{' '}
            <strong>
              {DateTime.fromISO(contest.contest_end)
                .endOf('day')
                .diffNow(['days', 'hours', 'minute'])
                .toHuman({
                  maximumFractionDigits: 0,
                  unitDisplay: 'short',
                })}
            </strong>
          </div>

          {registrationClosed ? (
            <div>No longer accepting new participants</div>
          ) : (
            <div>
              Registrations open until{' '}
              <strong>
                {DateTime.fromISO(contest.registration_end).toLocaleString(
                  DateTime.DATE_MED,
                )}
              </strong>
            </div>
          )}
        </>
      )}
      <div>
        Tadoku time:{' '}
        <strong>
          {now.toLocaleString(DateTime.DATETIME_MED_WITH_SECONDS)}
        </strong>
      </div>
    </div>
    <div className="-mx-4 -mb-4 px-4 py-2 bg-slate-500/5 flex items-center space-x-1">
      <span>Administered by</span>
      {contest.official ? (
        <div className="flex items-center">
          <strong>Tadoku</strong>
          <CheckBadgeIcon className="ml-1 w-4 h-4 text-lime-700" />
        </div>
      ) : (
        <Link
          href={routes.userProfileStatistics(contest.owner_user_id!)}
          className="font-bold"
        >
          {contest.owner_user_display_name}
        </Link>
      )}
    </div>
  </div>
)
