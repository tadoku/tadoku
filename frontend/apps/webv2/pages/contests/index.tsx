import type { NextPage } from 'next'
import { Breadcrumb, Tabbar } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import Link from 'next/link'
import { Contests, useContestList } from '@app/contests/api'
import { useState } from 'react'
import { DateTime } from 'luxon'

interface Props {}

const Contests: NextPage<Props> = () => {
  const [filters, setFilters] = useState({
    page: 0,
    pageSize: 100,
    official: true,
    includeDeleted: false,
  })
  const list = useContestList(filters)

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
            Create contest
          </Link>
        </div>
      </div>
      <Tabbar
        links={[
          {
            href: '/contests',
            label: 'Official contests',
            active: true,
          },
          {
            href: '/contests/user-contests',
            label: 'User contests',
            active: false,
          },
          {
            href: '/contests/my-contests',
            label: 'My contests',
            active: false,
          },
        ]}
      />
      <div className="mt-8">
        {list.isLoading ? <p>Loading...</p> : null}
        {list.isError ? (
          <span className="flash error">
            Could not load page, please try again later.
          </span>
        ) : null}
        {list.isSuccess ? (
          <>
            <ContestList list={list.data} />
          </>
        ) : null}
      </div>
    </>
  )
}

export default Contests

const ContestList = ({ list }: { list: Contests }) => {
  return (
    <div className="table-container">
      <table className="default">
        <thead>
          <tr>
            <th className="default">Round</th>
            <th className="default">Starting date</th>
            <th className="default">Ending date</th>
          </tr>
        </thead>
        <tbody>
          {list.contests.map(c => (
            <tr key={c.id} className="link">
              <td className="link">
                <Link href={`/contests/${c.id}`} className="reset">
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
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
