import { routes } from '@app/common/routes'
import { DocumentTextIcon, HomeIcon, PlusIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb } from 'ui'
import { NextPageWithLayout } from '../_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { ContentEditor } from '@app/content/ContentEditor'
import { postsConfig } from '@app/content/posts'

const Page: NextPageWithLayout = () => {
  return (
    <>
      <Head>
        <title>New Post - Admin - Tadoku</title>
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
              label: 'New Post',
              href: routes.postNew(),
              IconComponent: PlusIcon,
            },
          ]}
        />
      </div>
      <h1 className="title mb-6">New Post</h1>
      <ContentEditor config={postsConfig} />
    </>
  )
}

Page.getLayout = getDashboardLayout('posts')

export default Page
