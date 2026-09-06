import { Tabbar } from 'paper-ui'
import { useState } from 'react'

const contestLinks = [
  { id: 'official', label: 'Official contests', href: '#official' },
  { id: 'community', label: 'Community contests', href: '#community' },
  { id: 'mine', label: 'My contests', href: '#mine' },
] as const

export default function TabbarExample() {
  const [path, setPath] = useState('#mine')
  return (
    <div className="paper-stack">
      <Tabbar
        label="Contest views"
        currentPath={path}
        links={contestLinks.map(link => ({
          ...link,
          onSelect: () => setPath(link.href),
        }))}
      />
      <p role="status">Selected destination: {path.slice(1)}.</p>
    </div>
  )
}
