import type { NextPage } from 'next'
import { Breadcrumb, ButtonGroup, Loading, Pagination, Tabbar } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { PlusIcon } from '@heroicons/react/24/solid'
import {
  useContestConfigurationOptions,
  useContestCreatePermissionCheck,
  useContestList,
} from '@app/immersion/api'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/router'
import { getQueryStringIntParameter } from '@app/common/router'
import { ContestList } from '@app/immersion/ContestList'
import { useSession } from '@app/common/session'
import { routes } from '@app/common/routes'
import Head from 'next/head'

interface Props {}

const Contests: NextPage<Props> = () => {
  const router = useRouter()

  const newFilter = () => {
    return {
      page: getQueryStringIntParameter(router.query.page, 1),
      pageSize: 25,
      official: false,
      includeDeleted: false,
    }
  }
  const [filters, setFilters] = useState(() => newFilter())
  useEffect(() => {
    setFilters(newFilter())
  }, [router.asPath])

  const [session] = useSession()
  const list = useContestList(filters)
  const options = useContestConfigurationOptions({ enabled: !!session })
  const createContestPermission = useContestCreatePermissionCheck({
    enabled: !!session,
  })

  return (
    <>
      <Head>
        <title>User contests - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            { label: 'User contests', href: routes.contestListUserContests() },
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
                visible: createContestPermission.isSuccess,
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
            active: true,
          },
          {
            href: routes.contestListMyContests(),
            label: 'My contests',
            active: false,
          },
        ]}
      />
      <div className="mt-2 md:mt-8">
        {list.isLoading ? <Loading /> : null}
        {list.isError ? (
          <span className="flash error">
            Could not load page, please try again later.
          </span>
        ) : null}
        {list.isSuccess ? (
          <>
            <ContestList list={list.data} options={options.data} />
            {list.data.total_size / filters.pageSize > 1 ? (
              <div className="mt-8">
                <Pagination
                  currentPage={filters.page}
                  totalPages={Math.ceil(
                    list.data.total_size / filters.pageSize,
                  )}
                  getHref={page => routes.contestListUserContests(page)}
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
