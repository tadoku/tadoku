import { useCurrentDateTime, useCurrentLocation } from '@app/common/hooks'
import { routes } from '@app/common/routes'
import { useSession } from '@app/common/session'
import { useContest, useContestRegistration } from '@app/contests/api'
import { ContestRegistrationForm } from '@app/contests/ContestRegistration'
import { ExclamationCircleIcon, HomeIcon } from '@heroicons/react/20/solid'
import { DateTime, Interval } from 'luxon'
import { useRouter } from 'next/router'
import { Breadcrumb, Flash } from 'ui'

const Page = () => {
  const router = useRouter()
  const id = router.query['id']?.toString() ?? ''

  const now = useCurrentDateTime()

  const contest = useContest(id)
  const [session] = useSession()
  const currentUrl = useCurrentLocation()

  const registration = useContestRegistration(id, { enabled: !!session })

  if (!session) {
    router.push(routes.authLogin(currentUrl))
  }

  if (contest.isLoading || contest.isIdle) {
    return <p>Loading...</p>
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
    DateTime.fromISO(contest.data.registration_end),
  ).isBefore(now)

  return (
    <>
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

      {registration.isLoading || registration.isIdle ? (
        'Loading...'
      ) : (
        <ContestRegistrationForm
          contest={contest.data}
          data={registration.data}
          isClosed={registrationClosed}
        />
      )}
    </>
  )
}

export default Page
