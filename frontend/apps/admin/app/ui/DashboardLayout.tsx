'use client'

import { ReactElement, useState } from 'react'
import { routes } from '@app/common/routes'
import {
  DocumentDuplicateIcon,
  DocumentTextIcon,
  UsersIcon,
  ArrowTopRightOnSquareIcon,
  Bars3Icon,
  XMarkIcon,
} from '@heroicons/react/20/solid'
import { Logo, Sidebar } from 'ui'
import Link from 'next/link'
import classNames from 'classnames'

type ActiveLink = 'posts' | 'pages' | 'users'

interface Props {
  children: React.ReactNode
  activeLink?: ActiveLink
}

const sidebarSections = (activeLink?: ActiveLink) => [
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
]

export function DashboardLayout({ children, activeLink }: Props) {
  const [sidebarOpen, setSidebarOpen] = useState(false)

  return (
    <div className="flex min-h-screen">
      {/* Mobile header */}
      <div className={classNames(
        'fixed top-0 left-0 right-0 z-30 flex items-center justify-between bg-white border-b border-slate-200 p-4 md:hidden',
        { hidden: sidebarOpen },
      )}>
        <Link href={routes.home()}>
          <Logo scale={0.7} />
        </Link>
        <button
          onClick={() => setSidebarOpen(true)}
          className="p-2 text-slate-600 hover:text-slate-900"
        >
          <Bars3Icon className="w-6 h-6" />
        </button>
      </div>

      {/* Mobile sidebar overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 md:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <div
        className={classNames(
          'fixed inset-y-0 left-0 z-50 w-64 bg-white border-r border-slate-200 shadow-sm flex flex-col transform transition-transform duration-200 ease-in-out',
          'md:relative md:translate-x-0 md:flex-shrink-0 md:z-auto',
          sidebarOpen ? 'translate-x-0' : '-translate-x-full',
        )}
      >
        {/* Sidebar header - desktop */}
        <div className="p-4 hidden md:block">
          <Link href={routes.home()}>
            <Logo scale={0.8} />
          </Link>
        </div>
        {/* Sidebar header - mobile */}
        <div className="p-4 flex items-center justify-between md:hidden">
          <Link href={routes.home()}>
            <Logo scale={0.7} />
          </Link>
          <button
            onClick={() => setSidebarOpen(false)}
            className="p-2 text-slate-600 hover:text-slate-900"
          >
            <XMarkIcon className="w-6 h-6" />
          </button>
        </div>
        <div className="pl-4 pr-0 pb-4 mt-4 flex-1">
          <Sidebar sections={sidebarSections(activeLink)} />
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

      {/* Main content */}
      <div className="flex-1 pt-20 md:pt-0 min-w-0">
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
