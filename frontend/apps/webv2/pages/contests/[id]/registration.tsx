import { useCurrentDateTime } from '@app/common/hooks'
import { useRoutes } from '@app/common/routes'
import { useSessionOrRedirect } from '@app/common/session'
import { useContest, useContestRegistration } from '@app/immersion/contest-api'
import {
  ContestRegistrationForm,
  registrationChangesClosed,
} from '@app/immersion/ContestRegistration'
import { ExclamationCircleIcon, HomeIcon } from '@heroicons/react/20/solid'
import { DateTime, Interval } from 'luxon'
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

  const registrationClosed = Interval.fromDateTimes(
    DateTime.fromISO(contest.data.contest_start),
    DateTime.fromISO(contest.data.registration_end).endOf('day'),
  ).isBefore(now)
  const contestEnded = registrationChangesClosed(contest.data.contest_end, now)

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
        visible={registrationClosed || contestEnded}
        style="error"
        IconComponent={ExclamationCircleIcon}
        className="mb-4"
      >
        {!contestEnded
          ? 'Unfortunately, registrations for this contest have ended.'
          : registration.data
            ? 'This contest has ended, so your registration can no longer be changed.'
            : 'This contest has ended, so it no longer accepts registrations.'}
      </Flash>

      {registration.isLoading ? (
        <Loading />
      ) : (
        <ContestRegistrationForm
          contest={contest.data}
          data={registration.data ?? undefined}
          isClosed={registrationClosed || contestEnded}
        />
      )}
    </>
  )
}

export default Page
