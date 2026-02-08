import { Navbar } from 'ui'
import {
  ArrowRightOnRectangleIcon,
  Cog8ToothIcon,
  UserIcon,
  WrenchScrewdriverIcon,
} from '@heroicons/react/20/solid'

export default function NavbarExample() {
  return (
    <div className="relative">
      <Navbar
        logoHref="/"
        navigation={[
          { type: 'link', label: 'Home', href: '#', current: true },
          { type: 'link', label: 'Blog', href: '#', current: false },
          { type: 'link', label: 'Ranking', href: '#', current: false },
          { type: 'link', label: 'Manual', href: '#', current: false },
          {
            type: 'dropdown',
            label: 'John Doe',
            links: [
              {
                label: 'Admin',
                href: '#',
                IconComponent: WrenchScrewdriverIcon,
              },
              {
                label: 'Settings',
                href: '#',
                IconComponent: Cog8ToothIcon,
              },
              { label: 'Profile', href: '#', IconComponent: UserIcon },
              {
                label: 'Log out',
                href: '#',
                IconComponent: ArrowRightOnRectangleIcon,
              },
            ],
          },
        ]}
      />
    </div>
  )
}
