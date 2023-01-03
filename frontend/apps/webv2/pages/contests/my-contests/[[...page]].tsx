import type { NextPage } from 'next'
import { Breadcrumb, ButtonGroup, Pagination, Tabbar } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { PlusIcon } from '@heroicons/react/24/solid'
import {
  useContestConfigurationOptions,
  useContestList,
} from '@app/contests/api'
import { useState } from 'react'
import { useRouter } from 'next/router'
import { getQueryStringIntParameter } from '@app/common/router'
import { useSessionOrRedirect } from '@app/common/session'
import { ContestList } from '@app/contests/ContestList'
import { routes } from '@app/common/routes'

interface Props {}

const Contests: NextPage<Props> = () => {
  const router = useRouter()
  const [session, _] = useSessionOrRedirect()

  const [filters, setFilters] = useState(() => {
    return {
      page: getQueryStringIntParameter(router.query.page, 1),
      pageSize: 25,
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
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            { label: 'My contests', href: routes.contestListMyContests() },
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
                href: routes.contestNew(),
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
            href: routes.contestListOfficial(),
            label: 'Official contests',
            active: false,
          },
          {
            href: routes.contestListUserContests(),
            label: 'User contests',
            active: false,
          },
          {
            href: routes.contestListMyContests(),
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
                    router.push(routes.contestListMyContests(page))
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
