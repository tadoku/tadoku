import 'tadoku-ui/styles/globals.css'
import Navbar, {
  NavigationDropDownProps,
  NavigationLinkProps,
} from 'tadoku-ui/components/Navbar'
import {
  ArrowRightOnRectangleIcon,
  Cog8ToothIcon,
} from '@heroicons/react/20/solid'
import { useLogoutHandler, useSession } from './session'

export default function Navigation() {
  const [session] = useSession()
  const onLogout = useLogoutHandler([session])

  const userNavigation: (NavigationLinkProps | NavigationDropDownProps)[] =
    session
      ? [
          {
            type: 'dropdown',
            label: session.identity.traits.display_name ?? 'User',
            links: [
              { label: 'Settings', href: '/', IconComponent: Cog8ToothIcon },
              {
                label: 'Log out',
                href: '#',
                onClick: onLogout,
                IconComponent: ArrowRightOnRectangleIcon,
              },
            ],
          },
        ]
      : []

  return (
    <Navbar
      width="max-w-xl"
      navigation={[
        {
          type: 'link',
          label: 'Home',
          href: 'https://tadoku.app/',
          current: false,
        },
        ...userNavigation,
      ]}
    />
  )
}
