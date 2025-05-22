import { ToastContainer } from 'ui'
import type { AppProps } from 'next/app'
import Link from 'next/link'
import { Fragment } from 'react'
import 'ui/styles/globals.css'

import Head from 'next/head'

interface NavigationBlock {
  title: string
  links: {
    title: string
    href: string
    todo?: boolean
  }[]
}

const navigation: NavigationBlock[] = [
  {
    title: 'Foundation',
    links: [
      { title: 'Color', href: '/color' },
      { title: 'Typography', href: '/typography' },
      { title: 'Branding', href: '/branding' },
      { title: 'Templates', href: '/templates' },
    ],
  },
  {
    title: 'Components',
    links: [
      { title: 'Forms', href: '/forms' },
      { title: 'Buttons', href: '/buttons' },
      { title: 'Navigation', href: '/navigation' },
      { title: 'Toasts', href: '/toasts' },
      { title: 'Flash messages', href: '/flash' },
      { title: 'Charts', href: '/charts' },
      { title: 'Modals', href: '/modals' },
      { title: 'Tables', href: '/tables' },
      { title: 'Breadcrumb', href: '/breadcrumb' },
      { title: 'Action menu', href: '/action-menu' },
      { title: 'Pagination', href: '/pagination' },
    ],
  },
  {
    title: 'Interactions',
    links: [{ title: 'Logging flow', href: '/logging' }],
  },
]

export default function App({ Component, pageProps }: AppProps) {
  return (
    <div className="flex min-h-screen">
      <Head>
        <title>Tadoku Design System</title>
      </Head>
      <div className="bg-white w-52 p-8 hidden">
        <h1 className="text-2xl font-bold mb-4">tadoku-ui</h1>
        {navigation.map((block, i) => (
          <Fragment key={i}>
            <h2 className="text-l font-semibold">{block.title}</h2>
            <ul className="mt-2 mb-4">
              {block.links.map(l => (
                <li
                  key={l.href}
                  className={`border-l-2 border-neutral-200 text-neutral-600 ${
                    l.todo ? '' : 'hover:border-primary'
                  }`}
                >
                  <Link
                    href={l.todo ? '#' : l.href}
                    className={`block pl-4 py-1 ${
                      l.todo ? 'opacity-40 pointer-events-none' : ''
                    }`}
                  >
                    {l.title}
                  </Link>
                </li>
              ))}
            </ul>
          </Fragment>
        ))}
      </div>
      <div className="p-8 flex-grow">
        <Component {...pageProps} />
        <ToastContainer />
      </div>
    </div>
  )
}
