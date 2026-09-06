import { VerticalTabbar } from 'paper-ui'
import { useState } from 'react'

const profileLinks = [
  { id: 'entries', label: 'Reading entries', href: '#entries' },
  { id: 'statistics', label: 'Statistics', href: '#statistics' },
  { id: 'contests', label: 'Contests', href: '#profile-contests' },
] as const

export default function VerticalTabbarExample() {
  const [path, setPath] = useState('#statistics')
  return (
    <div className="paper-stack" style={{ maxWidth: '22rem' }}>
      <VerticalTabbar
        label="Profile views"
        currentPath={path}
        links={profileLinks.map(link => ({
          ...link,
          onSelect: () => setPath(link.href),
        }))}
      />
      <p role="status">Selected destination: {path.slice(1)}.</p>
    </div>
  )
}
