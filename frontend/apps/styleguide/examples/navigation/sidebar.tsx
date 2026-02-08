import { Sidebar } from 'ui'
import {
  BookOpenIcon,
  Cog8ToothIcon,
  HomeIcon,
  TrophyIcon,
  UserGroupIcon,
  UserIcon,
} from '@heroicons/react/20/solid'

export default function SidebarExample() {
  return (
    <div className="w-64">
      <Sidebar
        sections={[
          {
            title: 'Getting Started',
            links: [
              {
                href: '#',
                label: 'Home',
                active: true,
                IconComponent: HomeIcon,
              },
              {
                href: '#',
                label: 'Documentation',
                IconComponent: BookOpenIcon,
              },
            ],
          },
          {
            title: 'Contests',
            links: [
              {
                href: '#',
                label: 'Browse Contests',
                IconComponent: TrophyIcon,
              },
              {
                href: '#',
                label: 'Leaderboard',
                IconComponent: UserGroupIcon,
              },
              {
                href: '#',
                label: 'Disabled Link',
                disabled: true,
              },
            ],
          },
          {
            title: 'Account',
            links: [
              { href: '#', label: 'Settings', IconComponent: Cog8ToothIcon },
              { href: '#', label: 'Profile', IconComponent: UserIcon },
            ],
          },
        ]}
      />
    </div>
  )
}
