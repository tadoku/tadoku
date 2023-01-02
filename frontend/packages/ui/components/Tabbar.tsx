import { ChevronDownIcon } from '@heroicons/react/20/solid'
import classNames from 'classnames'
import Link from 'next/link'
import { ComponentType } from 'react'
import { ActionMenu } from './ActionMenu'

interface Props {
  links: Link[]
}

interface Link {
  label: string
  href: string
  active: boolean
  IconComponent?: ComponentType<any>
  disabled?: boolean
}

export function Tabbar({ links }: Props) {
  const activeLink = links.find(l => l.active)
  const fallbackLink = links[0]

  if (!fallbackLink) {
    throw Error('need at least one link for the tabbar to work')
  }

  return (
    <nav className="relative h-12">
      <div className="block md:hidden">
        <ActionMenu links={links}>
          {activeLink ? (
            <>
              {activeLink.IconComponent ? (
                <activeLink.IconComponent className="mr-2" />
              ) : null}
              {activeLink.label}
            </>
          ) : (
            <>
              {fallbackLink.IconComponent ? (
                <fallbackLink.IconComponent className="mr-2" />
              ) : null}
              {fallbackLink.label}
            </>
          )}
          <ChevronDownIcon className="w-4 h-4" />
        </ActionMenu>
      </div>

      <div className="hidden md:flex h-full space-x-10">
        {links.map((link, i) => (
          <TabbarButtonLink key={`${link.href}-${link.label}`} {...link} />
        ))}
      </div>
      <div className="border-b-2 absolute border-slate-200 left-0 right-0 bottom-0 z-0 hidden md:flex"></div>
    </nav>
  )
}

const TabbarButtonLink = ({
  href,
  IconComponent,
  label,
  active,
  disabled,
}: Link) => (
  <Link
    href={href}
    className={classNames(
      'border-b-4 h-full inline-flex flex-col justify-center items-center z-10 hover:border-primary',
      {
        'border-primary font-semibold': active,
        'border-transparent': !active,
        'pointer-events-none opacity-50': disabled,
      },
    )}
    data-label={label}
  >
    {IconComponent ? <IconComponent className="w-4 h-4 mr-2" /> : null}
    {label}
  </Link>
)
