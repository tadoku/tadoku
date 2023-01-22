import { useCurrentDateTime } from '@app/common/hooks'
import { useSession } from '@app/common/session'
import { useContest, useContestRegistration } from '@app/immersion/api'
import { useRouter } from 'next/router'
import { Breadcrumb, ButtonGroup, Tabbar } from 'ui'
import { DateTime, Interval } from 'luxon'
import { HomeIcon } from '@heroicons/react/20/solid'
import { PencilSquareIcon, PlusIcon } from '@heroicons/react/24/solid'
import { routes } from '@app/common/routes'
import { ContestLeaderboard } from '@app/immersion/ContestLeaderboard'

const Page = () => {
  const router = useRouter()
  const id = router.query['id']?.toString() ?? ''

  const now = useCurrentDateTime()

  const contest = useContest(id)
  const [session] = useSession()

  const registration = useContestRegistration(id, { enabled: !!session })

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
    DateTime.fromISO(contest.data.contest_start),
    DateTime.fromISO(contest.data.contest_end),
  )
  const hasEnded = contestInterval.isBefore(now)
  const hasStarted = contestInterval.contains(now) || hasEnded
  const isOngoing = hasStarted && !hasEnded

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
      <ContestLeaderboard
        contest={contest.data}
        id={id}
        routeForPage={page => routes.contestLeaderboard(id, page)}
      />
    </>
  )
}

export default Page
