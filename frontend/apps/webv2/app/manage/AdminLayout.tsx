import { ReactElement } from 'react'
import { routes } from '@app/common/routes'
import { DocumentDuplicateIcon, DocumentTextIcon } from '@heroicons/react/20/solid'
import { Sidebar } from 'ui'

type ActiveLink = 'posts' | 'pages'

interface Props {
  children: React.ReactNode
  activeLink?: ActiveLink
}

export function AdminLayout({ children, activeLink }: Props) {
  return (
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
                  active: activeLink === 'posts',
                  IconComponent: DocumentTextIcon,
                },
                {
                  href: routes.managePages(),
                  label: 'Pages',
                  active: activeLink === 'pages',
                  IconComponent: DocumentDuplicateIcon,
                },
              ],
            },
          ]}
        />
      </div>
      <div className="flex-1">{children}</div>
    </div>
  )
}

export function getAdminLayout(activeLink?: ActiveLink) {
  return function Layout(page: ReactElement) {
    return <AdminLayout activeLink={activeLink}>{page}</AdminLayout>
  }
}
