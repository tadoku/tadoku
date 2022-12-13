import 'tadoku-ui/styles/globals.css'
import Navbar, {
  NavigationDropDownProps,
  NavigationLinkProps,
} from 'tadoku-ui/components/Navbar'
import {
  ArrowRightOnRectangleIcon,
  Cog8ToothIcon,
} from '@heroicons/react/20/solid'
import { useLogoutHandler, useSession } from '../common/session'
import getConfig from 'next/config'
import { useCurrentLocation } from '../common/hooks'

const { publicRuntimeConfig } = getConfig()

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
                label: 'Account settings',
                href:
                  publicRuntimeConfig.authUiUrl + `/?return_to=${currentUrl}`,
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
            href:
              publicRuntimeConfig.authUiUrl + `/login?return_to=${currentUrl}`,
          },
          {
            type: 'link',
            label: 'Sign up',
            href:
              publicRuntimeConfig.authUiUrl +
              `/register?return_to=${currentUrl}`,
          },
        ]

  return (
    <Navbar
      logoHref="/"
      navigation={[
        {
          type: 'link',
          label: 'Home',
          href: '/',
          current: false,
        },
        ...userNavigation,
      ]}
    />
  )
}
