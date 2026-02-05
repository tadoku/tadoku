import { routes } from '@app/common/routes'
import { DocumentTextIcon, HomeIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb } from 'ui'
import { NextPageWithLayout } from '../../_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { ContentList } from '@app/content/ContentList'
import { postsConfig } from '@app/content/posts'
import { NamespaceSelector, useNamespace } from '@app/content/NamespaceSelector'
import Link from 'next/link'
import { useRouter } from 'next/router'

const Page: NextPageWithLayout = () => {
  const router = useRouter()
  const namespace = useNamespace()

  const handleNamespaceChange = (ns: string) => {
    router.push(routes.posts(ns))
  }

  return (
    <>
      <Head>
        <title>Posts - Admin - Tadoku</title>
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
              label: 'Posts',
              href: routes.posts(namespace),
              IconComponent: DocumentTextIcon,
            },
          ]}
        />
      </div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="title">Posts</h1>
        <div className="flex items-center gap-2">
          <NamespaceSelector value={namespace} onChange={handleNamespaceChange} />
          <Link href={postsConfig.routes.new(namespace)} className="btn primary">
            New {postsConfig.label}
          </Link>
        </div>
      </div>
      <ContentList config={postsConfig} namespace={namespace} />
    </>
  )
}

Page.getLayout = getDashboardLayout('posts')

export default Page
