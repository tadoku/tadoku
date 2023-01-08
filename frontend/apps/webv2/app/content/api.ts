import { z } from 'zod'
import getConfig from 'next/config'
import { useQuery } from 'react-query'

const { publicRuntimeConfig } = getConfig()

const root = `${publicRuntimeConfig.apiEndpoint}/content`

const Page = z.object({
  id: z.string(),
  slug: z.string(),
  title: z.string(),
  html: z.string(),
})

export type Page = z.infer<typeof Page>

export const usePage = (slug: string) =>
  useQuery(['content_page', slug], async ({ queryKey }): Promise<Page> => {
    const [_, slug] = queryKey
    const response = await fetch(`${root}/pages/tadoku/${slug}`)

    if (response.status !== 200) {
      throw new Error('could not fetch page')
    }

    return Page.parse(await response.json())
  })

const Post = z.object({
  id: z.string(),
  slug: z.string(),
  title: z.string(),
  content: z.string(),
  published_at: z.string().datetime({ offset: true }),
})

export type Post = z.infer<typeof Post>

export const usePost = (slug: string) =>
  useQuery(['content_post', slug], async ({ queryKey }): Promise<Post> => {
    const [_, slug] = queryKey
    const response = await fetch(`${root}/posts/tadoku/${slug}`)

    if (response.status !== 200) {
      throw new Error('could not fetch post')
    }

    return Post.parse(await response.json())
  })

const PostList = z.object({
  posts: z.array(Post),
  next_page_token: z.string(),
  total_size: z.number(),
})

export type PostList = z.infer<typeof PostList>

export const usePostList = ({
  pageSize,
  page,
}: {
  pageSize: number
  page: number
}) =>
  useQuery(
    ['content_post', 'list', page],
    async ({ queryKey }): Promise<PostList> => {
      const page = queryKey[2]
      const response = await fetch(
        `${root}/posts/tadoku?page_size=${pageSize}&page=${page}`,
      )

      if (response.status !== 200) {
        throw new Error('could not fetch post list')
      }

      return PostList.parse(await response.json())
    },
  )
