import {
  Navbar,
  NavigationDropDownProps,
  NavigationLinkProps,
} from 'ui/components/Navbar'
import {
  ArrowRightOnRectangleIcon,
  Cog8ToothIcon,
  PlusIcon,
  UserIcon,
  WrenchScrewdriverIcon,
} from '@heroicons/react/20/solid'
import { useLogoutHandler, useSession, useUserRole } from '@app/common/session'
import { useCurrentLocation } from '@app/common/hooks'
import { routes } from '@app/common/routes'
import { useIsFetching } from 'react-query'

export default function Navigation() {
  const isFetching = useIsFetching()

  const [session] = useSession()
  const role = useUserRole()
  const isAdmin = role === 'admin'
  const onLogout = useLogoutHandler([session])
  const currentUrl = useCurrentLocation()

  const userNavigation: (NavigationLinkProps | NavigationDropDownProps)[] =
    session
      ? [
          {
            type: 'dropdown',
            label: session.identity.traits.display_name ?? 'User',
            links: [
              {
                label: 'Profile',
                href: routes.userProfileStatistics(session.identity.id),
                IconComponent: UserIcon,
              },
              {
                label: 'New log',
                href: routes.logCreate(),
                IconComponent: PlusIcon,
                divider: !isAdmin,
              },
              ...(isAdmin
                ? [
                    {
                      label: 'Admin',
                      href: routes.manage(),
                      IconComponent: WrenchScrewdriverIcon,
                      divider: true,
                    },
                  ]
                : []),
              {
                label: 'Account settings',
                href: routes.authSettings(currentUrl),
                IconComponent: Cog8ToothIcon,
              },
              {
                label: 'Log out',
                href: '#',
                onClick: onLogout,
                IconComponent: ArrowRightOnRectangleIcon,
              },
            ],
          },
        ]
      : [
          {
            type: 'link',
            label: 'Log in',
            href: routes.authLogin(currentUrl),
          },
          {
            type: 'link',
            label: 'Sign up',
            href: routes.authSignup(currentUrl),
          },
        ]

  return (
    <Navbar
      logoHref={routes.home()}
      navigation={[
        {
          type: 'link',
          label: 'Home',
          href: routes.home(),
          current: false,
        },
        {
          type: 'link',
          label: 'Leaderboard',
          href: routes.leaderboardLatestOfficial(),
          current: false,
        },
        {
          type: 'link',
          label: 'Contests',
          href: routes.contestListOfficial(),
          current: false,
        },
        {
          type: 'link',
          label: 'Blog',
          href: routes.blogList(),
          current: false,
        },
        ...userNavigation,
      ]}
      isLoading={!!isFetching}
    />
  )
}
