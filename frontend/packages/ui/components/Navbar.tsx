import { ComponentType, Fragment } from 'react'
import { Disclosure, Menu, Transition } from '@headlessui/react'
import { Bars3Icon, XMarkIcon } from '@heroicons/react/24/outline'
import { Logo } from './branding'
import { ChevronDownIcon } from '@heroicons/react/24/solid'
import classNames from 'classnames'
import { useRouter } from 'next/router'
import Link from 'next/link'

export interface NavigationDropDownProps {
  type: 'dropdown'
  label: string
  links: {
    label: string
    href: string
    IconComponent?: ComponentType<any>
    onClick?: () => void
  }[]
}

export interface NavigationLinkProps {
  type: 'link'
  label: string
  href: string
  current?: boolean
}

interface Props {
  navigation: (NavigationLinkProps | NavigationDropDownProps)[]
  width?: string
  logoHref: string
}

export function Navbar({ navigation, logoHref, width = 'max-w-7xl' }: Props) {
  const router = useRouter()

  return (
    <>
      <div className="h-10 sm:h-16 absolute top-0 left-0 right-0 bg-white z-0"></div>
      <Disclosure
        as="nav"
        className="bg-white sticky top-0 z-50 backdrop-blur bg-white/70 border-b border-black/10 shadow shadow-slate-500/10"
      >
        {({ open }) => (
          <>
            <div className={`mx-auto ${width} px-2 sm:px-6 lg:px-8 z-10`}>
              <div className="relative flex h-10 sm:h-16 items-center justify-between">
                <div className="absolute inset-y-0 left-0 flex items-center sm:hidden">
                  {/* Mobile menu button*/}
                  <Disclosure.Button className="inline-flex items-center justify-center p-2 text-secondary hover:bg-secondary hover:text-white focus:outline-none focus:ring-3 focus:ring-inset focus:ring-white">
                    <span className="sr-only">Open main menu</span>
                    {open ? (
                      <XMarkIcon className="block h-6 w-6" aria-hidden="true" />
                    ) : (
                      <Bars3Icon className="block h-6 w-6" aria-hidden="true" />
                    )}
                  </Disclosure.Button>
                </div>
                <div className="flex flex-1 items-center justify-center sm:items-stretch sm:justify-between">
                  <div className="flex flex-shrink-0 items-center">
                    <Link href={logoHref}>
                      <span className="hidden sm:block">
                        <Logo scale={0.8} />
                      </span>
                      <span className="block sm:hidden">
                        <Logo scale={0.5} />
                      </span>
                    </Link>
                  </div>
                  <div className="hidden sm:ml-6 sm:block">
                    <div className="flex space-x-1 md:space-x-2">
                      {navigation.map(item => {
                        if (item.type === 'dropdown') {
                          return <DropDown {...item} key={item.label} />
                        }

                        if (item.type === 'link') {
                          return (
                            <Link
                              key={item.label}
                              href={item.href}
                              className={classNames(
                                item.current || router.pathname === item.href
                                  ? 'bg-secondary !text-white hover:bg-secondary/80'
                                  : 'text-secondary hover:bg-secondary hover:text-white',
                                'reset text-xs px-2 py-1 md:px-3 md:py-2 md:text-sm font-bold inline-flex items-center justify-center',
                              )}
                              aria-current={item.current ? 'page' : undefined}
                            >
                              {item.label}
                            </Link>
                          )
                        }
                      })}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <Disclosure.Panel className="sm:hidden">
              <div className="space-y-1 px-2 pt-2 pb-3">
                {navigation.map(item => {
                  if (item.type === 'dropdown') {
                    return (
                      <div
                        key={item.label}
                        className={classNames(
                          'text-secondary hover:bg-secondary hover:text-white',
                          'reset block px-3 py-2 text-base font-bold',
                        )}
                      >
                        <DropDownMobile {...item} />
                      </div>
                    )
                  }

                  if (item.type === 'link') {
                    return (
                      <Disclosure.Button
                        key={item.label}
                        as={Link}
                        href={item.href}
                        className={classNames(
                          item.current
                            ? 'bg-secondary text-white'
                            : 'text-secondary hover:bg-secondary hover:text-white',
                          'reset transition-[background-color] block px-3 py-2 text-base font-bold',
                        )}
                        aria-current={item.current ? 'page' : undefined}
                      >
                        {item.label}
                      </Disclosure.Button>
                    )
                  }
                })}
              </div>
            </Disclosure.Panel>
          </>
        )}
      </Disclosure>
    </>
  )
}

const DropDown = ({ label, links }: NavigationDropDownProps) => (
  <div className="">
    <Menu as="div" className="relative">
      <div>
        <Menu.Button className="text-secondary hover:bg-secondary hover:text-white text-xs px-2 py-1 md:px-3 md:py-2 md:text-sm font-bold flex items-center justify-center">
          <span className="sr-only">Open navigation menu</span>
          {label}
          <ChevronDownIcon
            className="ml-2 h-4 w-3 md:h-5 md:w-4"
            aria-hidden="true"
          />
        </Menu.Button>
      </div>
      <Transition
        as={Fragment}
        enter="transition ease-out duration-100"
        enterFrom="transform opacity-0 scale-95"
        enterTo="transform opacity-100 scale-100"
        leave="transition ease-in duration-75"
        leaveFrom="transform opacity-100 scale-100"
        leaveTo="transform opacity-0 scale-95"
      >
        <Menu.Items className="absolute right-0 z-10 mt-2 origin-top-right bg-white py-1 shadow-md shadow-slate-500/20 ring-1 ring-secondary ring-opacity-5 focus:outline-none">
          {links.map(({ label, href, IconComponent, onClick }, i) => (
            <Menu.Item key={i}>
              {({ active }) => (
                <Link
                  href={href}
                  onClick={onClick}
                  className={classNames(
                    active ? 'bg-secondary !text-white' : '',
                    'reset whitespace-nowrap flex-inline transition-[background-color] items-center px-3 py-2 text-sm text-gray-700 flex font-bold',
                  )}
                >
                  {IconComponent && <IconComponent className="w-4 h-4 mr-3" />}{' '}
                  {label}
                </Link>
              )}
            </Menu.Item>
          ))}
        </Menu.Items>
      </Transition>
    </Menu>
  </div>
)

const DropDownMobile = ({ label, links }: NavigationDropDownProps) => (
  <div className="">
    {/* Profile dropdown */}
    <Menu as="div" className="relative">
      <div>
        <Menu.Button className="flex items-center justify-between w-full">
          <span className="sr-only">Open navigation menu</span>
          {label}
          <ChevronDownIcon
            className="ml-2 h-4 w-3 md:h-5 md:w-4"
            aria-hidden="true"
          />
        </Menu.Button>
      </div>
      <Transition
        as={Fragment}
        enter="transition ease-out duration-100"
        enterFrom="transform opacity-0 scale-95"
        enterTo="transform opacity-100 scale-100"
        leave="transition ease-in duration-75"
        leaveFrom="transform opacity-100 scale-100"
        leaveTo="transform opacity-0 scale-95"
      >
        <Menu.Items className="absolute left-0 right-0 z-10 mt-2 min-w-48 origin-top-right bg-white py-1 shadow-md shadow-slate-500/20 ring-1 ring-secondary ring-opacity-5 focus:outline-none">
          {links.map(({ label, href, IconComponent, onClick }, i) => (
            <Menu.Item key={i}>
              {({ active }) => (
                <Disclosure.Button
                  as={Link}
                  href={href}
                  onClick={onClick}
                  className={classNames(
                    active ? 'bg-secondary !text-white' : '',
                    'reset transition-[background-color] flex-inline items-center px-3 py-4 text-sm text-gray-700 flex font-bold',
                  )}
                >
                  {IconComponent && <IconComponent className="w-4 h-4 mr-3" />}{' '}
                  {label}
                </Disclosure.Button>
              )}
            </Menu.Item>
          ))}
        </Menu.Items>
      </Transition>
    </Menu>
  </div>
)
