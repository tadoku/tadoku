import { useRouter } from 'next/router'
import { Breadcrumb, ButtonGroup, Flash, Loading, Pagination, Tabbar } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { PencilSquareIcon, PlusIcon } from '@heroicons/react/24/solid'
import { routes } from '@app/common/routes'
import {
  useContest,
  useContestLogs,
  useContestRegistration,
} from '@app/immersion/api'
import { getQueryStringIntParameter } from '@app/common/router'
import Head from 'next/head'
import { useEffect, useState } from 'react'
import LogsList from '@app/immersion/LogsList'
import { useCurrentDateTime } from '@app/common/hooks'
import { DateTime, Interval } from 'luxon'
import { useSession } from '@app/common/session'

const Page = () => {
  const router = useRouter()
  const id = router.query['id']?.toString() ?? ''

  const newFilter = () => {
    return {
      page: getQueryStringIntParameter(router.query.page, 1),
      pageSize: 50,
      includeDeleted: false,
      contestId: id,
    }
  }

  const [filters, setFilters] = useState(() => newFilter())
  useEffect(() => {
    setFilters(newFilter())
  }, [router.asPath])

  const now = useCurrentDateTime()

  const contest = useContest(id)
  const logs = useContestLogs(filters)

  const [session] = useSession()
  const registration = useContestRegistration(id, { enabled: !!session })

  if (contest.isLoading || contest.isIdle) {
    return <Loading />
  }

  if (contest.isError) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  const contestInterval = Interval.fromDateTimes(
    DateTime.fromISO(contest.data.contest_start),
    DateTime.fromISO(contest.data.contest_end).endOf('day'),
  )
  const hasEnded = contestInterval.isBefore(now)
  const hasStarted = contestInterval.contains(now) || hasEnded
  const isOngoing = hasStarted && !hasEnded

  const logsTotalPages = logs.data
    ? Math.ceil(logs.data.total_size / filters.pageSize)
    : 0

  return (
    <>
      <Head>
        <title>Contest updates - Tadoku</title>
      </Head>
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
              label: contest.data.title,
              href: routes.contestLeaderboard(id),
            },
          ]}
        />
      </div>
      <div className="h-stack justify-between items-center w-full">
        <div>
          <h1 className="title">Contest</h1>
          <h2 className="subtitle">{contest.data.title}</h2>
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
                href: routes.logCreate(),
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
            active: false,
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
            active: true,
            href: routes.contestUpdates(id),
            label: 'Updates',
            disabled: false,
          },
        ]}
      />

      <Flash style="info" className="mt-4">
        This page does not show the username yet, it will be added soon.
      </Flash>
      <div className="card p-0 mt-4">
        <LogsList logs={logs} />
      </div>

      {logsTotalPages > 1 ? (
        <div className="mt-4">
          <Pagination
            currentPage={filters.page}
            totalPages={logsTotalPages}
            getHref={page => routes.contestUpdates(id, page)}
          />
        </div>
      ) : null}
    </>
  )
}

export default Page
