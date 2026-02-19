import type { NextPage } from 'next'
import { Breadcrumb, Loading } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import {
  useLogConfigurationOptions,
  useOngoingContestRegistrations,
} from '@app/immersion/api'
import { routes } from '@app/common/routes'
import { LogForm } from '@app/immersion/NewLogForm/Form'
import { LogFormV2 } from '@app/immersion/NewLogFormV2/Form'
import { useRouter } from 'next/router'
import { getQueryStringIntParameter } from '@app/common/router'
import Head from 'next/head'
import { useSessionOrRedirect, useUserRole } from '@app/common/session'

interface Props {}

const Page: NextPage<Props> = () => {
  const role = useUserRole()
  const options = useLogConfigurationOptions()
  const registrations = useOngoingContestRegistrations({
    enabled: role !== 'admin',
  })

  useSessionOrRedirect()

  const router = useRouter()
  const amount = getQueryStringIntParameter(router.query['amount'], 0)

  if (role === undefined) {
    return <Loading />
  }

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
      {role === 'admin' ? (
        <>
          {options.isLoading ? <Loading /> : null}
          {options.isError ? (
            <span className="flash error">
              Could not load page, please try again later.
            </span>
          ) : null}
          {options.isSuccess ? (
            <LogFormV2
              options={options.data}
              defaultValues={{ amountValue: amount }}
            />
          ) : null}
        </>
      ) : (
        <>
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
              defaultValues={{ amountValue: amount }}
            />
          ) : null}
        </>
      )}
    </>
  )
}

export default Page
