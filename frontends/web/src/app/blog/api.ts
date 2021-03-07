import GhostContentAPI from '@tryghost/content-api'
import { PostOrPage } from './interfaces'
import { postOrPageMapper } from './transform'
import getConfig from 'next/config'

const { publicRuntimeConfig } = getConfig()

const api = GhostContentAPI({
  url: publicRuntimeConfig.GHOST_URL || '',
  key: publicRuntimeConfig.GHOST_KEY || '',
  version: 'canary',
})

const getPosts = async (): Promise<PostOrPage[]> => {
  try {
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
      .map(postOrPageMapper.fromRaw)
  } catch (_) {
    return []
  }
}

const getPage = async (slug: string): Promise<PostOrPage | undefined> => {
  try {
    const response = await api.pages.read(
      { slug },
      {
        formats: ['html'],
      },
    )

    return postOrPageMapper.fromRaw(response)
  } catch (_) {
    return undefined
  }
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
