import { PostOrPage } from './interfaces'
import { postOrPageMapper } from './transform'
import { createApiClient } from '../api'
import { getService } from '@app/services'

const apiClient = createApiClient(getService('blog').externalUrl)

const getPosts = async (): Promise<PostOrPage[]> => {
  const response = await apiClient.get(`/posts`)

  if (response.status !== 200) {
    return []
  }

  const posts = (await response.json()).posts

  return posts.map(postOrPageMapper.fromRaw)
}

const getPage = async (slug: string): Promise<PostOrPage | undefined> => {
  const response = await apiClient.get(`/pages/${slug}`)

  if (response.status !== 200) {
    return undefined
  }

  const page = await response.json()

  return postOrPageMapper.fromRaw(page)
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
