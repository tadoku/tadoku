import { UseQueryResult } from 'react-query'
import { Leaderboard as LeaderboardType } from '@app/contests/api'
import { Flash } from 'ui'
import { ExclamationCircleIcon } from '@heroicons/react/20/solid'
import Link from 'next/link'
import { routes } from '@app/common/routes'
import { formatScore } from '@app/common/format'

interface Props {
  contestId: string
  leaderboard: UseQueryResult<LeaderboardType>
}

export const Leaderboard = ({ contestId, leaderboard }: Props) => {
  if (leaderboard.isLoading || leaderboard.isIdle) {
    return <>Loading...</>
  }

  if (leaderboard.isError) {
    return (
      <Flash style="error" IconComponent={ExclamationCircleIcon}>
        Could not retrieve leaderboard, please try again later.
      </Flash>
    )
  }

  return (
    <>
      <div className="table-container">
        <table className="default">
          <thead>
            <tr>
              <th className="default !text-center">Rank</th>
              <th className="default">User</th>
              <th className="default !text-right">Score</th>
            </tr>
          </thead>
          <tbody>
            {leaderboard.data.entries.map(it => (
              <tr key={it.rank} className="link">
                <td className="link w-10">
                  <Link
                    href={routes.contestUserProfile(contestId, it.user_id)}
                    className="reset justify-center text-lg"
                  >
                    {it.is_tie ? `T${it.rank}` : it.rank}
                  </Link>
                </td>
                <td className="link">
                  <Link
                    href={routes.contestUserProfile(contestId, it.user_id)}
                    className="reset text-lg"
                  >
                    {it.user_display_name}
                  </Link>
                </td>
                <td className="link">
                  <Link
                    href={routes.contestUserProfile(contestId, it.user_id)}
                    className="reset justify-end text-lg"
                  >
                    {formatScore(it.score)}
                  </Link>
                </td>
              </tr>
            ))}
            {leaderboard.data.entries.length === 0 ? (
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
    </>
  )
}
