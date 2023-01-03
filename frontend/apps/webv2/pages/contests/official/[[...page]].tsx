import type { NextPage } from 'next'
import { Breadcrumb, ButtonGroup, Pagination, Tabbar } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { PlusIcon } from '@heroicons/react/24/solid'
import { useContestList } from '@app/contests/api'
import { useState } from 'react'
import { useRouter } from 'next/router'
import { getQueryStringIntParameter } from '@app/common/router'
import { ContestList } from '@app/contests/ContestList'

interface Props {}

const Contests: NextPage<Props> = () => {
  const router = useRouter()

  const [filters, setFilters] = useState(() => {
    return {
      page: getQueryStringIntParameter(router.query.page, 1),
      pageSize: 25,
      official: true,
      includeDeleted: false,
    }
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
          <ButtonGroup
            orientation="right"
            actions={[
              {
                href: '/contests/new',
                label: 'Create contest',
                style: 'secondary',
                IconComponent: PlusIcon,
              },
            ]}
          />
        </div>
      </div>
      <Tabbar
        links={[
          {
            href: '/contests/official',
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
      <div className="mt-2 md:mt-8">
        {list.isLoading ? <p>Loading...</p> : null}
        {list.isError ? (
          <span className="flash error">
            Could not load page, please try again later.
          </span>
        ) : null}
        {list.isSuccess ? (
          <>
            <ContestList list={list.data} />
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
