import { useRouter } from 'next/router'
import { Breadcrumb, Loading, Pagination, Tabbar } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'
import { useProfileLogs, useUserProfile } from '@app/immersion/api'
import { getQueryStringIntParameter } from '@app/common/router'
import Head from 'next/head'
import { useEffect, useState } from 'react'
import LogsList from '@app/immersion/LogsList'

const Page = () => {
  const router = useRouter()
  const userId = router.query['id']?.toString() ?? ''

  const newFilter = () => {
    return {
      page: getQueryStringIntParameter(router.query.page, 1),
      pageSize: 50,
      includeDeleted: false,
      userId,
    }
  }

  const [filters, setFilters] = useState(() => newFilter())
  useEffect(() => {
    setFilters({
      page: getQueryStringIntParameter(router.query.page, 1),
      pageSize: 50,
      includeDeleted: false,
      userId,
    })
  }, [router.asPath, userId, router.query.page])

  const profile = useUserProfile({ userId })
  const logs = useProfileLogs(filters)

  if (profile.isLoading || profile.isIdle) {
    return <Loading />
  }

  if (profile.isError) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  const logsTotalPages = logs.data
    ? Math.ceil(logs.data.total_size / filters.pageSize)
    : 0

  return (
    <>
      <Head>
        <title>Profile updates - {profile.data.display_name} - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: `Profile - ${profile.data.display_name}`,
              href: routes.userProfileUpdates(userId),
            },
          ]}
        />
      </div>
      <div className="h-stack justify-between items-center w-full">
        <div>
          <h1 className="title">Profile</h1>
          <h2 className="subtitle">{profile.data.display_name}</h2>
        </div>
        <div></div>
      </div>
      <Tabbar
        links={[
          {
            href: routes.userProfileStatistics(userId),
            label: 'Statistics',
            active: false,
          },
          {
            href: routes.userProfileUpdates(userId),
            label: 'Updates',
            active: true,
          },
        ]}
      />

      <div className="card p-0 mt-4">
        <LogsList logs={logs} />
      </div>

      {logsTotalPages > 1 ? (
        <div className="mt-4">
          <Pagination
            currentPage={filters.page}
            totalPages={logsTotalPages}
            getHref={page => routes.userProfileUpdates(userId, page)}
          />
        </div>
      ) : null}
    </>
  )
}

export default Page
