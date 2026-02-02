import { routes } from '@app/common/routes'
import { DocumentDuplicateIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb } from 'ui'
import { NextPageWithLayout } from './_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'

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
              href: routes.pages(),
              IconComponent: DocumentDuplicateIcon,
            },
          ]}
        />
      </div>
      <h1 className="title">Pages</h1>
      <p className="mt-2 text-slate-600">Manage static pages.</p>
    </>
  )
}

Page.getLayout = getDashboardLayout('pages')

export default Page
