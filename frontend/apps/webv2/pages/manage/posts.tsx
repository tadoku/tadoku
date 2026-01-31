import { routes } from '@app/common/routes'
import { DocumentDuplicateIcon, DocumentTextIcon, HomeIcon } from '@heroicons/react/20/solid'
import type { NextPage } from 'next'
import Head from 'next/head'
import { Breadcrumb, Sidebar } from 'ui'

interface Props {}

const Page: NextPage<Props> = () => {
  return (
    <>
      <Head>
        <title>Posts - Admin - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: 'Admin',
              href: routes.manage(),
            },
            {
              label: 'Posts',
              href: routes.managePosts(),
            },
          ]}
        />
      </div>
      <div className="flex gap-8">
        <div className="w-64 flex-shrink-0">
          <Sidebar
            sections={[
              {
                title: 'Content management',
                links: [
                  {
                    href: routes.managePosts(),
                    label: 'Posts',
                    active: true,
                    IconComponent: DocumentTextIcon,
                  },
                  {
                    href: routes.managePages(),
                    label: 'Pages',
                    IconComponent: DocumentDuplicateIcon,
                  },
                ],
              },
            ]}
          />
        </div>
        <div className="flex-1">
          <h1 className="title">Posts</h1>
        </div>
      </div>
    </>
  )
}

export default Page
