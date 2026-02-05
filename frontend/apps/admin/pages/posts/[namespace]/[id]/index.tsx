import { routes } from '@app/common/routes'
import { DocumentTextIcon, HomeIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb } from 'ui'
import { NextPageWithLayout } from '../../../_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { ContentPreview } from '@app/content/ContentPreview'
import { postsConfig } from '@app/content/posts'
import { useRouter } from 'next/router'
import { useNamespace } from '@app/content/NamespaceSelector'

const Page: NextPageWithLayout = () => {
  const router = useRouter()
  const id = router.query.id as string
  const namespace = useNamespace()

  return (
    <>
      <Head>
        <title>Post - Admin - Tadoku</title>
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
            {
              label: 'View',
              href: id ? routes.postPreview(namespace, id) : '#',
            },
          ]}
        />
      </div>
      {id ? <ContentPreview config={postsConfig} id={id} /> : null}
    </>
  )
}

Page.getLayout = getDashboardLayout('posts')

export default Page
