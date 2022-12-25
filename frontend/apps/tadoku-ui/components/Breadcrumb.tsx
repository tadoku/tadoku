import { ChevronRightIcon } from '@heroicons/react/20/solid'
import classNames from 'classnames'
import { ComponentType } from 'react'

interface Props {
  links: {
    label: string
    href: string
    IconComponent?: ComponentType<any>
  }[]
}

export default function Breadcrumb({ links }: Props) {
  return (
    <nav className="flex" aria-label="Breadcrumb">
      <ol className="inline-flex items-center space-x-1 md:space-x-3">
        {links.map(({ label, href, IconComponent }, i) => (
          <li className="inline-flex items-center" key={i}>
            <a
              href={href}
              className={classNames(
                'reset inline-flex items-center text-sm font-medium',
                {
                  'text-gray-800 hover:text-primary': i !== links.length - 1,
                  'text-gray-500 pointer-events-none': i === links.length - 1,
                },
              )}
            >
              {IconComponent ? (
                <IconComponent className="w-4 h-4 mr-2" />
              ) : null}
              {label}
            </a>
            {i !== links.length - 1 ? (
              <ChevronRightIcon className="w-4 h-4 ml-3" />
            ) : null}
          </li>
        ))}
      </ol>
    </nav>
  )
}
