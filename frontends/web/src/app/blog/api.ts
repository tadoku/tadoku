import { PostOrPage } from './interfaces'
import { postOrPageMapper } from './transform'
import { get } from '../api'

const getPosts = async (): Promise<PostOrPage[]> => {
  const response = await get(`/blog/posts`)

  if (response.status !== 200) {
    return []
  }

  const posts = (await response.json()).posts

  return posts.map(postOrPageMapper.fromRaw)
}

const getPage = async (slug: string): Promise<PostOrPage | undefined> => {
  const response = await get(`/blog/pages/${slug}`)

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
