export interface OutlineItem {
  id: string
  label: string
}

interface TableOfContentsProps {
  items: readonly OutlineItem[]
}

export function TableOfContents({ items }: TableOfContentsProps) {
  return (
    <nav className="local-outline" aria-label="On this page">
      <h2 className="paper-type-metadata">On this page</h2>
      <ol>
        {items.map((item) => (
          <li key={item.id}>
            <a className="paper-focus-ring" href={`#${item.id}`}>
              {item.label}
            </a>
          </li>
        ))}
      </ol>
    </nav>
  )
}
