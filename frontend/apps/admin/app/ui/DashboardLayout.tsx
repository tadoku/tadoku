import { ReactElement } from 'react'
import { routes } from '@app/common/routes'
import {
  DocumentDuplicateIcon,
  DocumentTextIcon,
  UsersIcon,
} from '@heroicons/react/20/solid'
import { Sidebar } from 'ui'

type ActiveLink = 'posts' | 'pages' | 'users'

interface Props {
  children: React.ReactNode
  activeLink?: ActiveLink
}

export function DashboardLayout({ children, activeLink }: Props) {
  return (
    <div className="flex min-h-screen">
      <div className="w-64 flex-shrink-0 bg-slate-50 border-r border-slate-200">
        <div className="p-4 border-b border-slate-200">
          <h1 className="text-xl font-bold text-slate-900">Tadoku Admin</h1>
        </div>
        <div className="p-4">
          <Sidebar
            sections={[
              {
                title: 'Content',
                links: [
                  {
                    href: routes.posts(),
                    label: 'Posts',
                    active: activeLink === 'posts',
                    IconComponent: DocumentTextIcon,
                  },
                  {
                    href: routes.pages(),
                    label: 'Pages',
                    active: activeLink === 'pages',
                    IconComponent: DocumentDuplicateIcon,
                  },
                ],
              },
              {
                title: 'Moderation',
                links: [
                  {
                    href: routes.users(),
                    label: 'Users',
                    active: activeLink === 'users',
                    IconComponent: UsersIcon,
                  },
                ],
              },
            ]}
          />
        </div>
      </div>
      <div className="flex-1 bg-white">
        <div className="p-8">{children}</div>
      </div>
    </div>
  )
}

export function getDashboardLayout(activeLink?: ActiveLink) {
  return function Layout(page: ReactElement) {
    return <DashboardLayout activeLink={activeLink}>{page}</DashboardLayout>
  }
}
