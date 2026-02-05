import { routes } from '@app/common/routes'
import { DocumentTextIcon, HomeIcon, PencilIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb } from 'ui'
import { NextPageWithLayout } from '../../_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { ContentEditor } from '@app/content/ContentEditor'
import { postsConfig } from '@app/content/posts'
import { useRouter } from 'next/router'

const Page: NextPageWithLayout = () => {
  const router = useRouter()
  const slug = router.query.slug as string

  return (
    <>
      <Head>
        <title>Edit Post - Admin - Tadoku</title>
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
            {
              label: 'Edit',
              href: slug ? routes.postEdit(slug) : '#',
              IconComponent: PencilIcon,
            },
          ]}
        />
      </div>
      <h1 className="title mb-6">Edit Post</h1>
      {slug ? <ContentEditor config={postsConfig} slug={slug} /> : null}
    </>
  )
}

Page.getLayout = getDashboardLayout('posts')

export default Page
