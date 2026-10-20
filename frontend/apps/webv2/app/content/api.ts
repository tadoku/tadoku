import {
  contentAnnouncementListActive,
  contentPageFindBySlug,
  contentPostFindBySlug,
  contentPostList,
  type ContentAnnouncement,
  type ContentPage,
  type ContentPost,
  type ContentPosts,
} from 'api-client'
import { apiData, browserApi, requireFields, type WithRequired } from 'ui/api'
import { useQuery } from '@tanstack/react-query'

export type Page = WithRequired<ContentPage, 'html'>

export const usePage = (slug: string) =>
  useQuery(['content_page', slug], async ({ queryKey }): Promise<Page> => {
    const [_, slug] = queryKey
    const page = await apiData(
      contentPageFindBySlug({
        ...browserApi(),
        path: { namespace: 'tadoku', slug },
      }),
    )
    return requireFields(page, 'html')
  })

export type Post = WithRequired<ContentPost, 'published_at'>

export const usePost = (slug: string) =>
  useQuery(['content_post', slug], async ({ queryKey }): Promise<Post> => {
    const [_, slug] = queryKey
    const post = await apiData(
      contentPostFindBySlug({
        ...browserApi(),
        path: { namespace: 'tadoku', slug },
      }),
    )
    return requireFields(post, 'published_at')
  })

export type PostList = Omit<ContentPosts, 'posts'> & { posts: Post[] }

export type ActiveAnnouncement = WithRequired<ContentAnnouncement, 'id'>

export const useActiveAnnouncements = () =>
  useQuery(
    ['content_announcements', 'active'],
    async (): Promise<ActiveAnnouncement[]> => {
      const { data } = await contentAnnouncementListActive({
        ...browserApi(),
        path: { namespace: 'tadoku' },
      })
      return (data?.announcements ?? []).map(announcement =>
        requireFields(announcement, 'id'),
      )
    },
    {
      staleTime: 60_000,
      retry: false,
    },
  )

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
      const data = await apiData(
        contentPostList({
          ...browserApi(),
          path: { namespace: 'tadoku' },
          query: { page_size: pageSize, page: queryKey[2] as number },
        }),
      )
      return {
        ...data,
        posts: data.posts.map(post => requireFields(post, 'published_at')),
      }
    },
  )
