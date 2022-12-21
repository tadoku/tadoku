import { z } from 'zod'
import getConfig from 'next/config'
import { QueryFunctionContext, useQuery } from 'react-query'

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
      throw new Error('could not fetch postOrPage')
    }

    return Page.parse(await response.json())
  })

const Post = z
  .object({
    id: z.string(),
    slug: z.string(),
    title: z.string(),
    content: z.string(),
    published_at: z.string().datetime({ offset: true }),
  })
  .transform(post => {
    const { published_at: publishedAt, ...rest } = post
    return {
      ...rest,
      publishedAt,
    }
  })

export type Post = z.infer<typeof Post>

export const usePost = (slug: string) =>
  useQuery(['content_post', slug], async ({ queryKey }): Promise<Post> => {
    const [_, slug] = queryKey
    const response = await fetch(`${root}/posts/tadoku/${slug}`)

    if (response.status !== 200) {
      throw new Error('could not fetch postOrPage')
    }

    return Post.parse(await response.json())
  })
