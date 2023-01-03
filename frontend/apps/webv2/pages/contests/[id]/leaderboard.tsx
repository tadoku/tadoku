import {
  useCurrentDateTime,
  useCurrentLocation,
  useInterval,
} from '@app/common/hooks'
import { useSession } from '@app/common/session'
import { useContest, useContestRegistration } from '@app/contests/api'
import Link from 'next/link'
import { useRouter } from 'next/router'
import { Breadcrumb, ButtonGroup, Flash, Pagination, Tabbar } from 'ui'
import { DateTime, Interval } from 'luxon'
import { useState } from 'react'
import { ContestOverview } from '@app/contests/ContestOverview'
import {
  ExclamationCircleIcon,
  HomeIcon,
  InformationCircleIcon,
} from '@heroicons/react/20/solid'
import { ContestConfiguration } from '@app/contests/ContestConfiguration'
import { PencilSquareIcon, PlusIcon } from '@heroicons/react/24/solid'
import { routes } from '@app/common/routes'

const data = [
  { rank: '1', user: 'powz', score: 5054.25054 },
  { rank: '2', user: 'Bijak', score: 3605.23605 },
  { rank: '3', user: 'ShockOLatte', score: 2518.72518 },
  { rank: '4', user: 'Ludie', score: 2517.32517 },
  { rank: '5', user: 'Chamsae', score: 2434.42434 },
  { rank: '6', user: 'Salome', score: 2107.12107 },
  { rank: '7', user: 'mmmm', score: 2060.1206 },
  { rank: '8', user: 'Yaku', score: 1667.21667 },
  { rank: '9', user: 'Socks', score: 1635.81635 },
  { rank: '10', user: 'clair', score: 1592.91592 },
]

const updates = [
  ['antonve', 30],
  ['sheodox', 44],
  ['Pokemod97', 32.2],
  ['Salome', 10.5],
  ['clair', 65],
  ['Yaku', 111],
  ['mmmm', 20],
  ['mmmm', 33],
  ['ShockOLatte', 1],
  ['antonve', 2],
  ['Bijak', 287],
  ['Bijak', 121],
  ['powz', 202],
  ['powz', 321],
  ['Ludie', 203],
  ['Chamsae', 140],
]

const Page = () => {
  const router = useRouter()
  const id = router.query['id']?.toString() ?? ''

  const now = useCurrentDateTime()

  const contest = useContest(id)
  const [session] = useSession()
  const currentUrl = useCurrentLocation()

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
          <div className="table-container">
            <table className="default">
              <thead>
                <tr>
                  <th className="default !text-center">Rank</th>
                  <th className="default">Nickname</th>
                  <th className="default !text-right">Score</th>
                </tr>
              </thead>
              <tbody>
                {data.map(u => (
                  <tr key={u.rank} className="link">
                    <td className="link w-10">
                      <Link
                        href={routes.contestUserProfile(id, u.user)}
                        className="reset justify-center text-lg"
                      >
                        {u.rank}
                      </Link>
                    </td>
                    <td className="link">
                      <Link
                        href={routes.contestUserProfile(id, u.user)}
                        className="reset text-lg"
                      >
                        {u.user}
                      </Link>
                    </td>
                    <td className="link">
                      <Link
                        href={routes.contestUserProfile(id, u.user)}
                        className="reset justify-end text-lg"
                      >
                        {Math.round(u.score * 10) / 10}
                      </Link>
                    </td>
                  </tr>
                ))}
                {data.length === 0 ? (
                  <tr>
                    <td
                      colSpan={3}
                      className="default h-32 font-bold text-center text-xl text-slate-400"
                    >
                      No partipants yet, be the first to sign up!
                    </td>
                  </tr>
                ) : null}
              </tbody>
            </table>
          </div>
          <div className="mt-4">
            <Pagination currentPage={1} totalPages={4} onClick={() => {}} />
          </div>
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

          <div className="card">
            <div className="-m-7 pt-4 px-4 text-sm">
              <h3 className="subtitle text-sm">Recent updates</h3>
              <ul className="divide-y-2 divide-slate-500/5 -mx-4">
                {updates.map(u => (
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
        </div>
      </div>
    </>
  )
}

export default Page
