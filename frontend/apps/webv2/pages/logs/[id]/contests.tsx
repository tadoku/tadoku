import { Breadcrumb, Loading } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { useLog, useOngoingContestRegistrations } from '@app/immersion/api'
import { routes } from '@app/common/routes'
import { SubmitToContest } from '@app/immersion/SubmitToContest/SubmitToContest'
import { useRouter } from 'next/router'
import Head from 'next/head'
import { useSessionOrRedirect } from '@app/common/session'

const Page = () => {
  const router = useRouter()
  const id = router.query['id']?.toString() ?? ''
  const log = useLog(id, { enabled: !!id })
  const registrations = useOngoingContestRegistrations()

  useSessionOrRedirect()

  if (log.isLoading || log.isIdle || registrations.isLoading || registrations.isIdle) {
    return <Loading />
  }

  if (log.isError || registrations.isError) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  return (
    <>
      <Head>
        <title>Submit to contests - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: log.data.user_display_name!,
              href: routes.userProfileStatistics(log.data.user_id),
            },
            { label: 'Log details', href: routes.log(log.data.id) },
            {
              label: 'Contests',
              href: routes.logContests(log.data.id),
            },
          ]}
        />
      </div>
      <h1 className="title mb-4">Submit to contests</h1>
      <div className="max-w-2xl">
        <SubmitToContest
          log={log.data}
          registrations={registrations.data.registrations}
        />
      </div>
    </>
  )
}

export default Page
