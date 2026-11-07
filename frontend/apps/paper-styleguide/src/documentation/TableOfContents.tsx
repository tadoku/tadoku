import * as stylex from '@stylexjs/stylex'
import { typography, VerticalTabbar, type NavigationLinkProps, type TabbarItem } from 'paper-ui'
import { shellStyles } from '../styles/shell.styles'
import { descendantStyles } from '../descendants.styles'

export interface OutlineItem {
  id: string
  label: string
}

interface TableOfContentsProps {
  items: readonly OutlineItem[]
}

function renderFragmentLink(props: NavigationLinkProps) {
  return <a {...props} />
}

export function TableOfContents({ items }: TableOfContentsProps) {
  const links: readonly TabbarItem[] = items.map((item) => ({
    id: item.id,
    label: item.label,
    href: `#${item.id}`,
  }))

  return (
    <aside {...stylex.props(shellStyles.localOutline)}>
      <h2 {...stylex.props(typography.metadata, descendantStyles.localOutlineU_uH2)}>On this page</h2>
      <VerticalTabbar
        label="On this page"
        links={links}
        renderLink={renderFragmentLink}
      />
    </aside>
  )
}
