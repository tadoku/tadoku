import type { NextPage } from 'next'
import { Breadcrumb, Loading } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import {
  useLogConfigurationOptions,
  useOngoingContestRegistrations,
} from '@app/immersion/api'
import { routes } from '@app/common/routes'
import { LogForm } from '@app/immersion/NewLogForm/Form'
import { useRouter } from 'next/router'
import { getQueryStringIntParameter } from '@app/common/router'
import Head from 'next/head'
import { useSessionOrRedirect } from '@app/common/session'

interface Props { }

const Page: NextPage<Props> = () => {
  const registrations = useOngoingContestRegistrations()
  const options = useLogConfigurationOptions()

  useSessionOrRedirect()

  const router = useRouter()
  const amount = getQueryStringIntParameter(router.query['amount'], 0)

  return (
    <>
      <Head>
        <title>New log - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            { label: 'New log', href: routes.logCreate() },
          ]}
        />
      </div>
      <h1 className="title mb-4">New log</h1>
      {options.isLoading || registrations.isLoading ? <Loading /> : null}
      {options.isError || registrations.isError ? (
        <span className="flash error">
          Could not load page, please try again later.
        </span>
      ) : null}
      {options.isSuccess && registrations.isSuccess ? (
        <LogForm
          registrations={registrations.data}
          options={options.data}
          defaultValues={{ amount }}
        />
      ) : null}
    </>
  )
}

export default Page
