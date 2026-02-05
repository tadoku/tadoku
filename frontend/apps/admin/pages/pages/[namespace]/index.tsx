import { routes } from '@app/common/routes'
import { DocumentDuplicateIcon, HomeIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb } from 'ui'
import { NextPageWithLayout } from '../../_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { ContentList } from '@app/content/ContentList'
import { pagesConfig } from '@app/content/pages'
import { NamespaceSelector, useNamespace } from '@app/content/NamespaceSelector'
import Link from 'next/link'
import { useRouter } from 'next/router'

const Page: NextPageWithLayout = () => {
  const router = useRouter()
  const namespace = useNamespace()

  const handleNamespaceChange = (ns: string) => {
    router.push(routes.pages(ns))
  }

  return (
    <>
      <Head>
        <title>Pages - Admin - Tadoku</title>
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
          ]}
        />
      </div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="title">Pages</h1>
        <div className="flex items-center gap-2">
          <NamespaceSelector value={namespace} onChange={handleNamespaceChange} />
          <Link href={pagesConfig.routes.new(namespace)} className="btn primary">
            New {pagesConfig.label}
          </Link>
        </div>
      </div>
      <ContentList config={pagesConfig} namespace={namespace} />
    </>
  )
}

Page.getLayout = getDashboardLayout('pages')

export default Page
