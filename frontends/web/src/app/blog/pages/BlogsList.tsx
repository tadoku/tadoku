import React from 'react'
import BlogApi from '../api'
import { PageTitle, ContentContainer } from '../../ui/components'
import { useCachedApiState } from '../../cache'
import { PostOrPage } from '../interfaces'
import BlogPost from '../components/BlogPost'

const BlogsList = () => {
  const { data: posts } = useCachedApiState<PostOrPage[]>({
    cacheKey: `blog_list?i=1`,
    defaultValue: [],
    fetchData: BlogApi.posts.list,
    dependencies: [],
  })

  return (
    <ContentContainer>
      <PageTitle>Blog</PageTitle>
      {posts.map(p => (
        <BlogPost key={p.slug} post={p} />
      ))}
    </ContentContainer>
  )
}

export default BlogsList
