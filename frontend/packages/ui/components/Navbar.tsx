import { ComponentType, useEffect } from 'react'
import {
  Disclosure,
  DisclosureButton,
  DisclosurePanel,
  Menu,
  MenuButton,
  MenuItems,
  MenuItem,
} from '@headlessui/react'
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
    divider?: boolean
    mobilePrimary?: boolean
    mobileBottom?: boolean
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
  isLoading?: boolean
}

export function Navbar({
  navigation,
  logoHref,
  width = 'max-w-7xl',
  isLoading = false,
}: Props) {
  const router = useRouter()
  const hasMobileBottomLinks = navigation.some(
    item =>
      item.type === 'dropdown' && item.links.some(link => link.mobileBottom),
  )

  return (
    <>
      <div className="h-10 sm:h-16 absolute top-0 left-0 right-0 bg-white z-0"></div>
      <Disclosure
        as="nav"
        className="sticky top-0 z-40 border-b border-black/10 bg-white shadow shadow-slate-500/10"
      >
        {({ open }) => (
          <>
            <MobileScrollLock active={open} />
            <div className={`mx-auto ${width} px-2 sm:px-6 lg:px-8 z-10`}>
              <div className="relative flex h-10 sm:h-16 items-center justify-between">
                <div className="absolute inset-y-0 left-0 flex items-center sm:hidden">
                  {/* Mobile menu button*/}
                  <DisclosureButton className="inline-flex items-center justify-center p-2 text-secondary hover:bg-secondary/5 focus:bg-secondary/5 focus:outline-none focus:ring-3 focus:ring-inset focus:ring-secondary/20">
                    <span className="sr-only">Open main menu</span>
                    {open ? (
                      <XMarkIcon className="block h-6 w-6" aria-hidden="true" />
                    ) : (
                      <Bars3Icon className="block h-6 w-6" aria-hidden="true" />
                    )}
                  </DisclosureButton>
                </div>
                <div className="flex flex-1 items-center justify-center sm:items-stretch sm:justify-between">
                  <div className="flex flex-shrink-0 items-center">
                    <Link href={logoHref}>
                      <span className="hidden sm:block">
                        <Logo scale={0.8} priority />
                      </span>
                      <span className="block sm:hidden">
                        <Logo scale={0.5} priority />
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
                          const isCurrent =
                            item.current ?? router.pathname === item.href

                          return (
                            <Link
                              key={item.label}
                              href={item.href}
                              className={classNames(
                                isCurrent
                                  ? 'bg-secondary !text-white hover:bg-secondary/80'
                                  : 'text-secondary hover:bg-secondary/5 focus:bg-secondary/5',
                                'reset text-xs px-2 py-1 md:px-3 md:py-2 md:text-sm font-bold inline-flex items-center justify-center',
                              )}
                              aria-current={isCurrent ? 'page' : undefined}
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

            <DisclosurePanel className="sm:hidden">
              <div className="fixed inset-x-0 bottom-0 top-10 z-30 overflow-y-auto bg-white p-2 shadow-lg">
                <div className="flex min-h-full flex-col">
                  <div className="space-y-1">
                    {navigation.map(item =>
                      item.type === 'dropdown'
                        ? item.links.map(
                            (
                              {
                                label,
                                href,
                                IconComponent,
                                onClick,
                                mobilePrimary,
                              },
                              i,
                            ) =>
                              mobilePrimary ? (
                                <MobileDropDownLink
                                  key={`${label}-${i}`}
                                  label={label}
                                  href={href}
                                  onClick={onClick}
                                  IconComponent={IconComponent}
                                  highlighted
                                />
                              ) : null,
                          )
                        : null,
                    )}
                    <div className="px-3 pb-1 text-xs font-semibold uppercase tracking-wide text-slate-500">
                      Navigation
                    </div>
                    {navigation.map(item => {
                      if (item.type === 'link') {
                        return (
                          <DisclosureButton
                            key={item.label}
                            as={Link}
                            href={item.href}
                            className={classNames(
                              item.current
                                ? 'bg-secondary text-white'
                                : 'text-secondary hover:bg-secondary/5 focus:bg-secondary/5',
                              'reset flex min-h-12 w-full items-center px-3 py-2 text-base font-bold transition-[background-color]',
                            )}
                            aria-current={item.current ? 'page' : undefined}
                          >
                            {item.label}
                          </DisclosureButton>
                        )
                      }

                      return null
                    })}
                  </div>

                  {navigation.map(item => {
                    if (
                      item.type === 'dropdown' &&
                      item.links.some(
                        link => !link.mobilePrimary && !link.mobileBottom,
                      )
                    ) {
                      return (
                        <div key={item.label} className="pt-3">
                          <div className="px-3 py-2 text-xs font-semibold uppercase tracking-wide text-slate-500">
                            {item.label}
                          </div>
                          <div>
                            {item.links.map(
                              (
                                {
                                  label,
                                  href,
                                  IconComponent,
                                  onClick,
                                  mobilePrimary,
                                  mobileBottom,
                                },
                                i,
                              ) =>
                                !mobilePrimary && !mobileBottom ? (
                                  <MobileDropDownLink
                                    key={`${label}-${i}`}
                                    label={label}
                                    href={href}
                                    onClick={onClick}
                                    IconComponent={IconComponent}
                                  />
                                ) : null,
                            )}
                          </div>
                        </div>
                      )
                    }

                    return null
                  })}

                  {hasMobileBottomLinks ? (
                    <div className="mt-auto border-t border-slate-500/20">
                      {navigation.map(item =>
                        item.type === 'dropdown'
                          ? item.links.map(
                              (
                                {
                                  label,
                                  href,
                                  IconComponent,
                                  onClick,
                                  mobileBottom,
                                },
                                i,
                              ) =>
                                mobileBottom ? (
                                  <MobileDropDownLink
                                    key={`${label}-${i}`}
                                    label={label}
                                    href={href}
                                    onClick={onClick}
                                    IconComponent={IconComponent}
                                  />
                                ) : null,
                            )
                          : null,
                      )}
                    </div>
                  ) : null}
                </div>
              </div>
            </DisclosurePanel>
            <div
              className={`motion-reduce:hidden ${
                isLoading ? 'opacity-100' : 'opacity-0'
              } bg-gradient-to-br from-primary via-cyan-500 to-emerald-300 h-1 absolute left-0 right-0 top-0 animate-gradient-loading transition-all duration-500 bg-[length:400%_400%]`}
            ></div>
          </>
        )}
      </Disclosure>
    </>
  )
}

const MobileScrollLock = ({ active }: { active: boolean }) => {
  useEffect(() => {
    if (!active) return

    const scrollY = window.scrollY
    const previousOverflow = document.body.style.overflow
    const previousPosition = document.body.style.position
    const previousTop = document.body.style.top
    const previousWidth = document.body.style.width

    document.body.style.overflow = 'hidden'
    document.body.style.position = 'fixed'
    document.body.style.top = `-${scrollY}px`
    document.body.style.width = '100%'

    return () => {
      document.body.style.overflow = previousOverflow
      document.body.style.position = previousPosition
      document.body.style.top = previousTop
      document.body.style.width = previousWidth
      window.scrollTo(0, scrollY)
    }
  }, [active])

  return null
}

interface MobileDropDownLinkProps {
  label: string
  href: string
  IconComponent?: ComponentType<any>
  onClick?: () => void
  highlighted?: boolean
}

const MobileDropDownLink = ({
  label,
  href,
  IconComponent,
  onClick,
  highlighted = false,
}: MobileDropDownLinkProps) => (
  <DisclosureButton
    as={Link}
    href={href}
    onClick={onClick}
    className={classNames(
      'reset flex min-h-12 w-full items-center px-3 py-2 text-base font-bold transition-[background-color]',
      {
        'mb-3 bg-secondary/5 text-secondary hover:bg-secondary/10 focus:bg-secondary/10':
          highlighted,
        'text-secondary hover:bg-secondary/5 focus:bg-secondary/5':
          !highlighted,
      },
    )}
  >
    {IconComponent && (
      <IconComponent className="mr-3 h-5 w-5 shrink-0" aria-hidden="true" />
    )}
    {label}
  </DisclosureButton>
)

const DropDown = ({ label, links }: NavigationDropDownProps) => (
  <div className="">
    <Menu as="div" className="relative">
      <div>
        <MenuButton className="text-secondary hover:bg-secondary/5 focus:bg-secondary/5 text-xs px-2 py-1 md:px-3 md:py-2 md:text-sm font-bold flex items-center justify-center">
          <span className="sr-only">Open navigation menu</span>
          {label}
          <ChevronDownIcon
            className="ml-2 h-4 w-3 md:h-5 md:w-4"
            aria-hidden="true"
          />
        </MenuButton>
      </div>
      <MenuItems
        anchor="bottom end"
        modal={false}
        transition
        className="z-50 mt-2 origin-top-right bg-white py-1 shadow-md shadow-slate-500/20 ring-1 ring-secondary ring-opacity-5 focus:outline-none transition ease-out duration-100 data-[closed]:scale-95 data-[closed]:opacity-0"
      >
        {links.map(({ label, href, IconComponent, onClick, divider }, i) => (
          <MenuItem key={i}>
            <Link
              href={href}
              onClick={onClick}
              className={classNames(
                'reset whitespace-nowrap flex-inline transition-[background-color] items-center px-3 py-2 text-sm text-gray-700 flex font-bold data-[focus]:bg-secondary/5',
                {
                  'border-b border-slate-500/20': !!divider,
                },
              )}
            >
              {IconComponent && <IconComponent className="w-4 h-4 mr-3" />}{' '}
              {label}
            </Link>
          </MenuItem>
        ))}
      </MenuItems>
    </Menu>
  </div>
)
