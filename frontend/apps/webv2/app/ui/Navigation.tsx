import {
  Navbar,
  NavigationDropDownProps,
  NavigationLinkProps,
} from 'ui/components/Navbar'
import {
  ArrowRightOnRectangleIcon,
  Cog8ToothIcon,
  PlusIcon,
} from '@heroicons/react/20/solid'
import { useLogoutHandler, useSession } from '@app/common/session'
import { useCurrentLocation } from '@app/common/hooks'
import { routes } from '@app/common/routes'

export default function Navigation() {
  const [session] = useSession()
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
                label: 'New log',
                href: routes.logCreate(),
                IconComponent: PlusIcon,
                divider: true,
              },
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
        {
          type: 'link',
          label: 'Forum',
          href: routes.forum(),
          current: false,
        },
        ...userNavigation,
      ]}
    />
  )
}
