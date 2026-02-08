import { Logo, Sidebar, ToastContainer } from 'ui'
import type { AppProps } from 'next/app'
import { useRouter } from 'next/router'
import Link from 'next/link'
import 'ui/styles/globals.css'
import Head from 'next/head'

const getSidebarSections = (currentPath: string) => [
  {
    title: 'Foundation',
    links: [
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
        label: 'Logging flow',
        href: '/logging',
        active: currentPath === '/logging',
      },
    ],
  },
]

export default function App({ Component, pageProps }: AppProps) {
  const router = useRouter()

  return (
    <div className="flex min-h-screen">
      <Head>
        <title>Tadoku Design System</title>
      </Head>
      <div className="w-64 bg-white border-r border-slate-200 flex flex-col flex-shrink-0">
        <div className="p-4 text-center">
          <Link href="/" className="inline-block">
            <Logo scale={0.8} />
          </Link>
          <p className="subtitle mt-1">Design System</p>
        </div>
        <div className="pl-4 pr-0 pb-4 flex-1">
          <Sidebar sections={getSidebarSections(router.pathname)} />
        </div>
      </div>
      <div className="p-8 flex-grow min-w-0">
        <Component {...pageProps} />
        <ToastContainer />
      </div>
    </div>
  )
}
