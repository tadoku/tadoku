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
  const contest = useContest(id)

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

  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            { label: 'Logs', href: '#' },
          ]}
        />
      </div>
    </>
  )
}

export default Page
