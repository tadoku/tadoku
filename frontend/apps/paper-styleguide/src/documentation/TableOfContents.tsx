import {
  VerticalTabbar,
  type NavigationLinkProps,
  type TabbarItem,
} from 'paper-ui'

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
    <aside className="local-outline">
      <h2 className="paper-type-metadata">On this page</h2>
      <VerticalTabbar
        label="On this page"
        links={links}
        renderLink={renderFragmentLink}
      />
    </aside>
  )
}
