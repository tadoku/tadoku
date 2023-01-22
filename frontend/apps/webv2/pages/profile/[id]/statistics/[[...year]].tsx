import { useRouter } from 'next/router'
import { Breadcrumb, HeatmapChart, Tabbar, VerticalTabbar } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'
import Link from 'next/link'
import { ScoreList } from '@app/immersion/ScoreList'
import {
  useProfileScores,
  useUserProfile,
  useUserYearlyActivity,
} from '@app/immersion/api'
import { getQueryStringIntParameter } from '@app/common/router'
import { DateTime, Interval } from 'luxon'
import { formatScore } from '@app/common/format'

const Page = () => {
  const router = useRouter()
  const userId = router.query['id']?.toString() ?? ''
  const year = getQueryStringIntParameter(
    router.query['year'],
    new Date().getFullYear(),
  )

  const profile = useUserProfile({ userId })
  const activitySummary = useUserYearlyActivity({ userId, year })
  const scores = useProfileScores({ userId, year })

  if (profile.isLoading || profile.isIdle) {
    return <p>Loading...</p>
  }

  if (profile.isError) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  const years = Interval.fromDateTimes(
    DateTime.fromISO(profile.data.created_at).startOf('year'),
    DateTime.now(),
  )
    .splitBy({ year: 1 })
    .map(it => it.start.year)
    .reverse()

  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: `Profile - ${profile.data.display_name}`,
              href: routes.userProfileStatistics(userId, year),
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
            active: true,
          },
          {
            href: routes.userProfileUpdates(userId),
            label: 'Updates',
            active: false,
            disabled: true,
          },
        ]}
      />
      <div className="h-stack mt-4 spaced">
        <div className="lg:w-1/5">
          <ScoreList
            languages={
              scores.data?.scores.map(it => ({
                code: it.language_code,
                name: it.language_name ?? '',
              })) ?? []
            }
            list={scores.data?.scores ?? []}
          />
        </div>
        <div className="flex-grow v-stack spaced">
          <div className="card narrow">
            <h3 className="subtitle">
              {activitySummary.data
                ? `${activitySummary.data.total_updates} updates in ${year}`
                : 'Loading updates...'}
            </h3>
            <div className="w-full h-28 mt-4">
              <HeatmapChart
                id={`heatmap-${year}-${userId}`}
                year={year}
                data={
                  activitySummary.data
                    ? activitySummary.data.scores.map(it => ({
                        date: it.date,
                        value: it.score,
                        tooltip: `${formatScore(
                          it.score,
                        )} points on ${DateTime.fromISO(it.date).toLocaleString(
                          DateTime.DATE_MED,
                        )}`,
                      }))
                    : []
                }
              />
            </div>
          </div>
          <div className="h-stack spaced flex-grow">
            <div className="card w-full p-0">
              <h3 className="subtitle p-4">Contests</h3>
              <ul className="divide-y-2 divide-slate-500/5 border-t-2 border-slate-500/5">
                {[].map(u => (
                  <li key={`${u[0]}-${u[1]}`}>
                    <Link
                      href="#"
                      className="reset px-4 py-2 flex justify-between items-center hover:bg-slate-500/5"
                    >
                      <span className="font-bold text-base">{u}</span>
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
            <div className="card narrow w-full">
              <h3 className="subtitle">Activities</h3>
              <div className="bg-lime-400 w-full h-64 mt-4"></div>
            </div>
          </div>
        </div>
        <div className="flex-shrink pl-3 min-w-32">
          <VerticalTabbar
            links={years.map(it => ({
              href: routes.userProfileStatistics(userId, it),
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
