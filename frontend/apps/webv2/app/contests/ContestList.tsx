import Link from 'next/link'
import { ContestConfigurationOptions, Contests } from '@app/contests/api'
import { DateTime } from 'luxon'
import { EyeSlashIcon } from '@heroicons/react/24/outline'

interface Props {
  list: Contests
  options?: ContestConfigurationOptions | undefined
}

export const ContestList = ({ list, options }: Props) => {
  const languages = new Map<string, string>()
  const activities = new Map<number, string>()

  if (options) {
    options.activities.forEach(a => {
      activities.set(a.id, a.name)
    })
    options.languages.forEach(l => {
      languages.set(l.code, l.name)
    })
  }

  return (
    <div className="table-container">
      <table className="default">
        <thead>
          <tr>
            <th className="default">Round</th>
            <th className="default">Starting date</th>
            <th className="default">Ending date</th>
            {options ? (
              <>
                <th className="default">Languages</th>
                <th className="default">Activities</th>
              </>
            ) : null}
          </tr>
        </thead>
        <tbody>
          {list.contests.map(c => (
            <tr key={c.id} className="link">
              <td className="link">
                <Link href={`/contests/${c.id}`} className="reset">
                  {c.private ? (
                    <EyeSlashIcon
                      className="w-5 h-5 mr-2"
                      title="Private contest, only visible to those with the link"
                    />
                  ) : null}
                  {c.description}
                </Link>
              </td>
              <td className="link">
                <Link href={`/contests/${c.id}`} className="reset">
                  {c.contestStart.toLocaleString(DateTime.DATE_FULL)}
                </Link>
              </td>
              <td className="link">
                <Link href={`/contests/${c.id}`} className="reset">
                  {c.contestEnd.toLocaleString(DateTime.DATE_FULL)}
                </Link>
              </td>
              {options ? (
                <>
                  <td className="default text-ellipsis">
                    {c.languageCodeAllowList
                      .map(l => languages.get(l))
                      .join(', ')}
                  </td>
                  <td className="default text-ellipsis">
                    {c.activityTypeIdAllowList
                      .map(it => activities.get(it))
                      .join(', ')}
                  </td>
                </>
              ) : null}
            </tr>
          ))}
          {list.contests.length === 0 ? (
            <tr>
              <td
                colSpan={3}
                className="default h-32 font-bold text-center text-xl text-slate-400"
              >
                No contests founds
              </td>
            </tr>
          ) : null}
        </tbody>
      </table>
    </div>
  )
}
