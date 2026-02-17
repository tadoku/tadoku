import { useState } from 'react'
import { Logo, Sidebar, ToastContainer } from 'ui'
import type { AppProps } from 'next/app'
import { useRouter } from 'next/router'
import Link from 'next/link'
import classNames from 'classnames'
import 'ui/styles/globals.css'
import Head from 'next/head'
import {
  ArrowTopRightOnSquareIcon,
  Bars3Icon,
  XMarkIcon,
} from '@heroicons/react/20/solid'

const getSidebarSections = (currentPath: string) => [
  {
    title: 'Foundation',
    links: [
      {
        label: 'Getting Started',
        href: '/',
        active: currentPath === '/',
      },
      { label: 'Color', href: '/color', active: currentPath === '/color' },
      {
        label: 'Typography',
        href: '/typography',
        active: currentPath === '/typography',
      },
      {
        label: 'Branding',
        href: '/branding',
        active: currentPath === '/branding',
      },
      {
        label: 'Templates',
        href: '/templates',
        active: currentPath === '/templates',
      },
    ],
  },
  {
    title: 'Components',
    links: [
      { label: 'Forms', href: '/forms', active: currentPath === '/forms' },
      {
        label: 'Buttons',
        href: '/buttons',
        active: currentPath === '/buttons',
      },
      {
        label: 'Navigation',
        href: '/navigation',
        active: currentPath === '/navigation',
      },
      { label: 'Toasts', href: '/toasts', active: currentPath === '/toasts' },
      {
        label: 'Flash messages',
        href: '/flash',
        active: currentPath === '/flash',
      },
      { label: 'Charts', href: '/charts', active: currentPath === '/charts' },
      { label: 'Modals', href: '/modals', active: currentPath === '/modals' },
      { label: 'Tables', href: '/tables', active: currentPath === '/tables' },
      {
        label: 'Breadcrumb',
        href: '/breadcrumb',
        active: currentPath === '/breadcrumb',
      },
      {
        label: 'Action menu',
        href: '/action-menu',
        active: currentPath === '/action-menu',
      },
      {
        label: 'Pagination',
        href: '/pagination',
        active: currentPath === '/pagination',
      },
    ],
  },
  {
    title: 'Interactions',
    links: [
      {
        label: 'Logs overview',
        href: '/logging',
        active: currentPath === '/logging',
      },
      {
        label: 'Logging flow v2',
        href: '/logging-v2',
        active: currentPath === '/logging-v2',
      },
    ],
  },
]

export default function App({ Component, pageProps }: AppProps) {
  const router = useRouter()
  const [sidebarOpen, setSidebarOpen] = useState(false)

  return (
    <div className="flex min-h-screen">
      <Head>
        <title>Tadoku Design System</title>
      </Head>

      {/* Mobile header */}
      <div
        className={classNames(
          'fixed top-0 left-0 right-0 z-30 flex items-center justify-between bg-white border-b border-slate-200 p-4 md:hidden',
          { hidden: sidebarOpen },
        )}
      >
        <Link href="/">
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
        <div className="p-4 hidden md:block text-center">
          <Link href="/" className="inline-block">
            <Logo scale={0.8} />
          </Link>
          <p className="subtitle mt-1">Design System</p>
        </div>
        {/* Sidebar header - mobile */}
        <div className="p-4 flex items-center justify-between md:hidden">
          <div>
            <Link href="/">
              <Logo scale={0.7} />
            </Link>
            <p className="subtitle">Design System</p>
          </div>
          <button
            onClick={() => setSidebarOpen(false)}
            className="p-2 text-slate-600 hover:text-slate-900"
          >
            <XMarkIcon className="w-6 h-6" />
          </button>
        </div>
        <div className="pl-4 pr-0 pb-4 flex-1">
          <Sidebar sections={getSidebarSections(router.pathname)} />
        </div>
        <div className="p-4">
          <a
            href="https://tadoku.app"
            className="flex items-center gap-2 text-sm text-slate-600 hover:text-slate-900"
          >
            <ArrowTopRightOnSquareIcon className="w-4 h-4" />
            Back to Tadoku
          </a>
        </div>
      </div>

      {/* Main content */}
      <div className="flex-1 pt-20 md:pt-0 min-w-0">
        <div className="p-4 md:p-8">
          <Component {...pageProps} />
          <ToastContainer />
        </div>
      </div>
    </div>
  )
}
