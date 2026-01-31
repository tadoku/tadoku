import { routes } from '@app/common/routes'
import { HomeIcon } from '@heroicons/react/20/solid'
import type { NextPage } from 'next'
import Head from 'next/head'
import { Breadcrumb } from 'ui'

interface Props {}

const Page: NextPage<Props> = () => {
  return (
    <>
      <Head>
        <title>Admin - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: 'Admin',
              href: routes.manage(),
            },
          ]}
        />
      </div>
      <div>
        <h1 className="title">Admin</h1>
      </div>
    </>
  )
}

export default Page
