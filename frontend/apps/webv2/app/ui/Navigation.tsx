import {
  Navbar,
  NavigationActionProps,
  NavigationDropDownProps,
  NavigationLinkProps,
} from 'ui/components/Navbar'
import {
  ArrowRightOnRectangleIcon,
  ArrowTopRightOnSquareIcon,
  Cog8ToothIcon,
  UserIcon,
  WrenchScrewdriverIcon,
} from '@heroicons/react/20/solid'
import { useLogoutHandler, useSession, useUserRole } from '@app/common/session'
import { useCurrentLocation } from '@app/common/hooks'
import { useRoutes } from '@app/common/routes'
import { useIsFetching } from '@tanstack/react-query'
import { useFeatureFlag } from '@app/feature-flags/client'

export default function Navigation() {
  const routes = useRoutes()
  const isFetching = useIsFetching()

  const [session] = useSession()
  const role = useUserRole()
  const isAdmin = role === 'admin'
  const onLogout = useLogoutHandler()
  const currentUrl = useCurrentLocation()
  const webv3Enabled = useFeatureFlag('release-webv3')

  const userNavigation: (
    | NavigationLinkProps
    | NavigationDropDownProps
    | NavigationActionProps
  )[] = session
    ? [
        {
          type: 'action',
          label: 'Log activity',
          href: routes.logCreate(),
        },
        {
          type: 'dropdown',
          label: session.display_name ?? 'User',
          links: [
            {
              label: 'Profile',
              href: routes.userProfileStatistics(session.id),
              IconComponent: UserIcon,
              divider: !isAdmin,
            },
            ...(isAdmin
              ? [
                  {
                    label: 'Admin',
                    href: routes.admin(),
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
            ...(webv3Enabled
              ? [{
                  label: 'Try webv3',
                  href: routes.webv3(),
                  IconComponent: ArrowTopRightOnSquareIcon,
                }]
              : []),
            {
              label: 'Log out',
              href: '#',
              onClick: onLogout,
              IconComponent: ArrowRightOnRectangleIcon,
              mobileBottom: true,
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
