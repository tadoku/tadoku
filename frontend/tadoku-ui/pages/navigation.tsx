import { CodeBlock, Preview, Title } from '@components/example'
import Navbar from '@components/Navbar'
import {
  ArrowRightOnRectangleIcon,
  Cog8ToothIcon,
  UserIcon,
  WrenchScrewdriverIcon,
} from '@heroicons/react/20/solid'

export default function Toasts() {
  return (
    <>
      <h1 className="title mb-8">Toasts</h1>
      <Title>Navigation bar</Title>
      <Preview className="!bg-neutral-100">
        <Navbar
          navigation={[
            { type: 'link', label: 'Home', href: '#', current: true },
            { type: 'link', label: 'Blog', href: '#', current: false },
            { type: 'link', label: 'Ranking', href: '#', current: false },
            { type: 'link', label: 'Manual', href: '#', current: false },
            { type: 'link', label: 'Forum', href: '#', current: false },
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
      </Preview>
      <CodeBlock
        code={`import Navbar from '@components/Navbar'

const NavigationExample = () => (
  <Navbar
    navigation={[
      { type: 'link', label: 'Home', href: '#', current: true },
      { type: 'link', label: 'Blog', href: '#', current: false },
      { type: 'link', label: 'Ranking', href: '#', current: false },
      { type: 'link', label: 'Manual', href: '#', current: false },
      { type: 'link', label: 'Forum', href: '#', current: false },
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
    </>
  )
}
