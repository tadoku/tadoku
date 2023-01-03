import { ChevronDownIcon } from '@heroicons/react/20/solid'
import Link from 'next/link'
import { ComponentType } from 'react'
import { ActionMenu } from './ActionMenu'

interface Props {
  actions: Action[]
  orientation?: 'left' | 'right'
}

interface Action {
  label: React.ReactNode
  href: string
  onClick?: () => void
  IconComponent?: ComponentType<any>
  disabled?: boolean
  style?: 'primary' | 'secondary' | 'default' | 'danger' | 'ghost'
  visible?: boolean
}

export function ButtonGroup({ actions, orientation }: Props) {
  const filtered = actions.filter(it => it.visible !== false)
  return (
    <>
      <div className="block lg:hidden">
        <ActionMenu
          links={filtered.map(a => ({
            ...a,
            type: a.style === 'danger' ? 'danger' : 'normal',
          }))}
          orientation={orientation}
        >
          Actions <ChevronDownIcon className="w-4 h-4" />
        </ActionMenu>
      </div>

      <div className="hidden lg:flex h-full space-x-3">
        {filtered.map(link => (
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
    className={`reset btn ${style} ${disabled ? 'pointer-events-none' : ''}`}
  >
    {IconComponent ? <IconComponent className="w-4 h-4 mr-2" /> : null}
    {label}
  </Link>
)
