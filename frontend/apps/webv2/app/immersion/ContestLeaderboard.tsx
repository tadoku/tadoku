import { useCurrentDateTime, useCurrentLocation } from '@app/common/hooks'
import { useSession } from '@app/common/session'
import {
  useContestLeaderboard,
  ContestView,
  useContestSummary,
  useContestLogs,
} from '@app/immersion/api'
import Link from 'next/link'
import { useRouter } from 'next/router'
import { Flash, Pagination } from 'ui'
import { DateTime, Interval } from 'luxon'
import { useEffect, useState } from 'react'
import { ContestOverview } from '@app/immersion/ContestOverview'
import {
  ExclamationCircleIcon,
  InformationCircleIcon,
} from '@heroicons/react/20/solid'
import { ContestConfiguration } from '@app/immersion/ContestConfiguration'
import { routes } from '@app/common/routes'
import { getQueryStringIntParameter } from '@app/common/router'
import { Leaderboard } from '@app/immersion/Leaderboard'
import { formatScore } from '@app/common/format'

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
  const summary = useContestSummary(id)

  useEffect(() => {
    setFilters({
      contestId: id,
      page: getQueryStringIntParameter(router.query.page, 1),
      pageSize: 50,
    })
  }, [router.asPath, id, router.query.page])

  const contestInterval = Interval.fromDateTimes(
    DateTime.fromISO(contest.contest_start),
    DateTime.fromISO(contest.contest_end).endOf('day'),
  )
  const hasEnded = contestInterval.isBefore(now)
  const hasStarted = contestInterval.contains(now) || hasEnded
  const registrationClosed = Interval.fromDateTimes(
    DateTime.fromISO(contest.contest_start),
    DateTime.fromISO(contest.registration_end).endOf('day'),
  ).isBefore(now)

  const showPagination =
    leaderboard.data && leaderboard.data.total_size > filters.pageSize

  const logs = useContestLogs(
    {
      contestId: id,
      includeDeleted: false,
      page: 1,
      pageSize: 10,
    },
    { enabled: hasStarted && !hasEnded },
  )

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

      <div className="v-stack md:h-stack mt-4 space-y-4 md:space-y-0 md:space-x-4">
        <div className="flex-grow">
          <Leaderboard
            leaderboard={leaderboard}
            urlForRow={userId => routes.contestUserProfile(id, userId)}
          />
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
          <p className="text-sm text-slate-500 mt-4 text-center italic">
            Scores may take a few seconds to update.
          </p>
        </div>
        <div className="md:w-[300px] space-y-4">
          <ContestOverview
            contest={contest}
            hasStarted={hasStarted}
            hasEnded={hasEnded}
            registrationClosed={registrationClosed}
            now={now}
          />

          {contest.description ? (
            <div className="card narrow text-sm">
              <h3 className="subtitle mb-2">Description</h3>
              {contest.description}
            </div>
          ) : null}

          <ContestConfiguration contest={contest} />

          {summary.data ? (
            <div className="card narrow text-sm">
              <h3 className="subtitle text-sm mb-2">Summary</h3>
              <strong>{summary.data.participant_count}</strong> participant
              {summary.data.participant_count != 1 ? 's' : ''} immersing in{' '}
              <strong>{summary.data.language_count}</strong> language
              {summary.data.language_count != 1 ? 's' : ''} for a total score of{' '}
              <strong>{formatScore(summary.data.total_score)}</strong>.
            </div>
          ) : null}

          {logs.data?.logs.length ?? 0 > 0 ? (
            <div className="card narrow text-sm">
              <h3 className="subtitle">Recent updates</h3>
              <ul className="divide-y-2 divide-slate-500/5 -mx-4">
                {logs.data!.logs.map(log => (
                  <li key={log.id}>
                    <Link
                      href={routes.log(log.id)}
                      className="reset px-4 py-2 flex justify-between items-center hover:bg-slate-500/5"
                    >
                      <span className="font-bold text-base">
                        {log.user_display_name ?? 'Unknown user'}
                      </span>
                      <span className="font-bold text-lime-700 text-lg">
                        + {formatScore(log.score)}
                      </span>
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ) : null}
        </div>
      </div>
    </>
  )
}
