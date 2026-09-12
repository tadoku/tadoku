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

export interface NavigationActionProps {
  type: 'action'
  label: string
  href: string
}

interface Props {
  navigation: (
    | NavigationLinkProps
    | NavigationDropDownProps
    | NavigationActionProps
  )[]
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
      <div className="h-14 md:h-16 absolute top-0 left-0 right-0 bg-white z-0"></div>
      <Disclosure
        as="nav"
        className="sticky top-0 z-40 border-b border-black/10 bg-white shadow shadow-slate-500/10"
      >
        {({ open }) => (
          <>
            <MobileScrollLock active={open} />
            <div className={`mx-auto ${width} px-2 sm:px-6 lg:px-8 z-10`}>
              <div className="relative flex h-14 md:h-16 items-center justify-between">
                <div className="flex min-w-0 flex-1 items-center justify-between">
                  <div className="flex flex-shrink-0 items-center">
                    <Link href={logoHref}>
                      <span className="hidden md:block">
                        <Logo scale={0.8} priority />
                      </span>
                      <span className="block md:hidden">
                        <Logo scale={0.625} priority />
                      </span>
                    </Link>
                  </div>
                  <div className="hidden md:ml-4 md:block">
                    <div className="flex items-center space-x-1 lg:space-x-2">
                      {navigation.map(item => {
                        if (item.type === 'action') {
                          return <NavigationAction {...item} key={item.label} />
                        }

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
                                'reset text-xs px-2 py-1 lg:px-3 lg:py-2 lg:text-sm font-bold inline-flex items-center justify-center',
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
                <div className="ml-2 flex items-center gap-2 md:hidden">
                  {!open &&
                    navigation.map(item =>
                      item.type === 'action' ? (
                        <NavigationAction {...item} key={item.label} />
                      ) : null,
                    )}
                  <DisclosureButton className="btn ghost flex !h-11 !w-11 items-center justify-center !p-0 text-secondary focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary">
                    <span className="sr-only">
                      {open ? 'Close main menu' : 'Open main menu'}
                    </span>
                    {open ? (
                      <XMarkIcon className="!h-5 !w-5" aria-hidden="true" />
                    ) : (
                      <Bars3Icon className="!h-5 !w-5" aria-hidden="true" />
                    )}
                  </DisclosureButton>
                </div>
              </div>
            </div>

            <DisclosurePanel className="md:hidden">
              <div className="fixed inset-x-0 bottom-0 top-14 z-30 overflow-y-auto bg-white p-2 shadow-lg">
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

const NavigationAction = ({ label, href }: NavigationActionProps) => {
  const router = useRouter()
  const isCurrent = router.pathname === href

  return (
    <Link
      href={href}
      aria-label={label}
      aria-current={isCurrent ? 'page' : undefined}
      onClick={event => {
        // Keep an in-progress form intact when its header action is selected again.
        if (
          isCurrent &&
          event.button === 0 &&
          !event.metaKey &&
          !event.ctrlKey &&
          !event.shiftKey &&
          !event.altKey
        ) {
          event.preventDefault()
        }
      }}
      className="btn ghost !h-11 shrink-0 whitespace-nowrap px-2 text-secondary md:text-xs lg:text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
    >
      {label}
    </Link>
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
        <MenuButton className="max-w-24 lg:max-w-48 text-secondary hover:bg-secondary/5 focus:bg-secondary/5 text-xs px-2 py-1 lg:px-3 lg:py-2 lg:text-sm font-bold flex items-center justify-center">
          <span className="sr-only">Open navigation menu</span>
          <span className="truncate">{label}</span>
          <ChevronDownIcon
            className="ml-2 h-4 w-3 shrink-0 lg:h-5 lg:w-4"
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
