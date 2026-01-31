import { routes } from '@app/common/routes'
import { HomeIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb } from 'ui'
import { NextPageWithLayout } from '../_app'
import { getAdminLayout } from '@app/manage/AdminLayout'

const Page: NextPageWithLayout = () => {
  return (
    <>
      <Head>
        <title>Pages - Admin - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: 'Admin',
              href: routes.manage(),
            },
            {
              label: 'Pages',
              href: routes.managePages(),
            },
          ]}
        />
      </div>
      <h1 className="title">Pages</h1>
    </>
  )
}

Page.getLayout = getAdminLayout('pages')

export default Page
