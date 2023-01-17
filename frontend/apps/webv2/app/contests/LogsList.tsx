import { ChevronRightIcon } from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'
import { DateTime } from 'luxon'
import Link from 'next/link'
import { Logs } from '@app/contests/api'
import { UseQueryResult } from 'react-query'

function truncate(text: string | undefined, len: number) {
  if (text === undefined) {
    return text
  }

  if (text.length > len - 3) {
    return text.substring(0, len - 3) + '...'
  }

  return text
}

interface Props {
  logs: UseQueryResult<Logs>
}

const LogsList = ({ logs }: Props) => {
  if (logs.isLoading || logs.isIdle) {
    return <p>Loading...</p>
  }

  if (logs.isError) {
    return <span className="flash error">Could not load updates</span>
  }

  return (
    <div className="table-container shadow-transparent w-auto">
      <table className="default shadow-transparent">
        <thead>
          <tr>
            <th className="default w-36">Date</th>
            <th className="default w-32">Language</th>
            <th className="default w-28 hidden md:table-cell">Activity</th>
            <th className="default hidden lg:table-cell">Description</th>
            <th className="default w-36 hidden md:table-cell">Amount</th>
            <th className="default w-24 !text-right">Score</th>
            <th className="default"></th>
          </tr>
        </thead>
        <tbody>
          {logs.data.logs.map(it => (
            <tr key={it.id} className="link">
              <td className="default link">
                <Link className="reset" href={routes.log(it.id)}>
                  {DateTime.fromISO(it.created_at).toLocaleString(
                    DateTime.DATE_MED,
                  )}
                </Link>
              </td>
              <td className="default link">
                <Link className="reset" href={routes.log(it.id)}>
                  {it.language.name}
                </Link>
              </td>
              <td className="default link hidden md:table-cell">
                <Link className="reset" href={routes.log(it.id)}>
                  {it.activity.name}
                </Link>
              </td>
              <td
                className={`default text-sm link hidden lg:table-cell ${
                  !it.description ? 'opacity-50' : ''
                }`}
              >
                <Link className="reset" href={routes.log(it.id)}>
                  {truncate(it.description, 38) ?? 'N/A'}
                </Link>
              </td>
              <td className="default link font-bold hidden md:table-cell">
                <Link className="reset" href={routes.log(it.id)}>
                  {it.amount} {it.unit_name}
                  {it.amount !== 1 ? 's' : ''}
                </Link>
              </td>
              <td className="default link font-bold">
                <Link className="reset justify-end" href={routes.log(it.id)}>
                  {it.score}
                </Link>
              </td>
              <td className="default link w-12">
                <Link className="reset flex-shrink" href={routes.log(it.id)}>
                  <ChevronRightIcon className="w-5 h-5" />
                </Link>
              </td>
            </tr>
          ))}
          {logs.data.logs.length === 0 ? (
            <tr>
              <td
                colSpan={3}
                className="default h-32 font-bold text-center text-xl text-slate-400"
              >
                No updates submitted yet
              </td>
            </tr>
          ) : null}
        </tbody>
      </table>
    </div>
  )
}

export default LogsList
