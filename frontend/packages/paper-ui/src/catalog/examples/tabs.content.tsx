import { Tabs } from 'paper-ui'
import { useState } from 'react'

export default function TabsExample() {
  const [view, setView] = useState('summary')
  return (
    <div className="paper-stack">
      <section>
        <h3>Automatic selection</h3>
        <Tabs.Root defaultValue="summary">
          <Tabs.List aria-label="Reading log views">
            <Tabs.Tab value="summary">Summary</Tabs.Tab>
            <Tabs.Tab value="entries">Entries</Tabs.Tab>
            <Tabs.Tab value="moderation" disabled>
              Moderation
            </Tabs.Tab>
          </Tabs.List>
          <Tabs.Panel value="summary">1,240 pages read in Japanese.</Tabs.Panel>
          <Tabs.Panel value="entries">12 reading entries.</Tabs.Panel>
          <Tabs.Panel value="moderation">No moderation notes.</Tabs.Panel>
        </Tabs.Root>
        <p>
          Use Left and Right arrows to select. Moderation is unavailable for
          this reader account.
        </p>
      </section>
      <section>
        <h3>Controlled vertical tabs with manual selection</h3>
        <Tabs.Root orientation="vertical" value={view} onValueChange={setView}>
          <Tabs.List aria-label="Reading details" activateOnFocus={false}>
            <Tabs.Tab value="summary">Summary</Tabs.Tab>
            <Tabs.Tab value="entries">Entries</Tabs.Tab>
          </Tabs.List>
          <Tabs.Panel value="summary">1,240 pages read in Japanese.</Tabs.Panel>
          <Tabs.Panel value="entries">12 reading entries.</Tabs.Panel>
        </Tabs.Root>
        <p>
          Use Up and Down arrows to move focus, then Enter or Space to select.
          Selected view: {view}.
        </p>
      </section>
    </div>
  )
}
