import { useCurrentDateTime, useCurrentLocation } from '@app/common/hooks'
import { useSession } from '@app/common/session'
import {
  useContest,
  useContestLeaderboard,
  useContestRegistration,
} from '@app/contests/api'
import Link from 'next/link'
import { useRouter } from 'next/router'
import { Breadcrumb, ButtonGroup, Flash, Pagination, Tabbar } from 'ui'
import { Interval } from 'luxon'
import { useEffect, useState } from 'react'
import { ContestOverview } from '@app/contests/ContestOverview'
import {
  ExclamationCircleIcon,
  HomeIcon,
  InformationCircleIcon,
} from '@heroicons/react/20/solid'
import { ContestConfiguration } from '@app/contests/ContestConfiguration'
import { PencilSquareIcon, PlusIcon } from '@heroicons/react/24/solid'
import { routes } from '@app/common/routes'
import { getQueryStringIntParameter } from '@app/common/router'
import { Leaderboard } from '@app/contests/Leaderboard'

const Page = () => {
  const router = useRouter()
  const id = router.query['id']?.toString() ?? ''

  const now = useCurrentDateTime()

  const contest = useContest(id)
  const [session] = useSession()
  const currentUrl = useCurrentLocation()

  const registration = useContestRegistration(id, { enabled: !!session })

  const newFilter = () => {
    return {
      contestId: id,
      page: getQueryStringIntParameter(router.query.page, 1),
      pageSize: 50,
    }
  }

  const [filters, setFilters] = useState(() => newFilter())
  const leaderboard = useContestLeaderboard(filters)

  useEffect(() => {
    setFilters(newFilter())
  }, [router.asPath])

  if (contest.isLoading || contest.isIdle) {
    return <p>Loading...</p>
  }

  if (contest.isError) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  const contestInterval = Interval.fromDateTimes(
    contest.data.contestStart,
    contest.data.contestEnd,
  )
  const hasEnded = contestInterval.isBefore(now)
  const hasStarted = contestInterval.contains(now) || hasEnded
  const isOngoing = hasStarted && !hasEnded
  const registrationClosed = Interval.fromDateTimes(
    contest.data.contestStart,
    contest.data.registrationEnd,
  ).isBefore(now)

  const showPagination =
    leaderboard.data && leaderboard.data.totalSize > filters.pageSize

  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: contest.data.official
                ? 'Official contests'
                : 'User contests',
              href: contest.data.official
                ? routes.contestListOfficial()
                : routes.contestListUserContests(),
            },
            {
              label: contest.data.description,
              href: routes.contestLeaderboard(id),
            },
          ]}
        />
      </div>
      <div className="h-stack justify-between items-center w-full">
        <div>
          <h1 className="title">Contest</h1>
          <h2 className="subtitle">{contest.data.description}</h2>
        </div>
        <div>
          <ButtonGroup
            actions={[
              {
                href: routes.contestJoin(id),
                label: 'Join contest',
                IconComponent: PlusIcon,
                style: 'primary',
                visible: !hasEnded && registration.data === undefined,
              },
              {
                href: routes.contestLog(),
                label: 'Log update',
                IconComponent: PencilSquareIcon,
                style: 'secondary',
                visible: isOngoing && registration.data !== undefined,
              },
            ]}
            orientation="right"
          />
        </div>
      </div>
      <Tabbar
        links={[
          {
            active: true,
            href: routes.contestLeaderboard(id),
            label: 'Leaderboard',
          },
          {
            active: false,
            href: routes.contestLeaderboard(id),
            label: 'Statistics',
            disabled: true,
          },
          {
            active: false,
            href: routes.contestLeaderboard(id),
            label: 'Logs',
            disabled: true,
          },
        ]}
      />
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
                  leaderboard.data.totalSize / filters.pageSize,
                )}
                getHref={page => routes.contestLeaderboard(id, page)}
              />
            </div>
          ) : null}
        </div>
        <div className="w-[25%] space-y-4">
          <ContestOverview
            contest={contest.data}
            hasStarted={hasStarted}
            hasEnded={hasEnded}
            registrationClosed={registrationClosed}
            now={now}
          />

          <ContestConfiguration contest={contest.data} />

          <div className="card">
            <div className="-m-7 py-4 px-4 text-sm">
              <h3 className="subtitle text-sm mb-2">Contest summary</h3>
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

export default Page
