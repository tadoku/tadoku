import '../styles/globals.css'
import type { AppProps } from 'next/app'
import Link from 'next/link'

interface NavigationBlock {
  title: string
  links: {
    title: string
    href: string
  }[]
}

const navigation: NavigationBlock[] = [
  {
    title: 'Foundation',
    links: [
      { title: 'Color', href: '/forms' },
      { title: 'Typography', href: '/typography' },
      { title: 'Templates', href: '/templates' },
    ],
  },
  {
    title: 'Components',
    links: [
      { title: 'Forms', href: '/forms' },
      { title: 'Navigation', href: '/navigation' },
    ],
  },
]

export default function App({ Component, pageProps }: AppProps) {
  return (
    <div className="flex min-h-screen">
      <div className="bg-white w-48 p-8">
        <h1 className="text-2xl font-bold mb-4">tadoku-ui</h1>
        {navigation.map(block => (
          <>
            <h2 className="text-l font-semibold">{block.title}</h2>
            <ul className="mt-2 mb-4">
              {block.links.map(l => (
                <li className="border-l-2 border-neutral-200 hover:border-primary text-neutral-600">
                  <Link href={l.href} className="block pl-4 py-1">
                    {l.title}
                  </Link>
                </li>
              ))}
            </ul>
          </>
        ))}
      </div>
      <div className="p-8 flex-grow">
        <Component {...pageProps} />
      </div>
    </div>
  )
}
