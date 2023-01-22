import type { NextPage } from 'next'
import { Breadcrumb, ButtonGroup, Pagination, Tabbar } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { PlusIcon } from '@heroicons/react/24/solid'
import { useContestList } from '@app/immersion/api'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/router'
import { getQueryStringIntParameter } from '@app/common/router'
import { ContestList } from '@app/immersion/ContestList'
import { useSession } from '@app/common/session'
import { routes } from '@app/common/routes'

interface Props {}

const Contests: NextPage<Props> = () => {
  const router = useRouter()

  const newFilter = () => {
    return {
      page: getQueryStringIntParameter(router.query.page, 1),
      pageSize: 25,
      official: true,
      includeDeleted: false,
    }
  }
  const [filters, setFilters] = useState(() => newFilter())
  useEffect(() => {
    setFilters(newFilter())
  }, [router.asPath])

  const list = useContestList(filters)
  const [session] = useSession()

  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            { label: 'Contests', href: routes.contestListOfficial() },
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
                visible: !!session,
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
            active: true,
          },
          {
            href: routes.contestListUserContests(),
            label: 'User contests',
            active: false,
          },
          {
            href: routes.contestListMyContests(),
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
            {list.data.total_size / filters.pageSize > 1 ? (
              <div className="mt-8">
                <Pagination
                  currentPage={filters.page}
                  totalPages={Math.ceil(
                    list.data.total_size / filters.pageSize,
                  )}
                  getHref={page => routes.contestListOfficial(page)}
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
