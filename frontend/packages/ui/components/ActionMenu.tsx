import { ComponentType, ReactNode } from 'react'
import { Menu, MenuButton, MenuItems, MenuItem } from '@headlessui/react'
import classNames from 'classnames'

interface Props {
  children: ReactNode
  orientation?: 'left' | 'right'
  links: {
    label: React.ReactNode
    href: string
    IconComponent?: ComponentType<any>
    onClick?: () => void
    type?: 'normal' | 'danger'
  }[]
}

export const ActionMenu = ({
  children,
  links,
  orientation = 'left',
}: Props) => {
  const anchor = orientation === 'left' ? 'bottom start' : 'bottom end'

  return (
    <div className={links.length === 0 ? `hidden` : ``}>
      <Menu as="div" className="relative group">
        <MenuButton className="btn ghost px-0 md:px-2">{children}</MenuButton>
        <MenuItems
          anchor={anchor}
          modal={false}
          transition
          className={`z-10 min-w-[150px] bg-white py-1 shadow-md shadow-slate-500/20 ring-1 ring-secondary ring-opacity-5 focus:outline-none transition ease-out duration-100 data-[closed]:scale-95 data-[closed]:opacity-0`}
        >
          {links.map(
            ({ label, href, IconComponent, onClick, type = 'normal' }, i) => (
              <MenuItem key={i}>
                <a
                  href={href}
                  onClick={onClick}
                  className={classNames(
                    'reset flex-inline items-center px-3 py-2 text-sm flex font-medium',
                    {
                      'data-[focus]:bg-secondary': type === 'normal',
                      'data-[focus]:bg-red-700/80': type === 'danger',
                      'text-red-600 data-[focus]:text-white ': type === 'danger',
                      'text-gray-700 data-[focus]:text-white': type === 'normal',
                    },
                  )}
                >
                  {IconComponent && (
                    <IconComponent className="w-4 h-4 mr-3" />
                  )}{' '}
                  {label}
                </a>
              </MenuItem>
            ),
          )}
        </MenuItems>
      </Menu>
    </div>
  )
}
