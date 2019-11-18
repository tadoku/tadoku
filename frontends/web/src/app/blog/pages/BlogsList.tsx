import React from 'react'
import BlogApi from '../api'
import { PageTitle } from '../../ui/components'
import { useCachedApiState } from '../../cache'
import { Post } from '../domain'

interface Props {}

const BlogsList = ({  }: Props) => {
  const { data: posts } = useCachedApiState<Post[]>({
    cacheKey: `blog_list?i=1`,
    defaultValue: [],
    fetchData: BlogApi.get,
    dependencies: [],
  })

  return (
    <>
      <PageTitle>Blog</PageTitle>
      <ul>
        {posts.map(p => (
          <li key={p.uuid}>{p.html}</li>
        ))}
      </ul>
    </>
  )
}

export default BlogsList
