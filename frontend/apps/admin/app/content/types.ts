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

  // Route helpers
  routes: {
    list: () => string
    preview: (slug: string) => string
    edit: (slug: string) => string
    new: () => string
  }

  // UI
  icon: React.ComponentType<{ className?: string }>
  sidebarKey: 'posts' | 'pages'
}
