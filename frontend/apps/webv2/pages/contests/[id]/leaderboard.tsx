import { useCurrentLocation, useInterval } from '@app/common/hooks'
import { useSession } from '@app/common/session'
import { useContest } from '@app/contests/api'
import { CheckBadgeIcon } from '@heroicons/react/20/solid'
import { ArrowRightIcon } from '@heroicons/react/24/outline'
import Link from 'next/link'
import { useRouter } from 'next/router'
import { Pagination, Tabbar } from 'ui'
import getConfig from 'next/config'
import { DateTime, Duration, Interval } from 'luxon'
import { useState } from 'react'

const { publicRuntimeConfig } = getConfig()

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

  const [now, setNow] = useState(() => DateTime.utc())
  useInterval(() => setNow(DateTime.utc()), 1000)

  const contest = useContest(id)
  const [session] = useSession()
  const currentUrl = useCurrentLocation()

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

  return (
    <>
      <div className="h-stack justify-between items-center w-full">
        <div>
          <h1 className="title">Contest</h1>
          <h2 className="subtitle">{contest.data.description}</h2>
        </div>
        <div>
          {true ? (
            <a href="#" className="btn primary">
              Join contest
            </a>
          ) : null}
          {false ? (
            <a href="#" className="btn secondary">
              Submit log
            </a>
          ) : null}
        </div>
      </div>
      <Tabbar
        links={[
          {
            active: true,
            href: `/contests/${id}/leaderboard`,
            label: 'Leaderboard',
          },
          {
            active: false,
            href: `/contests/${id}/statistics`,
            label: 'Statistics',
            disabled: true,
          },
          {
            active: false,
            href: `/contests/${id}/logs`,
            label: 'Logs',
            disabled: true,
          },
        ]}
      />

      {!session && !hasEnded ? (
        <Link
          href={
            publicRuntimeConfig.authUiUrl + `/login?return_to=${currentUrl}`
          }
          className="reset block flash warning mt-4 hover:bg-amber-300 transition-all ease-in"
        >
          You need to log in to participate in this contest.
        </Link>
      ) : null}
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
                        href={`/contests/${id}/user/${u.user}`}
                        className="reset justify-center text-lg"
                      >
                        {u.rank}
                      </Link>
                    </td>
                    <td className="link">
                      <Link
                        href={`/contests/${id}/user/${u.user}`}
                        className="reset text-lg"
                      >
                        {u.user}
                      </Link>
                    </td>
                    <td className="link">
                      <Link
                        href={`/contests/${id}/user/${u.user}`}
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
          <div className="card text-sm">
            <div className="-mx-7 -mt-7 py-4 px-4 h-stack justify-between items-center">
              <div>
                <div className="subtitle text-sm">Started</div>
                <div className="font-bold">January 1, 2022</div>
              </div>
              <ArrowRightIcon className="w-7 h-7 text-slate-500/30" />
              <div>
                <div className="subtitle text-sm">Ending</div>
                <div className="font-bold">January 31, 2022</div>
              </div>
            </div>
            <div className="-mx-7 px-4 py-2 border-t-2 border-slate-500/5">
              {!hasStarted ? (
                <>
                  Starting in{' '}
                  <strong>
                    {contest.data.contestEnd
                      .plus(Duration.fromObject({ days: 1 }))
                      .diffNow(['days', 'hours', 'minute'])
                      .toHuman({
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
                    {contest.data.contestEnd
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
              <strong>
                {now.toLocaleString(DateTime.DATETIME_MED_WITH_SECONDS)}
              </strong>
            </div>
            <div className="-mx-7 -mb-7 px-4 py-2 bg-slate-500/5 flex items-center space-x-1">
              <span>Administered by</span>
              {contest.data.official ? (
                <div className="flex items-center">
                  <strong>Tadoku</strong>
                  <CheckBadgeIcon className="ml-1 w-4 h-4 text-lime-700" />
                </div>
              ) : (
                <Link
                  href={`/profile/${contest.data.ownerUserId}`}
                  className="font-bold"
                >
                  {contest.data.ownerUserDisplayName}
                </Link>
              )}
            </div>
          </div>

          <div className="card">
            <div className="-m-7 py-4 px-4 text-sm">
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
