import React from 'react'
import BlogApi from '../api'
import { PageTitle } from '../../ui/components'
import { useCachedApiState } from '../../cache'
import { Post } from '../domain'
import BlogPost from '../components/BlogPost'

interface Props {}

const BlogsList = ({}: Props) => {
  const { data: posts } = useCachedApiState<Post[]>({
    cacheKey: `blog_list?i=1`,
    defaultValue: [],
    fetchData: BlogApi.get,
    dependencies: [],
  })

  return (
    <>
      <PageTitle>Blog</PageTitle>
      {posts.map(p => (
        <BlogPost key={p.uuid} post={p} />
      ))}
    </>
  )
}

export default BlogsList
