import { z } from 'zod'
import getConfig from 'next/config'
import { QueryFunctionContext } from 'react-query'

const { publicRuntimeConfig } = getConfig()

const root = `${publicRuntimeConfig.apiEndpoint}/blog`

const PostOrPage = z
  .object({
    id: z.string(),
    slug: z.string(),
    title: z.string(),
    html: z.string(),
    published_at: z.string().datetime({ offset: true }),
  })
  .transform(postOrPage => {
    const { published_at: publishedAt, ...rest } = postOrPage
    return {
      ...rest,
      publishedAt,
    }
  })

export type PostOrPage = z.infer<typeof PostOrPage>

export const getPage = async ({
  queryKey,
}: QueryFunctionContext<[string, { slug: string }]>): Promise<PostOrPage> => {
  const [_, { slug }] = queryKey
  const response = await fetch(`${root}/pages/${slug}`)

  if (response.status !== 200) {
    throw new Error('could not fetch postOrPage')
  }

  return PostOrPage.parse(await response.json())
}
