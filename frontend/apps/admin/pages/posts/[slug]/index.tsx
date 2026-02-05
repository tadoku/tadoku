import { routes } from '@app/common/routes'
import { DocumentTextIcon, HomeIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb } from 'ui'
import { NextPageWithLayout } from '../../_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { ContentPreview } from '@app/content/ContentPreview'
import { postsConfig } from '@app/content/posts'
import { useRouter } from 'next/router'

const Page: NextPageWithLayout = () => {
  const router = useRouter()
  const slug = router.query.slug as string

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
              href: routes.posts(),
              IconComponent: DocumentTextIcon,
            },
            {
              label: slug ?? '...',
              href: slug ? routes.postPreview(slug) : '#',
            },
          ]}
        />
      </div>
      {slug ? <ContentPreview config={postsConfig} slug={slug} /> : null}
    </>
  )
}

Page.getLayout = getDashboardLayout('posts')

export default Page
