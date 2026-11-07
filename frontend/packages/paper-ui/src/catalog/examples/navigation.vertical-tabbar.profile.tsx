import * as stylex from '@stylexjs/stylex'
import { size } from 'paper-ui/tokens.stylex'
import { VerticalTabbar, layout } from 'paper-ui'
import { useState } from 'react'

const profileLinks = [
  { id: 'entries', label: 'Reading entries', href: '#entries' },
  { id: 'statistics', label: 'Statistics', href: '#statistics' },
  { id: 'contests', label: 'Contests', href: '#profile-contests' },
] as const

const styles = stylex.create({
  column: { maxInlineSize: size[352] },
})

export default function VerticalTabbarExample() {
  const [path, setPath] = useState('#statistics')
  return (
    <div {...stylex.props(layout.stack, styles.column)}>
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
