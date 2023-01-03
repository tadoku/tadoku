import { ChevronDownIcon } from '@heroicons/react/20/solid'
import Link from 'next/link'
import { ComponentType } from 'react'
import { ActionMenu } from './ActionMenu'

interface Props {
  actions: Action[]
}

interface Action {
  label: string
  href: string
  onClick?: () => void
  IconComponent?: ComponentType<any>
  disabled?: boolean
  style?: 'primary' | 'secondary' | 'default' | 'danger' | 'ghost'
}

export function ButtonGroup({ actions }: Props) {
  return (
    <>
      <div className="block lg:hidden">
        <ActionMenu
          links={actions.map(a => ({
            ...a,
            type: a.style === 'danger' ? 'danger' : 'normal',
          }))}
        >
          Actions <ChevronDownIcon className="w-4 h-4" />
        </ActionMenu>
      </div>

      <div className="hidden lg:flex h-full space-x-3">
        {actions.map((link, i) => (
          <TabbarButtonLink key={`${link.href}-${link.label}`} {...link} />
        ))}
      </div>
    </>
  )
}

const TabbarButtonLink = ({
  href,
  IconComponent,
  label,
  disabled,
  style = 'default',
}: Action) => (
  <Link
    href={href}
    className={`btn ${style} ${disabled ? 'pointer-events-none' : ''}`}
  >
    {IconComponent ? <IconComponent className="w-4 h-4 mr-2" /> : null}
    {label}
  </Link>
)
