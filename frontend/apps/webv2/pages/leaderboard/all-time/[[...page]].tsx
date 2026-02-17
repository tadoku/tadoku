import { Breadcrumb, Pagination, Tabbar } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'
import Head from 'next/head'
import { Leaderboard } from '@app/immersion/Leaderboard'
import { useAllTimeLeaderboard } from '@app/immersion/api'
import { useEffect, useState } from 'react'
import { getQueryStringIntParameter } from '@app/common/router'
import { useRouter } from 'next/router'

const Page = () => {
  const newFilter = () => {
    return {
      page: getQueryStringIntParameter(router.query.page, 1),
      pageSize: 50,
    }
  }

  const router = useRouter()
  const [filters, setFilters] = useState(() => newFilter())
  const leaderboard = useAllTimeLeaderboard(filters)

  useEffect(() => {
    setFilters({
      page: getQueryStringIntParameter(router.query.page, 1),
      pageSize: 50,
    })
  }, [router.asPath, router.query.page])

  const showPagination =
    leaderboard.data && leaderboard.data.total_size > filters.pageSize

  return (
    <>
      <Head>
        <title>All time leaderboard - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: 'Leaderboard',
              href: routes.leaderboardAllTimeOfficial(),
            },
          ]}
        />
      </div>
      <div className="h-stack justify-between items-center w-full">
        <div>
          <h1 className="title">Leaderboard</h1>
          <h2 className="subtitle">All time</h2>
        </div>
        <div></div>
      </div>
      <Tabbar
        links={[
          {
            active: false,
            href: routes.leaderboardLatestOfficial(),
            label: 'Latest',
          },
          {
            active: false,
            href: routes.leaderboardYearlyOfficial(),
            label: 'Yearly',
            disabled: false,
          },
          {
            active: true,
            href: routes.leaderboardAllTimeOfficial(),
            label: 'All time',
            disabled: false,
          },
        ]}
      />
      <div className="mt-4">
        <Leaderboard
          leaderboard={leaderboard}
          urlForRow={userId => routes.userProfileStatistics(userId)}
        />
        {showPagination ? (
          <div className="mt-4">
            <Pagination
              currentPage={filters.page}
              totalPages={Math.ceil(
                leaderboard.data.total_size / filters.pageSize,
              )}
              getHref={page => routes.leaderboardAllTimeOfficial(page)}
            />
          </div>
        ) : null}
        <p className="text-sm text-slate-500 mt-4 text-center italic">
          Scores may take a few seconds to update.
        </p>
      </div>
    </>
  )
}

export default Page
