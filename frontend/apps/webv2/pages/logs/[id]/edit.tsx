import { Breadcrumb, Loading } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { useLog, useLogConfigurationOptions } from '@app/immersion/api'
import { routes } from '@app/common/routes'
import { EditLogForm } from '@app/immersion/EditLogForm/Form'
import { useRouter } from 'next/router'
import Head from 'next/head'
import { useSessionOrRedirect, useUserRole } from '@app/common/session'

const Page = () => {
  const router = useRouter()
  const id = router.query['id']?.toString() ?? ''
  const role = useUserRole()
  const log = useLog(id, { enabled: !!id })
  const options = useLogConfigurationOptions()

  useSessionOrRedirect()

  if (role === undefined || log.isLoading || log.isIdle || options.isLoading) {
    return <Loading />
  }

  if (role !== 'admin') {
    return (
      <span className="flash error">This feature is not yet available.</span>
    )
  }

  if (log.isError || options.isError || !log.data || !options.data) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  return (
    <>
      <Head>
        <title>Edit log - Tadoku</title>
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
            { label: 'Edit', href: routes.logEdit(log.data.id) },
          ]}
        />
      </div>
      <h1 className="title mb-4">Edit log</h1>
      <EditLogForm options={options.data} log={log.data} />
    </>
  )
}

export default Page
