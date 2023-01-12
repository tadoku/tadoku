import { useCurrentDateTime, useCurrentLocation } from '@app/common/hooks'
import { useSession } from '@app/common/session'
import { useContestLeaderboard, ContestView } from '@app/contests/api'
import Link from 'next/link'
import { useRouter } from 'next/router'
import { Flash, Pagination } from 'ui'
import { DateTime, Interval } from 'luxon'
import { useEffect, useState } from 'react'
import { ContestOverview } from '@app/contests/ContestOverview'
import {
  ExclamationCircleIcon,
  InformationCircleIcon,
} from '@heroicons/react/20/solid'
import { ContestConfiguration } from '@app/contests/ContestConfiguration'
import { routes } from '@app/common/routes'
import { getQueryStringIntParameter } from '@app/common/router'
import { Leaderboard } from '@app/contests/Leaderboard'

interface Props {
  id: string
  contest: ContestView
  routeForPage: (page: number) => string
}

export const ContestLeaderboard = ({ id, contest, routeForPage }: Props) => {
  const newFilter = () => {
    return {
      contestId: id,
      page: getQueryStringIntParameter(router.query.page, 1),
      pageSize: 50,
    }
  }

  const router = useRouter()
  const now = useCurrentDateTime()
  const [session] = useSession()
  const currentUrl = useCurrentLocation()
  const [filters, setFilters] = useState(() => newFilter())
  const leaderboard = useContestLeaderboard(filters)

  useEffect(() => {
    setFilters(newFilter())
  }, [router.asPath])

  const contestInterval = Interval.fromDateTimes(
    DateTime.fromISO(contest.contest_start),
    DateTime.fromISO(contest.contest_end),
  )
  const hasEnded = contestInterval.isBefore(now)
  const hasStarted = contestInterval.contains(now) || hasEnded
  const registrationClosed = Interval.fromDateTimes(
    DateTime.fromISO(contest.contest_start),
    DateTime.fromISO(contest.registration_end),
  ).isBefore(now)

  const showPagination =
    leaderboard.data && leaderboard.data.total_size > filters.pageSize

  return (
    <>
      <Flash
        style="info"
        href={routes.authLogin(currentUrl)}
        IconComponent={InformationCircleIcon}
        className="mt-4"
        visible={!session && !hasEnded}
      >
        You need to log in to participate in this contest.
      </Flash>
      <Flash
        style="warning"
        IconComponent={ExclamationCircleIcon}
        className="mt-4"
        visible={hasEnded}
      >
        This contest has already ended and does not accept any new participants.
      </Flash>

      <div className="flex mt-4 space-x-4">
        <div className="flex-grow">
          <Leaderboard contestId={id} leaderboard={leaderboard} />
          {showPagination ? (
            <div className="mt-4">
              <Pagination
                currentPage={filters.page}
                totalPages={Math.ceil(
                  leaderboard.data.total_size / filters.pageSize,
                )}
                getHref={page => routeForPage(page)}
              />
            </div>
          ) : null}
        </div>
        <div className="w-[25%] space-y-4">
          <ContestOverview
            contest={contest}
            hasStarted={hasStarted}
            hasEnded={hasEnded}
            registrationClosed={registrationClosed}
            now={now}
          />

          {contest.description ? (
            <div className="card">
              <div className="-m-7 py-4 px-4 text-sm">
                <h3 className="subtitle text-sm mb-2">Description</h3>
                {contest.description}
              </div>
            </div>
          ) : null}

          <ContestConfiguration contest={contest} />

          <div className="card">
            <div className="-m-7 py-4 px-4 text-sm">
              <h3 className="subtitle text-sm mb-2">Summary</h3>
              <strong>100</strong> participants immersing in <strong>12</strong>{' '}
              languages for a total score of <strong>9000</strong>.
            </div>
          </div>

          {hasStarted && !hasEnded && false ? (
            <div className="card">
              <div className="-m-7 pt-4 px-4 text-sm">
                <h3 className="subtitle text-sm">Recent updates</h3>
                <ul className="divide-y-2 divide-slate-500/5 -mx-4">
                  {[].map(u => (
                    <li key={`${u[0]}-${u[1]}`}>
                      <Link
                        href="#"
                        className="reset px-4 py-2 flex justify-between items-center hover:bg-slate-500/5"
                      >
                        <span className="font-bold text-base">{u[0]}</span>
                        <span className="font-bold text-lime-700 text-lg">
                          +{u[1]}
                        </span>
                      </Link>
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          ) : null}
        </div>
      </div>
    </>
  )
}
