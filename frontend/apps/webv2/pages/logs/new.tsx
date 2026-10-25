import type { NextPage } from 'next'
import { Breadcrumb, Loading } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { useLogConfigurationOptions } from '@app/immersion/api'
import { routes } from '@app/common/routes'
import { LogFormV2 } from '@app/immersion/NewLogFormV2/Form'
import { useRouter } from 'next/router'
import { getQueryStringIntParameter } from '@app/common/router'
import Head from 'next/head'
import { useSessionOrRedirect } from '@app/common/session'

interface Props {}

const Page: NextPage<Props> = () => {
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
  )
}

export default Page
