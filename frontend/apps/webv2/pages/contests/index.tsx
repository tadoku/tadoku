import type { NextPage } from 'next'
import { Breadcrumb } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import Link from 'next/link'
import { Contests, useContestList } from '@app/contests/api'
import { useState } from 'react'

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
      <div className="h-stack justify-between items-center">
        <h1 className="title">Contests</h1>
        <div className="h-stack justify-end">
          <Link href="/contests/new" className="btn secondary">
            Create contest
          </Link>
        </div>
      </div>
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
    <>
      <ul>
        {list.contests.map(c => (
          <li key={c.id}>{c.description}</li>
        ))}
      </ul>
    </>
  )
}
