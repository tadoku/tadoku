import { routes } from '@app/common/routes'
import { MegaphoneIcon, HomeIcon, PlusIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb } from 'ui'
import { NextPageWithLayout } from '../../_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { AnnouncementEditor } from '@app/announcements/AnnouncementEditor'
import { useNamespace } from '@app/content/NamespaceSelector'

const Page: NextPageWithLayout = () => {
  const namespace = useNamespace()

  return (
    <>
      <Head>
        <title>New Announcement - Admin - Tadoku</title>
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
              label: 'Announcements',
              href: routes.announcements(namespace),
              IconComponent: MegaphoneIcon,
            },
            {
              label: 'New Announcement',
              href: routes.announcementNew(namespace),
              IconComponent: PlusIcon,
            },
          ]}
        />
      </div>
      <h1 className="title mb-6">New Announcement</h1>
      <AnnouncementEditor />
    </>
  )
}

Page.getLayout = getDashboardLayout('announcements')

export default Page
