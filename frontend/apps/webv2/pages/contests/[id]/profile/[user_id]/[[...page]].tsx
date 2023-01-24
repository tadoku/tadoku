import { useRouter } from 'next/router'
import { Breadcrumb, Pagination } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'
import { ActivityChart } from '@app/immersion/ActivityChart'
import {
  useContestProfileLogs,
  useContestProfileScores,
} from '@app/immersion/api'
import LogsList from '@app/immersion/LogsList'
import { useEffect, useState } from 'react'
import { getQueryStringIntParameter } from '@app/common/router'
import { ScoreList } from '@app/immersion/ScoreList'
import Head from 'next/head'

const Page = () => {
  const router = useRouter()
  const contestId = router.query['id']?.toString() ?? ''
  const userId = router.query['user_id']?.toString() ?? ''

  const newLogListParams = () => {
    return {
      userId,
      contestId,
      includeDeleted: false,
      page: getQueryStringIntParameter(router.query.page, 1),
      pageSize: 50,
    }
  }
  const [logListParams, setLogListParams] = useState(() => newLogListParams())
  useEffect(() => {
    setLogListParams(newLogListParams())
  }, [router.asPath])

  const profile = useContestProfileScores({ userId, contestId })
  const logs = useContestProfileLogs(logListParams)

  if (profile.isLoading || profile.isIdle) {
    return <p>Loading...</p>
  }

  const contest = profile.data?.registration.contest

  if (profile.isError || !contest) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  const logsTotalPages = logs.data
    ? Math.ceil(logs.data.total_size / logListParams.pageSize)
    : 0

  return (
    <>
      <Head>
        <title>
          Contest profile - {profile.data.registration.user_display_name} -
          Tadoku
        </title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: contest.official ? 'Official contests' : 'User contests',
              href: contest.official
                ? routes.contestListOfficial()
                : routes.contestListUserContests(),
            },
            {
              label: contest.title,
              href: routes.contestLeaderboard(contestId),
            },
            {
              label: profile.data.registration.user_display_name,
              href: routes.contestUserProfile(contestId, userId),
            },
          ]}
        />
      </div>
      <div className="h-stack justify-between items-center w-full">
        <div>
          <h1 className="title">
            {profile.data.registration.user_display_name}
          </h1>
          <h2 className="subtitle">{contest.title}</h2>
        </div>
        <div></div>
      </div>
      <div className="my-4 lg:space-x-4 flex flex-col lg:flex-row w-full">
        <div className="lg:w-1/5">
          <ScoreList
            languages={profile.data.registration.languages}
            list={profile.data.scores}
          />
        </div>
        <div className="mt-4 lg:mt-0 flex-grow flex flex-col card narrow">
          <h3 className="subtitle mb-2">Activity</h3>
          <div className="flex-1 max-h-72 lg:max-h-[28rem]">
            <ActivityChart
              userId={userId}
              registration={profile.data.registration}
            />
          </div>
        </div>
      </div>
      <div className="card p-0">
        <div className="h-stack w-full items-center justify-between">
          <h2 className="subtitle p-4">Updates</h2>
        </div>
        <LogsList logs={logs} />
      </div>
      {logsTotalPages > 1 ? (
        <div className="mt-4">
          <Pagination
            currentPage={logListParams.page}
            totalPages={logsTotalPages}
            getHref={page => routes.contestUserProfile(contestId, userId, page)}
          />
        </div>
      ) : null}
    </>
  )
}

export default Page
