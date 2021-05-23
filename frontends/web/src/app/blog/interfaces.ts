export interface PostOrPage {
  id: string
  slug: string
  title: string
  html: string
  publishedAt: Date
}

export interface RawPostOrPage {
  id: string
  slug: string
  title: string
  html: string
  published_at: Date
}
