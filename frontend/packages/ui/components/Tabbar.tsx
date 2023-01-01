import classNames from 'classnames'
import Link from 'next/link'
import { ComponentType } from 'react'

interface Props {
  links: Link[]
}

interface Link {
  label: string
  href: string
  active: boolean
  IconComponent?: ComponentType<any>
}

export function Tabbar({ links }: Props) {
  return (
    <nav className="relative h-12">
      <div className="flex h-full space-x-10">
        {links.map((link, i) => (
          <TabbarButtonLink key={`${link.href}-${link.label}`} {...link} />
        ))}
      </div>
      <div className="border-b-2 absolute border-slate-200 left-0 right-0 bottom-0 z-0"></div>
    </nav>
  )
}

const TabbarButtonLink = ({ href, IconComponent, label, active }: Link) => (
  <Link
    href={href}
    className={classNames(
      'border-b-4 h-full inline-flex flex-col justify-center items-center z-10 hover:border-primary',
      {
        'border-primary font-semibold': active,
        'border-transparent': !active,
      },
    )}
    data-label={label}
  >
    {IconComponent ? <IconComponent className="w-4 h-4 mr-2" /> : null}
    {label}
  </Link>
)
