import { type NavigationLinkProps, type TabbarItem, typography, VerticalTabbar } from 'paper-ui'

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
    <aside className="sticky top-24 self-start max-lg:hidden [&_ol]:m-0 [&_ol]:list-none [&_ol]:p-0 [&_h2]:ms-field-x [&_h2]:mt-0">
      <h2 className={typography({ role: 'metadata' })}>On this page</h2>
      <VerticalTabbar
        label="On this page"
        links={links}
        renderLink={renderFragmentLink}
      />
    </aside>
  )
}
