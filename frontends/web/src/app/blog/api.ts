import GhostContentAPI from '@tryghost/content-api'
import { PostOrPage } from './interfaces'
import { rawToPostOrPageMapper } from './transform'

const api = GhostContentAPI({
  url: process.env.GHOST_URL || '',
  key: process.env.GHOST_KEY || '',
  version: 'canary',
})

const getPosts = async (): Promise<PostOrPage[]> => {
  const response = await api.posts.browse({
    limit: 5,
    include: ['authors', 'tags'],
    formats: ['html'],
  })

  if (!response) {
    return []
  }

  return Object.entries(response)
    .filter(([key]) => key !== 'meta')
    .map(([, p]) => p)
    .map(rawToPostOrPageMapper)
}

const getPage = async (slug: string): Promise<PostOrPage> => {
  const response = await api.pages.read(
    { slug },
    {
      formats: ['html'],
    },
  )

  return rawToPostOrPageMapper(response)
}

const BlogApi = {
  posts: {
    list: getPosts,
  },
  pages: {
    get: getPage,
  },
}

export default BlogApi
