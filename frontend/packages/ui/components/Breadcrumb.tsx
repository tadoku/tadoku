import { ArrowUturnLeftIcon, ChevronRightIcon } from '@heroicons/react/20/solid'
import classNames from 'classnames'
import Link from 'next/link'
import { ComponentType } from 'react'

interface Props {
  links: Link[]
}

interface Link {
  label: string
  href: string
  IconComponent?: ComponentType<any>
}

export function Breadcrumb({ links }: Props) {
  return (
    <nav className="flex" aria-label="Breadcrumb">
      {links.length > 1 ? (
        <div className="md:hidden">
          <BreadcrumbLink
            {...links[links.length - 2]}
            IconComponent={ArrowUturnLeftIcon}
            isLast={false}
          />
        </div>
      ) : null}
      <ol className="hidden items-center space-x-1 md:space-x-3 md:inline-flex">
        {links.map((link, i) => (
          <li className="inline-flex items-center" key={i}>
            <BreadcrumbLink {...link} isLast={i === links.length - 1} />
            {i !== links.length - 1 ? (
              <ChevronRightIcon className="w-4 h-4 ml-3" />
            ) : null}
          </li>
        ))}
      </ol>
    </nav>
  )
}

interface BreadcrumbLinkProps extends Link {
  isLast: boolean
}

const BreadcrumbLink = ({
  href,
  IconComponent,
  label,
  isLast,
}: BreadcrumbLinkProps) => (
  <Link
    href={href}
    className={classNames(
      'reset inline-flex items-center text-sm font-medium',
      {
        'text-gray-800 hover:text-primary': !isLast,
        'text-gray-500 pointer-events-none': isLast,
      },
    )}
  >
    {IconComponent ? <IconComponent className="w-4 h-4 mr-2" /> : null}
    {label}
  </Link>
)
