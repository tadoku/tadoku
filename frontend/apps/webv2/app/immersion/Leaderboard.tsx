import { UseQueryResult } from 'react-query'
import { Leaderboard as LeaderboardType } from '@app/immersion/api'
import { Flash, Loading } from 'ui'
import { ExclamationCircleIcon } from '@heroicons/react/20/solid'
import Link from 'next/link'
import { formatScore } from '@app/common/format'

interface Props {
  leaderboard: UseQueryResult<LeaderboardType>
  embedded?: boolean
  emptyMessage?: string
  urlForRow: (userId: string) => string
}

export const Leaderboard = ({
  leaderboard,
  urlForRow,
  embedded = false,
  emptyMessage = 'No partipants yet, be the first to sign up!',
}: Props) => {
  if (leaderboard.isLoading || leaderboard.isIdle) {
    return <Loading className="p-5" />
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
      <div
        className={`table-container ${embedded ? 'shadow-transparent' : ''}`}
      >
        <table className="default">
          <thead>
            <tr>
              <th
                className={`default ${
                  embedded ? '!pl-4 lg:!pl-7' : '!text-center'
                }`}
              >
                Rank
              </th>
              <th className="default">User</th>
              <th
                className={`default !text-right ${
                  embedded ? '!pr-4 lg:!pr-7' : ''
                }`}
              >
                Score
              </th>
            </tr>
          </thead>
          <tbody>
            {leaderboard.data.entries.map(it => (
              <tr key={it.user_id} className="link font-bold">
                <td className="link w-10">
                  <Link
                    href={urlForRow(it.user_id)}
                    className="reset justify-center text-lg"
                  >
                    {it.is_tie ? `T${it.rank}` : it.rank}
                  </Link>
                </td>
                <td className="link">
                  <Link href={urlForRow(it.user_id)} className="reset text-lg">
                    {it.user_display_name}
                  </Link>
                </td>
                <td className="link">
                  <Link
                    href={urlForRow(it.user_id)}
                    className={`reset justify-end text-lg ${
                      embedded ? '!pr-4 lg:!pr-7' : ''
                    }`}
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
                  {emptyMessage}
                </td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>
    </>
  )
}
