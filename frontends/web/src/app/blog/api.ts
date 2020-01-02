import GhostContentAPI from '@tryghost/content-api'
import { PostOrPage } from './interfaces'
import { RawToPostOrPageMapper } from './transform'

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
    .map(RawToPostOrPageMapper)
}

const BlogApi = {
  posts: {
    list: getPosts,
  },
}

export default BlogApi
