import { routes } from '@app/common/routes'
import { DocumentDuplicateIcon, HomeIcon, PlusIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb } from 'ui'
import { NextPageWithLayout } from '../../_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { ContentEditor } from '@app/content/ContentEditor'
import { pagesConfig } from '@app/content/pages'
import { useNamespace } from '@app/content/NamespaceSelector'

const Page: NextPageWithLayout = () => {
  const namespace = useNamespace()

  return (
    <>
      <Head>
        <title>New Page - Admin - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            {
              label: 'Admin',
              href: routes.home(),
              IconComponent: HomeIcon,
            },
            {
              label: 'Pages',
              href: routes.pages(namespace),
              IconComponent: DocumentDuplicateIcon,
            },
            {
              label: 'New Page',
              href: routes.pageNew(namespace),
              IconComponent: PlusIcon,
            },
          ]}
        />
      </div>
      <h1 className="title mb-6">New Page</h1>
      <ContentEditor config={pagesConfig} />
    </>
  )
}

Page.getLayout = getDashboardLayout('pages')

export default Page
