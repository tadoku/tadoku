import { Sidebar } from 'paper-ui'
import { useState } from 'react'

const sidebarSections = [
  {
    id: 'content',
    title: 'Content',
    links: [
      { id: 'logs', label: 'Reading logs', href: '#logs' },
      { id: 'users', label: 'Users', href: '#users' },
    ],
  },
  {
    id: 'system',
    title: 'System',
    links: [
      { id: 'imports', label: 'Imports', href: '#imports' },
      { id: 'jobs', label: 'Background jobs', href: '#jobs', disabled: true },
    ],
  },
] as const

export default function SidebarExample() {
  const [path, setPath] = useState('#logs')
  return (
    <div className="paper-stack" style={{ maxWidth: '22rem' }}>
      <Sidebar
        label="Admin sections"
        currentPath={path}
        sections={sidebarSections.map(section => ({
          ...section,
          links: section.links.map(link => ({
            ...link,
            onSelect: () => setPath(link.href),
          })),
        }))}
      />
      <p role="status">
        Selected: {path.slice(1)}. Background jobs are unavailable while the
        worker is offline.
      </p>
    </div>
  )
}
