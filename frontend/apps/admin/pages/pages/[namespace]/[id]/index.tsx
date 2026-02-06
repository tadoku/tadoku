import { routes } from '@app/common/routes'
import { DocumentDuplicateIcon, HomeIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb } from 'ui'
import { NextPageWithLayout } from '../../../_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { ContentPreview } from '@app/content/ContentPreview'
import { pagesConfig } from '@app/content/pages'
import { useRouter } from 'next/router'
import { useNamespace } from '@app/content/NamespaceSelector'
import { useContentFindById } from '@app/content/api'

const Page: NextPageWithLayout = () => {
  const router = useRouter()
  const id = router.query.id as string
  const namespace = useNamespace()
  const item = useContentFindById(pagesConfig, namespace, id, { enabled: !!id })

  return (
    <>
      <Head>
        <title>Page - Admin - Tadoku</title>
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
              label: item.data?.title ?? 'View',
              href: id ? routes.pagePreview(namespace, id) : '#',
            },
          ]}
        />
      </div>
      {id ? <ContentPreview config={pagesConfig} id={id} /> : null}
    </>
  )
}

Page.getLayout = getDashboardLayout('pages')

export default Page
