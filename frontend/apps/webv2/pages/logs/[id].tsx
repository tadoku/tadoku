import { routes } from '@app/common/routes'
import { useLog } from '@app/immersion/api'
import { LogDetailsV2 } from '@app/immersion/LogDetailsV2'
import { HomeIcon } from '@heroicons/react/20/solid'
import Link from 'next/link'
import { useRouter } from 'next/router'
import { Breadcrumb, Loading } from 'ui'
import Head from 'next/head'

const Page = () => {
  const router = useRouter()
  const id = router.query['id']?.toString() ?? ''
  const log = useLog(id)

  if (log.isLoading) {
    return <Loading />
  }

  if (log.isError) {
    if (log.error instanceof Error && log.error.message === '404') {
      return (
        <>
          <Head>
            <title>Log not found - Tadoku</title>
          </Head>
          <div className="max-w-2xl">
            <h1 className="title">Log not found</h1>
            <p className="mt-2">
              This log may have been deleted, or the link may be incorrect.
            </p>
            <div className="mt-4">
              <Link href={routes.home()} className="btn">
                Back to home
              </Link>
            </div>
          </div>
        </>
      )
    }

    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  return (
    <>
      <Head>
        <title>Log details - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: log.data.user_display_name!,
              href: routes.userProfileStatistics(log.data.user_id),
            },
            {
              label: 'Log details',
              href: routes.log(log.data.id),
            },
          ]}
        />
      </div>
      <LogDetailsV2 log={log.data} />
    </>
  )
}

export default Page
