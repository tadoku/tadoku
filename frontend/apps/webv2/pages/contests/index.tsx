import type { NextPage } from 'next'
import { Breadcrumb } from 'ui'
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
      <nav className="relative h-12 mb-8 flex space-x-10">
        <Link
          href="/contests"
          className="border-b-4 border-primary h-full flex items-center z-10 font-bold"
        >
          Official contests
        </Link>
        <Link
          href="/contests/user-contests"
          className="border-b-4 border-transparent hover:border-primary h-full inline-flex flex-col justify-center items-center z-10 hover:font-semibold before:content-[attr(data-label)] before:h-0 before:overflow-hidden before:font-bold before:pointer-events-none before:select-none"
          data-label="User contests"
        >
          User contests
        </Link>
        <Link
          href="/contests/my-contests"
          className="border-b-4 border-transparent hover:border-primary h-full flex items-center z-10 hover:font-bold"
        >
          My contests
        </Link>
        <div className="border-b-2 absolute border-slate-200 left-0 right-0 bottom-0 z-0"></div>
      </nav>
      <div>
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
