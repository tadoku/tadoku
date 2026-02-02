import { ReactElement } from 'react'
import { routes } from '@app/common/routes'
import {
  DocumentDuplicateIcon,
  DocumentTextIcon,
  UsersIcon,
  ArrowTopRightOnSquareIcon,
} from '@heroicons/react/20/solid'
import { Logo, Sidebar } from 'ui'
import Link from 'next/link'

type ActiveLink = 'posts' | 'pages' | 'users'

interface Props {
  children: React.ReactNode
  activeLink?: ActiveLink
}

export function DashboardLayout({ children, activeLink }: Props) {
  return (
    <div className="flex min-h-screen">
      <div className="w-64 flex-shrink-0 bg-white border-r border-slate-200 shadow-sm flex flex-col">
        <div className="p-4">
          <Link href={routes.home()}>
            <Logo scale={0.8} />
          </Link>
        </div>
        <div className="pl-4 pr-0 pb-4 mt-4 flex-1">
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
        <div className="p-4">
          <a
            href={routes.mainApp()}
            className="flex items-center gap-2 text-sm text-slate-600 hover:text-slate-900"
          >
            <ArrowTopRightOnSquareIcon className="w-4 h-4" />
            Back to Tadoku
          </a>
        </div>
      </div>
      <div className="flex-1">
        <div className="p-4 md:p-8">{children}</div>
      </div>
    </div>
  )
}

export function getDashboardLayout(activeLink?: ActiveLink) {
  return function Layout(page: ReactElement) {
    return <DashboardLayout activeLink={activeLink}>{page}</DashboardLayout>
  }
}
