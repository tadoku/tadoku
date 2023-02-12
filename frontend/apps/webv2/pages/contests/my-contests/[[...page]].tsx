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
import { useSessionOrRedirect } from '@app/common/session'
import { ContestList } from '@app/immersion/ContestList'
import { routes } from '@app/common/routes'
import Head from 'next/head'

interface Props {}

const Contests: NextPage<Props> = () => {
  const router = useRouter()
  const [session, _] = useSessionOrRedirect()

  const newFilter = () => {
    return {
      page: getQueryStringIntParameter(router.query.page, 1),
      pageSize: 25,
      official: false,
      includeDeleted: false,
      userId: session?.identity.id,
    }
  }
  const [filters, setFilters] = useState(() => newFilter())
  useEffect(() => {
    setFilters(newFilter())
  }, [router.asPath])

  const options = useContestConfigurationOptions({ enabled: !!session })
  const list = useContestList(filters, { enabled: !!options.data })
  const createContestPermission = useContestCreatePermissionCheck({
    enabled: !!session,
  })

  if (!session) {
    return null
  }

  return (
    <>
      <Head>
        <title>My contests - Tadoku</title>
      </Head>
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
        {list.isLoading || options.isLoading ? <Loading /> : null}
        {list.isError || options.isError ? (
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
                  getHref={page => routes.contestListMyContests(page)}
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
