import { useRouter } from 'next/router'
import {
  Breadcrumb,
  Flash,
  HeatmapChart,
  Loading,
  Tabbar,
  VerticalTabbar,
} from 'ui'
import {
  ChevronRightIcon,
  ExclamationCircleIcon,
  EyeSlashIcon,
  HomeIcon,
  InformationCircleIcon,
} from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'
import Link from 'next/link'
import { ScoreList } from '@app/immersion/ScoreList'
import {
  useProfileScores,
  useUserProfile,
  useUserYearlyActivity,
  useUserYearlyActivitySplit,
  useYearlyContestRegistrations,
} from '@app/immersion/api'
import { getQueryStringIntParameter } from '@app/common/router'
import { DateTime, Interval } from 'luxon'
import { formatScore } from '@app/common/format'
import { ActivitySplitChart } from '@app/immersion/ActivitySplitChart'
import Head from 'next/head'

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
  const registrations = useYearlyContestRegistrations({ userId, year })
  const activitySplit = useUserYearlyActivitySplit({ userId, year })

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

  const years = Interval.fromDateTimes(
    DateTime.fromISO(profile.data.created_at).startOf('year'),
    DateTime.now(),
  )
    .splitBy({ year: 1 })
    .map(it => it.start.year)
    .reverse()

  return (
    <>
      <Head>
        <title>Profile statistics - {profile.data.display_name} - Tadoku</title>
      </Head>
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
                : 'Updates'}
            </h3>
            <div className="w-full max-h-28 mt-4">
              <Flash
                style="error"
                IconComponent={ExclamationCircleIcon}
                visible={activitySummary.isError}
                className="-mx-4 -mb-4"
              >
                Could not retrieve updates
              </Flash>
              {activitySummary.isLoading ? <Loading /> : null}
              {activitySummary.isSuccess ? (
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
                          )} points on ${DateTime.fromISO(
                            it.date,
                          ).toLocaleString(DateTime.DATE_MED)}`,
                        }))
                      : []
                  }
                />
              ) : null}
            </div>
          </div>
          <div className="h-stack spaced items-start">
            <div className="card w-full p-0">
              <h3 className="subtitle p-4">Contests</h3>
              <Flash
                style="error"
                IconComponent={ExclamationCircleIcon}
                visible={registrations.isError}
              >
                Could not retrieve contests
              </Flash>
              <ul
                className={`divide-y-2 divide-slate-500/5 border-slate-500/5 ${
                  !registrations.isError &&
                  !(
                    registrations.isSuccess &&
                    registrations.data.total_size === 0
                  )
                    ? 'border-t-2'
                    : ''
                }`}
              >
                {registrations.isLoading ? <Loading className="p-4" /> : null}
                {registrations.data
                  ? registrations.data.registrations.map(it => (
                      <li key={it.id}>
                        <Link
                          href={routes.contestUserProfile(
                            it.contest_id,
                            userId,
                          )}
                          className="reset px-4 py-2 flex justify-start items-center hover:bg-slate-500/5"
                        >
                          {it.contest!.private ? (
                            <EyeSlashIcon
                              className="w-5 h-5 mr-2"
                              title="Private contest, only visible to you"
                            />
                          ) : null}
                          <span className="font-bold text-base">
                            {it.contest!.title}
                          </span>

                          <ChevronRightIcon className="w-5 h-5 ml-auto" />
                        </Link>
                      </li>
                    ))
                  : null}
                <Flash
                  style="info"
                  IconComponent={InformationCircleIcon}
                  visible={
                    registrations.isSuccess &&
                    registrations.data.total_size === 0
                  }
                  className="-mt-0"
                >
                  Hasn&apos;t participated in any contests
                </Flash>
              </ul>
            </div>
            <div className="card narrow w-full">
              <h3 className="subtitle mb-4">Activities</h3>
              {activitySplit.isLoading ? <Loading /> : null}
              {activitySplit.data &&
              activitySplit.data.activities.length > 0 ? (
                <ActivitySplitChart
                  activities={activitySplit.data.activities}
                />
              ) : null}

              <Flash
                style="info"
                IconComponent={InformationCircleIcon}
                visible={
                  !activitySplit.isLoading &&
                  activitySplit.data &&
                  activitySplit.data.activities.length === 0
                }
                className="-mx-4 -mb-4"
              >
                Not enough data to show activity split
              </Flash>
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
