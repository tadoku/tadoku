import { routes } from '@app/common/routes'
import { DocumentDuplicateIcon } from '@heroicons/react/20/solid'
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
            {
              label: 'Pages',
              href: routes.managePages(),
              IconComponent: DocumentDuplicateIcon,
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
