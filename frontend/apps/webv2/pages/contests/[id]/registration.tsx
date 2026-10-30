import { useCurrentDateTime } from '@app/common/hooks'
import { useRoutes } from '@app/common/routes'
import { useSessionOrRedirect } from '@app/common/session'
import {
  acceptsNewRegistrations,
  useContest,
  useContestRegistration,
} from '@app/immersion/contest-api'
import { ContestRegistrationForm } from '@app/immersion/ContestRegistration'
import { ExclamationCircleIcon, HomeIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { useRouter } from 'next/router'
import { Breadcrumb, Flash, Loading } from 'ui'

const Page = () => {
  const routes = useRoutes()
  const router = useRouter()
  const id = router.query['id']?.toString() ?? ''

  const now = useCurrentDateTime()

  const contest = useContest(id)
  const [session] = useSessionOrRedirect()

  const registration = useContestRegistration(id, { enabled: !!session })

  if (!session || contest.isLoading) {
    return <Loading />
  }

  if (contest.isError) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  if (registration.isError) {
    return (
      <div role="alert" className="v-stack spaced">
        <Flash style="error" IconComponent={ExclamationCircleIcon}>
          Could not load your registration, please try again.
        </Flash>
        <div>
          <button
            type="button"
            className="btn primary"
            onClick={() => registration.refetch()}
          >
            Try again
          </button>
        </div>
      </div>
    )
  }

  const registrationClosed =
    registration.isSuccess &&
    registration.data === null &&
    !acceptsNewRegistrations(contest.data, now)

  return (
    <>
      <Head>
        <title>Contest registration - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: contest.data.official
                ? 'Official contests'
                : 'User contests',
              href: contest.data.official
                ? routes.contestListOfficial()
                : routes.contestListUserContests(),
            },
            {
              label: contest.data.title,
              href: routes.contestLeaderboard(id),
            },
            { label: 'Registration', href: routes.contestJoin(id) },
          ]}
        />
      </div>
      <Flash
        visible={registrationClosed}
        style="error"
        IconComponent={ExclamationCircleIcon}
        className="mb-4"
      >
        Unfortunately, registrations for this contest have ended.
      </Flash>

      {registration.isLoading ? (
        <Loading />
      ) : (
        <ContestRegistrationForm
          contest={contest.data}
          data={registration.data ?? undefined}
          isClosed={registrationClosed}
        />
      )}
    </>
  )
}

export default Page
