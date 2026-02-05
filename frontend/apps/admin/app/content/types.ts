import { ReactNode } from 'react'

export interface ContentItem {
  id: string
  namespace: string
  slug: string
  title: string
  body: string
  published_at: string | null
  created_at: string
  updated_at: string
}

export interface ContentListResponse {
  items: ContentItem[]
  total_size: number
  next_page_token: string
}

export interface ContentConfig {
  // Identifiers
  type: 'posts' | 'pages'
  label: string
  labelPlural: string

  // API field that holds the body content ('content' for posts, 'html' for pages)
  bodyField: 'content' | 'html'

  // How to render the body content in previews
  renderBody: (body: string) => ReactNode

  // Format the body content (e.g. via Prettier)
  formatBody: (body: string) => string

  // How to render the body editor
  renderEditor: (props: { value: string; onChange: (v: string) => void; placeholder?: string }) => ReactNode

  // Route helpers (preview/edit take item ID, not slug)
  routes: {
    list: () => string
    preview: (id: string) => string
    edit: (id: string) => string
    new: () => string
  }

  // UI
  icon: React.ComponentType<{ className?: string }>
  sidebarKey: 'posts' | 'pages'
}
