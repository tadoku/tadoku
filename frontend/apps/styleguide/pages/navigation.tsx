import { Separator } from '@components/example'
import { Showcase } from '@components/Showcase'

import NavbarExample from '@examples/navigation/navbar'
import navbarCode from '@examples/navigation/navbar.tsx?raw'

import TabbarExample from '@examples/navigation/tabbar'
import tabbarCode from '@examples/navigation/tabbar.tsx?raw'

import SidebarExample from '@examples/navigation/sidebar'
import sidebarCode from '@examples/navigation/sidebar.tsx?raw'

export default function Navigation() {
  return (
    <>
      <h1 className="title mb-8">Navigation</h1>

      <Showcase
        title="Navbar"
        code={navbarCode}
        previewClassName="!bg-neutral-100"
      >
        <NavbarExample />
      </Showcase>

      <Separator />

      <Showcase title="Tabbar" code={tabbarCode}>
        <TabbarExample />
      </Showcase>

      <Separator />

      <Showcase title="Sidebar" code={sidebarCode}>
        <SidebarExample />
      </Showcase>
    </>
  )
}
