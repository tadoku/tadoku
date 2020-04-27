import React from 'react'
import BlogApi from '../api'
import { PageTitle } from '../../ui/components'
import { useCachedApiState } from '../../cache'
import { PostOrPage } from '../interfaces'
import BlogPost from '../components/BlogPost'
import { postOrPagesSerializer } from '../transform'

const BlogsList = () => {
  const { data: posts } = useCachedApiState<PostOrPage[]>({
    cacheKey: `blog_list?i=2`,
    defaultValue: [],
    fetchData: BlogApi.posts.list,
    dependencies: [],
    serializer: postOrPagesSerializer,
  })

  return (
    <>
      <PageTitle>Blog</PageTitle>
      {posts.map(p => (
        <BlogPost key={p.slug} post={p} />
      ))}
    </>
  )
}

export default BlogsList
