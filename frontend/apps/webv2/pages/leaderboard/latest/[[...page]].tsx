import { useCurrentDateTime } from '@app/common/hooks'
import { useSession } from '@app/common/session'
import {
  useContestRegistration,
  useLatestOfficialContest,
} from '@app/immersion/api'
import { Breadcrumb, ButtonGroup, Loading, Tabbar } from 'ui'
import { DateTime, Interval } from 'luxon'
import { HomeIcon } from '@heroicons/react/20/solid'
import { PencilSquareIcon, PlusIcon } from '@heroicons/react/24/solid'
import { routes } from '@app/common/routes'
import { ContestLeaderboard } from '@app/immersion/ContestLeaderboard'
import Head from 'next/head'

const Page = () => {
  const now = useCurrentDateTime()

  const [session] = useSession()
  const contest = useLatestOfficialContest()
  const id = contest.data?.id ?? ''
  const registration = useContestRegistration(id, {
    enabled: !!session && !!contest.data,
  })

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

  return (
    <>
      <Head>
        <title>Latest leaderboard - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: 'Leaderboard',
              href: routes.leaderboardLatestOfficial(),
            },
          ]}
        />
      </div>
      <div className="h-stack justify-between items-center w-full">
        <div>
          <h1 className="title">Leaderboard</h1>
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
            active: true,
            href: routes.leaderboardLatestOfficial(),
            label: 'Latest',
          },
          {
            active: false,
            href: routes.leaderboardYearlyOfficial(),
            label: 'Yearly',
            disabled: false,
          },
          {
            active: false,
            href: routes.leaderboardAllTimeOfficial(),
            label: 'All time',
            disabled: false,
          },
        ]}
      />
      <ContestLeaderboard
        contest={contest.data}
        id={id}
        routeForPage={page => routes.contestLeaderboard(id, page)}
      />
    </>
  )
}

export default Page
