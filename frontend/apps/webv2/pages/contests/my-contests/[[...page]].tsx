import type { NextPage } from 'next'
import { Breadcrumb, Pagination, Tabbar } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { PlusIcon } from '@heroicons/react/24/solid'
import Link from 'next/link'
import {
  ContestConfigurationOptions,
  Contests,
  useContestConfigurationOptions,
  useContestList,
} from '@app/contests/api'
import { useState } from 'react'
import { DateTime } from 'luxon'
import { useRouter } from 'next/router'
import { getQueryStringIntParameter } from '@app/common/router'
import { useSessionOrRedirect } from '@app/common/session'
import { number } from 'zod'
import { EyeSlashIcon } from '@heroicons/react/24/outline'

interface Props {}

const Contests: NextPage<Props> = () => {
  const router = useRouter()
  const [session, _] = useSessionOrRedirect()

  const [filters, setFilters] = useState(() => {
    return {
      page: getQueryStringIntParameter(router.query.page, 1),
      pageSize: 50,
      official: false,
      includeDeleted: false,
      userId: session?.identity.id,
    }
  })
  const options = useContestConfigurationOptions({ enabled: !!session })
  const list = useContestList(filters, { enabled: !!options.data })

  if (!session) {
    return null
  }

  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: '/', IconComponent: HomeIcon },
            { label: 'Contests', href: '/contests' },
          ]}
        />
      </div>
      <div className="h-stack justify-between items-center mb-4">
        <h1 className="title">Contests</h1>
        <div className="h-stack justify-end">
          <Link href="/contests/new" className="btn secondary">
            <PlusIcon className="mr-2" />
            Create contest
          </Link>
        </div>
      </div>
      <Tabbar
        links={[
          {
            href: '/contests/official',
            label: 'Official contests',
            active: false,
          },
          {
            href: '/contests/user-contests',
            label: 'User contests',
            active: false,
          },
          {
            href: '/contests/my-contests',
            label: 'My contests',
            active: true,
          },
        ]}
      />
      <div className="mt-2 md:mt-8">
        {list.isLoading ? <p>Loading...</p> : null}
        {list.isError ? (
          <span className="flash error">
            Could not load page, please try again later.
          </span>
        ) : null}
        {list.isSuccess ? (
          <>
            <ContestList list={list.data} options={options.data} />
            {list.data.totalSize / filters.pageSize > 1 ? (
              <div className="mt-8">
                <Pagination
                  currentPage={filters.page}
                  totalPages={Math.ceil(list.data.totalSize / filters.pageSize)}
                  onClick={page => {
                    setFilters({ ...filters, page })
                    router.push(`/contests/official/${page}`)
                  }}
                />
              </div>
            ) : null}
          </>
        ) : null}
      </div>
    </>
  )
}

export default Contests

const ContestList = ({
  list,
  options,
}: {
  list: Contests
  options: ContestConfigurationOptions | undefined
}) => {
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
                  {c.private ? <EyeSlashIcon className="w-5 h-5 mr-2" /> : null}
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
