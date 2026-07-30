import { routes } from '@app/common/routes'
import { useSessionOrRedirect, useSession } from '@app/common/session'
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
  useSessionOrRedirect()

  const contest = useContest(id, { enabled: !!id })
  const options = useLogConfigurationOptions({ enabled: !!id })
  const isOwner =
    !!session &&
    !!contest.data?.owner_user_id &&
    session.identity?.id === contest.data.owner_user_id
  const contestRuleSets = useContestScoringRuleSets(id, { enabled: isOwner })
  const platformRuleSets = usePlatformScoringRuleSets({ enabled: isOwner })

  if (
    !session ||
    contest.isLoading ||
    contest.isIdle ||
    options.isLoading ||
    options.isIdle
  ) {
    return <Loading />
  }
  if (!isOwner) {
    return <span className="flash error">Only the contest owner can manage scoring.</span>
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
