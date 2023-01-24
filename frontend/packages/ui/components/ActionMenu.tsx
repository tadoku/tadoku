import { ComponentType, Fragment, ReactNode } from 'react'
import { Menu, Transition } from '@headlessui/react'
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
}: Props) => (
  <div className={links.length === 0 ? `hidden` : ``}>
    <Menu as="div" className="relative group">
      <Menu.Button className="btn ghost px-0 md:px-2">{children}</Menu.Button>
      <Transition
        as={Fragment}
        enter="transition ease-out duration-100"
        enterFrom="transform opacity-0 scale-95"
        enterTo="transform opacity-100 scale-100"
        leave="transition ease-in duration-75"
        leaveFrom="transform opacity-100 scale-100"
        leaveTo="transform opacity-0 scale-95"
      >
        <Menu.Items
          className={`absolute ${orientation}-0 z-10 min-w-[150px] origin-top-right bg-white py-1 shadow-md shadow-slate-500/20 ring-1 ring-secondary ring-opacity-5 focus:outline-none`}
        >
          {links.map(
            ({ label, href, IconComponent, onClick, type = 'normal' }, i) => (
              <Menu.Item key={i}>
                {({ active }) => (
                  <a
                    href={href}
                    onClick={onClick}
                    className={classNames(
                      'reset flex-inline items-center px-3 py-2 text-sm flex font-medium',
                      {
                        'bg-secondary': active && type === 'normal',
                        'bg-red-700/80': active && type === 'danger',
                        'text-red-600 hover:text-white ': type === 'danger',
                        'text-gray-700 hover:text-white': type === 'normal',
                      },
                    )}
                  >
                    {IconComponent && (
                      <IconComponent className="w-4 h-4 mr-3" />
                    )}{' '}
                    {label}
                  </a>
                )}
              </Menu.Item>
            ),
          )}
        </Menu.Items>
      </Transition>
    </Menu>
  </div>
)
