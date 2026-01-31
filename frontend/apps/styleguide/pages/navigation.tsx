import { CodeBlock, Preview, Separator, Title } from '@components/example'
import { Navbar, Sidebar, Tabbar } from 'ui'
import {
  ArrowRightOnRectangleIcon,
  BookOpenIcon,
  Cog8ToothIcon,
  HomeIcon,
  TrophyIcon,
  UserGroupIcon,
  UserIcon,
  WrenchScrewdriverIcon,
} from '@heroicons/react/20/solid'

export default function Toasts() {
  return (
    <>
      <h1 className="title mb-8">Navigation</h1>
      <Title>Navbar</Title>
      <Preview className="!bg-neutral-100">
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
      </Preview>
      <CodeBlock
        code={`import { Navbar } from 'ui'

const NavigationExample = () => (
  <Navbar
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
          { label: 'Settings', href: '#', IconComponent: Cog8ToothIcon },
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
)`}
      />

      <Separator />

      <Title>Navbar</Title>
      <Preview>
        <Tabbar
          links={[
            {
              href: '/contests/official',
              label: 'Official contests',
              active: false,
            },
            {
              href: '/contests/user-contests',
              label: 'User contests',
              active: false,
            },
            {
              href: '/contests/my-contests',
              label: 'My contests',
              active: true,
            },
          ]}
        />
      </Preview>
      <CodeBlock
        code={`import { Tabbar } from 'ui'

const TabbarExample = () => (
  <Tabbar
    links={[
      {
        href: '/contests/official',
        label: 'Official contests',
        active: false,
      },
      {
        href: '/contests/user-contests',
        label: 'User contests',
        active: false,
      },
      {
        href: '/contests/my-contests',
        label: 'My contests',
        active: true,
      },
    ]}
  />
)`}
      />

      <Separator />

      <Title>Sidebar</Title>
      <Preview>
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
      </Preview>
      <CodeBlock
        code={`import { Sidebar } from 'ui'
import { HomeIcon, BookOpenIcon, TrophyIcon } from '@heroicons/react/20/solid'

const SidebarExample = () => (
  <Sidebar
    sections={[
      {
        title: 'Getting Started',
        links: [
          { href: '/home', label: 'Home', active: true, IconComponent: HomeIcon },
          { href: '/docs', label: 'Documentation', IconComponent: BookOpenIcon },
        ],
      },
      {
        title: 'Contests',
        links: [
          { href: '/contests', label: 'Browse Contests', IconComponent: TrophyIcon },
          { href: '/leaderboard', label: 'Leaderboard' },
          { href: '/disabled', label: 'Disabled Link', disabled: true },
        ],
      },
      {
        title: 'Account',
        links: [
          { href: '/settings', label: 'Settings' },
          { href: '/profile', label: 'Profile' },
        ],
      },
    ]}
  />
)`}
      />
    </>
  )
}
