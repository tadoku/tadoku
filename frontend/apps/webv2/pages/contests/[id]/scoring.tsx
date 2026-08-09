import { routes } from '@app/common/routes'
import {
  useSessionOrRedirect,
  useSession,
  useUserRole,
} from '@app/common/session'
import {
  useContest,
  useContestScoringRuleSets,
  useLogConfigurationOptions,
  usePlatformScoringRuleSets,
} from '@app/immersion/api'
import { ContestScoringRules } from '@app/immersion/ContestScoringRules'
import { HomeIcon } from '@heroicons/react/20/solid'
import { DateTime } from 'luxon'
import Head from 'next/head'
import { useRouter } from 'next/router'
import { Breadcrumb, Loading } from 'ui'

const Page = () => {
  const router = useRouter()
  const id = router.query['id']?.toString() ?? ''
  const [session] = useSession()
  const role = useUserRole()
  useSessionOrRedirect()

  const isAdmin = role === 'admin'
  const contest = useContest(id, { enabled: !!id && isAdmin })
  const options = useLogConfigurationOptions({ enabled: !!id && isAdmin })
  const contestRuleSets = useContestScoringRuleSets(id, { enabled: isAdmin })
  const platformRuleSets = usePlatformScoringRuleSets({ enabled: isAdmin })

  if (!session || role === undefined) {
    return <Loading />
  }
  if (!isAdmin) {
    return (
      <span className="flash error">This feature is not yet available.</span>
    )
  }
  if (
    contest.isLoading ||
    contest.isIdle ||
    options.isLoading ||
    options.isIdle
  ) {
    return <Loading />
  }
  if (
    contest.isError ||
    options.isError ||
    contestRuleSets.isError ||
    platformRuleSets.isError ||
    !contest.data ||
    !options.data
  ) {
    return (
      <span className="flash error">
        Could not load scoring rules, please try again later.
      </span>
    )
  }
  if (
    contestRuleSets.isLoading ||
    contestRuleSets.isIdle ||
    platformRuleSets.isLoading ||
    platformRuleSets.isIdle
  ) {
    return <Loading />
  }

  const locked =
    DateTime.now() >= DateTime.fromISO(contest.data.contest_start).startOf('day')

  return (
    <>
      <Head>
        <title>Contest scoring - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: contest.data.title,
              href: routes.contestLeaderboard(id),
            },
            { label: 'Scoring', href: routes.contestScoring(id) },
          ]}
        />
      </div>
      <h1 className="title">Contest scoring</h1>
      <p className="subtitle mb-4">{contest.data.title}</p>
      <ContestScoringRules
        contestId={id}
        options={options.data}
        platformRuleSets={platformRuleSets.data}
        contestRuleSets={contestRuleSets.data}
        locked={locked}
      />
    </>
  )
}

export default Page
