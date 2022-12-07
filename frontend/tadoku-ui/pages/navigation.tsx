import { CodeBlock, Preview, Separator, Title } from '@components/example'
import Navbar from '@components/Navbar'
import {
  ArrowRightOnRectangleIcon,
  Cog8ToothIcon,
  UserIcon,
  WrenchScrewdriverIcon,
} from '@heroicons/react/24/solid'
import { toast } from 'react-toastify'

export default function Toasts() {
  return (
    <>
      <h1 className="title mb-8">Toasts</h1>
      <Title>Navigation bar</Title>
      <Preview className="!bg-neutral-100 !p-0">
        <Navbar
          navigation={[
            { label: 'Home', href: '#', current: true },
            { label: 'Blog', href: '#', current: false },
            { label: 'Ranking', href: '#', current: false },
            { label: 'Manual', href: '#', current: false },
            { label: 'Forum', href: '#', current: false },
          ]}
          user={{
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
          }}
        />
      </Preview>
      <CodeBlock code={``} />
    </>
  )
}
