import classNames from 'classnames'
import Link from 'next/link'
import { ComponentType } from 'react'

interface SidebarLink {
  label: string
  href: string
  active?: boolean
  IconComponent?: ComponentType<any>
  disabled?: boolean
}

interface SidebarSection {
  title: string
  links: SidebarLink[]
}

interface Props {
  sections: SidebarSection[]
}

export function Sidebar({ sections }: Props) {
  return (
    <nav className="flex flex-col space-y-6">
      {sections.map((section, sectionIndex) => (
        <div key={sectionIndex}>
          <h3 className="font-semibold mb-2">{section.title}</h3>
          <div className="flex flex-col">
            {section.links.map((link) => (
              <Link
                href={link.href}
                className={classNames(
                  'reset border-l-4 pl-4 pr-2 py-2 text-sm inline-flex items-center text-slate-600 hover:border-primary/50 hover:bg-neutral-50 transition-colors',
                  {
                    'border-primary font-semibold bg-neutral-100': link.active,
                    'border-slate-200': !link.active,
                    'pointer-events-none opacity-50': link.disabled,
                  },
                )}
                key={`${link.href}-${link.label}`}
              >
                {link.IconComponent && (
                  <link.IconComponent className="w-4 h-4 mr-2 flex-shrink-0" />
                )}
                {link.label}
              </Link>
            ))}
          </div>
        </div>
      ))}
    </nav>
  )
}
