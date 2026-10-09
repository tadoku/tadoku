import { useRouter } from 'next/router'
import { Breadcrumb, Flash, Loading, Pagination, Tabbar } from 'ui'
import { HomeIcon, InformationCircleIcon } from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'
import { useProfileLogs, useUserProfile } from '@app/immersion/api'
import {
  getQueryStringDateParameter,
  getQueryStringPageParameter,
} from '@app/common/router'
import Head from 'next/head'
import Link from 'next/link'
import { DateTime } from 'luxon'
import { useEffect, useState } from 'react'
import LogsList from '@app/immersion/LogsList'

const Page = () => {
  const router = useRouter()
  const userId = router.query['id']?.toString() ?? ''
  const date = getQueryStringDateParameter(router.query.date)

  const newFilter = () => {
    return {
      page: getQueryStringPageParameter(router.query.page),
      pageSize: 50,
      includeDeleted: false,
      userId,
      date,
    }
  }

  const [filters, setFilters] = useState(() => newFilter())
  useEffect(() => {
    setFilters({
      page: getQueryStringPageParameter(router.query.page),
      pageSize: 50,
      includeDeleted: false,
      userId,
      date,
    })
  }, [router.asPath, userId, router.query.page, date])

  const profile = useUserProfile({ userId })
  const logs = useProfileLogs(filters)

  if (profile.isLoading) {
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

      {date ? (
        <Flash
          style="info"
          IconComponent={InformationCircleIcon}
          className="mt-4"
        >
          Showing updates on{' '}
          {DateTime.fromISO(date).toLocaleString(DateTime.DATE_MED)}
          <Link
            href={routes.userProfileUpdates(userId)}
            className="ml-auto underline"
          >
            Show all updates
          </Link>
        </Flash>
      ) : null}

      <div className="card p-0 mt-4">
        <LogsList logs={logs} />
      </div>

      {logsTotalPages > 1 ? (
        <div className="mt-4">
          <Pagination
            currentPage={filters.page}
            totalPages={logsTotalPages}
            getHref={page => routes.userProfileUpdates(userId, page, date)}
          />
        </div>
      ) : null}
    </>
  )
}

export default Page
