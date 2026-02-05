import { routes } from '@app/common/routes'
import { DocumentTextIcon, HomeIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb } from 'ui'
import { NextPageWithLayout } from '../_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { ContentList } from '@app/content/ContentList'
import { postsConfig } from '@app/content/posts'

const Page: NextPageWithLayout = () => {
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
              href: routes.posts(),
              IconComponent: DocumentTextIcon,
            },
          ]}
        />
      </div>
      <h1 className="title mb-6">Posts</h1>
      <ContentList config={postsConfig} />
    </>
  )
}

Page.getLayout = getDashboardLayout('posts')

export default Page
