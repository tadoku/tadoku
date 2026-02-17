import { Breadcrumb, Pagination, Tabbar, VerticalTabbar } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'
import Head from 'next/head'
import { Leaderboard } from '@app/immersion/Leaderboard'
import { useYearlyLeaderboard } from '@app/immersion/api'
import { useEffect, useState } from 'react'
import { getQueryStringIntParameter } from '@app/common/router'
import { useRouter } from 'next/router'
import { DateTime, Interval } from 'luxon'

const Page = () => {
  const newFilter = () => {
    return {
      page: getQueryStringIntParameter(router.query.page, 1),
      year,
      pageSize: 50,
    }
  }

  const router = useRouter()
  const year = getQueryStringIntParameter(
    router.query['year'],
    new Date().getFullYear(),
  )
  const [filters, setFilters] = useState(() => newFilter())
  const leaderboard = useYearlyLeaderboard(filters)

  useEffect(() => {
    setFilters({
      page: getQueryStringIntParameter(router.query.page, 1),
      year,
      pageSize: 50,
    })
  }, [router.asPath, year, router.query.page])

  const showPagination =
    leaderboard.data && leaderboard.data.total_size > filters.pageSize

  const years = Interval.fromDateTimes(
    DateTime.fromObject({ year: 2020, month: 1, day: 1 }),
    DateTime.now(),
  )
    .splitBy({ year: 1 })
    .map(it => it.start!!.year)
    .reverse()

  return (
    <>
      <Head>
        <title>{year} leaderboard - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: 'Leaderboard',
              href: routes.leaderboardYearlyOfficial(year),
            },
          ]}
        />
      </div>
      <div className="h-stack justify-between items-center w-full">
        <div>
          <h1 className="title">Leaderboard</h1>
          <h2 className="subtitle">{year}</h2>
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
            active: true,
            href: routes.leaderboardYearlyOfficial(),
            label: 'Yearly',
            disabled: false,
          },
          {
            active: false,
            href: routes.leaderboardAllTimeOfficial(),
            label: 'All time',
            disabled: false,
          },
        ]}
      />
      <div className="flex mt-4 space-x-4">
        <div className="flex-grow">
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
                getHref={page => routes.leaderboardYearlyOfficial(year, page)}
              />
            </div>
          ) : null}
          <p className="text-sm text-slate-500 mt-4 text-center italic">
            Scores may take a few seconds to update.
          </p>
        </div>
        <div className="flex-shrink pl-3 min-w-32">
          <VerticalTabbar
            links={years.map(it => ({
              href: routes.leaderboardYearlyOfficial(it),
              label: it.toString(),
              active: year === it,
            }))}
          />
        </div>
      </div>
    </>
  )
}

export default Page
